package marcopoller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"cloud.google.com/go/datastore"
	"github.com/alexandre-normand/slackscot/store"
	"github.com/alexandre-normand/slackscot/store/datastoredb"
	"github.com/imroc/req"
	"github.com/lithammer/shortuuid"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/spf13/cast"
	opentelemetry "go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/metric"
	"google.golang.org/api/option"
)

// Environment variables
const (
	GCPProjectIDEnv = "PROJECT_ID"
	DebugEnabledEnv = "DEBUG"
)

// Storage keys
const (
	pollInfoKey = "pollInfo"
)

// App constants
const (
	appName               = "marco-poller"
	friendlyName          = "Marco Poller"
	persistenceKindName   = "marcoPoller"
	voteDelimiter         = ","
	buttonIDPartDelimiter = ","
)

// Fixed button identifiers
const (
	voteButtonValue   = "vote"
	deleteButtonValue = "delete"
	closeButtonValue  = "close"
)

// Interactive Prompt Identifiers
const (
	interactivePollCallbackID = "interactive-poll-create"

	pollConversationInputBlockID   = "poll_conversation_select"
	pollConversationSelectActionID = "poll_conversation_select"

	pollQuestionInputBlockID = "poll_question"
	pollQuestionActionID     = "poll_question"

	pollFeaturesInputBlockID = "poll_features"
	pollFeaturesActionID     = "poll_features"
	multiAnswerOptionID      = "multivoting"

	pollOptionsInputBlockID = "poll_answer_options"
	pollOptionsActionID     = "poll_answer_options"

	multiAnswerFeatureValue = "Allow voters to vote for many options"
)

// Slack slash command parameter names
const (
	textParam        = "text"
	channelParam     = "channel_id"
	creatorParam     = "user_id"
	responseURLParam = "response_url"
	triggerIDParam   = "trigger_id"
)

// Poll represents a poll
type Poll struct {
	ID       string       `json:"id"`
	Question string       `json:"question"`
	Options  []string     `json:"options"`
	Features PollFeatures `json:"features,omitempty"`
	Creator  string       `json:"creator"`
}

// PollFeatures represents features on a poll
type PollFeatures struct {
	MultiAnswers bool `json:"multianswers"`
}

// ActionResponse represents a response to a slash command or action
type ActionResponse struct {
	ResponseType    string        `json:"response_type,omitempty"`
	Text            string        `json:"text,omitempty"`
	Blocks          []slack.Block `json:"blocks,omitempty"`
	ReplaceOriginal bool          `json:"replace_original"`
}

// Voter represents a voting user
type Voter struct {
	userID    string
	avatarURL string
	name      string
}

// instruments
type instruments struct {
	pollCount   metric.BoundInt64Counter
	votingCount metric.BoundInt64Counter

	// TODO: Add this one once there's a mechanism for expiring/closing polls
	// since that would be the place to instrument this
	//votesPerPoll metric.Int64Measure
}

// MarcoPoller represents a Marco Poller instance
type MarcoPoller struct {
	storer       store.GlobalSiloStringStorer
	userFinder   UserFinder
	verifier     Verifier
	pollVerifier PollVerifier
	dialoguer    Dialoguer
	debug        bool
	meter        metric.Meter
	instruments  *instruments
}

// DeleteMessage represents the slack action response to delete an original message
type DeleteMessage struct {
	DeleteOriginal bool `json:"delete_original,omitempty"`
}

// UpdateMessage represents the slack action response to update an original message
type UpdateMessage struct {
	ActionResponse
}

// UserFinder is implemented by any value that has the GetInfo method
type UserFinder interface {
	// GetUserInfo will retrieve the complete user information. See https://godoc.org/github.com/slack-go/slack#Client.GetUserInfo
	GetUserInfo(user string) (*slack.User, error)
}

// Verifier is implemented by any value that has the Verify method
type Verifier interface {
	Verify(header http.Header, body []byte) (err error)
}

// Dialoguer is implemented by any value that has the OpenView method
type Dialoguer interface {
	// OpenView will open a block kit modal view. See https://pkg.go.dev/github.com/slack-go/slack?tab=doc#Client.OpenView
	OpenView(triggerID string, view slack.ModalViewRequest) (resp *slack.ViewResponse, err error)
}

// SlackVerifier represents a slack verifier backed by github.com/slack-go/slack
type SlackVerifier struct {
	slackSigningSecret string
}

// Verify verifies the slack request's authenticity (https://api.slack.com/docs/verifying-requests-from-slack). If the request
// can't be verified or if it fails verification, an error is returned. For a verified valid request, nil is returned
func (v *SlackVerifier) Verify(header http.Header, body []byte) (err error) {
	verifier, err := slack.NewSecretsVerifier(header, v.slackSigningSecret)
	if err != nil {
		return errors.Wrap(err, "Error creating slack secrets verifier")
	}

	_, err = verifier.Write(body)
	if err != nil {
		return errors.Wrap(err, "Error writing request body to slack verifier")
	}

	err = verifier.Ensure()
	if err != nil {
		return err
	}

	return nil
}

// PollVerifier is implemented by any value that has the Verify method. The PollVerifier
// returns an error when verifying a poll that is expired (read-only)
type PollVerifier interface {
	Verify(pollID string, eventTime time.Time) (err error)
}

// ExpirationPollVerifier holds a poll verifier's validity period before it's marked as expired
type ExpirationPollVerifier struct {
	ValidityPeriod time.Duration
}

// Verify extracts the creation time from the pollID and returns an error if the event (current) time is past the
// creation time + the poll's validity period
func (epv ExpirationPollVerifier) Verify(pollID string, eventTime time.Time) (err error) {
	creationTime := getPollCreationTime(pollID)

	pollAge := eventTime.Sub(creationTime)
	if pollAge.Seconds() > epv.ValidityPeriod.Seconds() {
		return fmt.Errorf("the poll is expired and is now read-only")
	}

	return nil
}

// AlwaysValidPollVerifier acts as a default PollVerifier that always verifies a poll without error
type AlwaysValidPollVerifier struct {
}

// Verify to a valid poll in all cases
func (avpv AlwaysValidPollVerifier) Verify(pollID string, actionTime time.Time) (err error) {
	return nil
}

// Option is a function that applies an option to a MarcoPoller instance
type Option func(mp *MarcoPoller) (err error)

// OptionSlackUserFinder sets a slack-go/slack.Client as the implementation of UserFinder
func OptionSlackUserFinder(token string, debug bool) Option {
	return func(mp *MarcoPoller) (err error) {
		sc := slack.New(token, slack.OptionDebug(debug))
		mp.userFinder = sc
		return nil
	}
}

// OptionSlackDialoguer sets a slack-go/slack.Client as the implementation of Dialoguer
func OptionSlackDialoguer(token string, debug bool) Option {
	return func(mp *MarcoPoller) (err error) {
		sc := slack.New(token, slack.OptionDebug(debug))
		mp.dialoguer = sc
		return nil
	}
}

// OptionSlackVerifier sets a slack-go-backed SlackVerifier as the implementation of Verifier
func OptionSlackVerifier(slackSigningSecret string) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.verifier = &SlackVerifier{slackSigningSecret: slackSigningSecret}

		return nil
	}
}

// OptionDatastore sets a datastoredb as the implementation of GlobalSiloStringStorer
func OptionDatastore(datastoreProjectID string, gcloudClientOpts ...option.ClientOption) Option {
	return func(mp *MarcoPoller) (err error) {
		meter := opentelemetry.MeterProvider().Meter("github.com/alexandre-normand/marcopoller")

		mp.storer, err = datastoredb.NewWithTelemetry(appName, meter, persistenceKindName, datastoreProjectID, gcloudClientOpts...)
		if err != nil {
			return errors.Wrapf(err, "Error initializing datastore persistence on project [%s]", datastoreProjectID)
		}

		return nil
	}
}

// OptionUserFinder sets a userFinder as the implementation on MarcoPoller
func OptionUserFinder(userFinder UserFinder) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.userFinder = userFinder
		return nil
	}
}

// OptionDialoguer sets a dialoguer as the implementation on MarcoPoller
func OptionDialoguer(dialoguer Dialoguer) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.dialoguer = dialoguer
		return nil
	}
}

// OptionStorer sets a storer as the implementation on MarcoPoller
func OptionStorer(storer store.GlobalSiloStringStorer) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.storer = storer
		return nil
	}
}

// OptionVerifier sets a verifier as the implementation on MarcoPoller
func OptionVerifier(verifier Verifier) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.verifier = verifier
		return nil
	}
}

// OptionPollVerifier provides a pollVerifier implementation to MarcoPoller
func OptionPollVerifier(pollVerifier PollVerifier) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.pollVerifier = pollVerifier
		return nil
	}
}

// OptionDebug enables debug logging
func OptionDebug(debug bool) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.debug = debug
		return nil
	}
}

// New returns a new MarcoPoller with the default slack client and datastoredb implementations
func New(slackToken string, slackSigningSecret string, datastoreProjectID string, gcloudClientOpts ...option.ClientOption) (mp *MarcoPoller, err error) {
	return NewWithOptions(OptionSlackVerifier(slackSigningSecret), OptionSlackUserFinder(slackToken, cast.ToBool(os.Getenv(DebugEnabledEnv))), OptionSlackDialoguer(slackToken, cast.ToBool(os.Getenv(DebugEnabledEnv))), OptionDatastore(datastoreProjectID, gcloudClientOpts...), OptionPollVerifier(AlwaysValidPollVerifier{}))
}

// NewWithOptions returns a new MarcoPoller with specified options
func NewWithOptions(opts ...Option) (mp *MarcoPoller, err error) {
	mp = new(MarcoPoller)

	for _, apply := range opts {
		err := apply(mp)
		if err != nil {
			return nil, err
		}
	}

	if mp.userFinder == nil {
		return nil, fmt.Errorf("UserFinder is nil after applying all Options. Did you forget to set one?")
	}

	if mp.verifier == nil {
		return nil, fmt.Errorf("Verifier is nil after applying all Options. Did you forget to set one?")
	}

	if mp.storer == nil {
		return nil, fmt.Errorf("Storer is nil after applying all Options. Did you forget to set one?")
	}

	if mp.pollVerifier == nil {
		return nil, fmt.Errorf("PollVerifier is nil after applying all Options. Did you forget to set one?")
	}

	if mp.dialoguer == nil {
		return nil, fmt.Errorf("Dialoguer is nil after applying all Options. Did you forget to set one?")
	}

	mp.meter = opentelemetry.MeterProvider().Meter("github.com/alexandre-normand/marcopoller")
	mp.instruments = newInstruments(mp.meter)

	return mp, err
}

func newInstruments(meter metric.Meter) *instruments {
	defaultLabels := meter.Labels(key.New("name").String(appName))

	pollCounter := meter.NewInt64Counter("pollCount")
	voteCounter := meter.NewInt64Counter("votingCount")

	return &instruments{
		pollCount:   pollCounter.Bind(defaultLabels),
		votingCount: voteCounter.Bind(defaultLabels),
	}
}

// StartPoll handles a slash command request to start a new poll. This function is meant to be wrapped
// by a function that knows how to fetch the slackToken and the slackSigningSecret secrets in order
// to be deployable to gcloud
//
// Example (the companion berglas-backed wrapping implementation):
//   var mp *marcopoller.MarcoPoller
//
//   func init() {
// 		mpoller, err := marcopoller.New(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), os.Getenv(marcopoller.GCPProjectIDEnv))
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to initialize Marco Poller: %s", err.Error()))
// 	 	}
//
// 	 	mp = mpoller
//   }
//
//   func StartPoll(w http.ResponseWriter, r *http.Request) {
//   	 mp.StartPoll(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), w, r)
//   }
func (mp *MarcoPoller) StartPoll(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	err = mp.verifier.Verify(r.Header, body)
	if err != nil {
		log.Printf("Error validating request: %v", err)
		http.Error(w, err.Error(), 403)
		return
	}

	pollText, creator, responseURL, triggerID, err := parseNewPollRequest(string(body))
	if err != nil {
		log.Printf("Error parsing poll request: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}

	// Now that the request is parsed, it's considered accepted and we return a 200 OK to slack
	// to avoid timeouts
	w.WriteHeader(http.StatusOK)

	interactive, question, options, err := parsePollParams(pollText)
	if err != nil {
		showErrorToUser(responseURL, ":warning: Wrong usage. `/poll \"Question\" \"Option 1\" \"Option 2\" ...`")

		return
	}

	if interactive {
		interactivePrompt := createInteractivePollPrompt()
		_, err := mp.dialoguer.OpenView(triggerID, interactivePrompt)
		if err != nil {
			log.Printf("Error opening up interactive prompt for trigger id [%s]: %s", triggerID, err.Error())
			showErrorToUser(responseURL, ":warning: Error opening up interactive prompt. Try again, maybe?")

			return
		}

		return
	}

	mp.createNewPoll(question, options, creator, PollFeatures{}, responseURL, w)
}

// showErrorToUser sends an ephemeral response to a user with a best effort. If there's an error
// sending the message, we log the error but can't do anything more
func showErrorToUser(responseURL string, errorMsg string) {
	actionResponse := ActionResponse{ResponseType: "ephemeral", Text: errorMsg, ReplaceOriginal: false}
	resp, err := req.Post(responseURL, req.BodyJSON(&actionResponse))
	if err != nil || resp.Response().StatusCode != 200 {
		if err != nil {
			log.Printf("Error sending error message [%s] to user: %s", errorMsg, err.Error())
		} else {
			log.Printf("Error sending error message [%s] to user: %s", errorMsg, resp.String())
		}

		return
	}
}

// createNewPoll creates a new poll and handles the persistence and posting to slack
func (mp *MarcoPoller) createNewPoll(question string, options []string, creator string, features PollFeatures, responseURL string, w http.ResponseWriter) {
	pollCreationTime := time.Now()
	poll := Poll{ID: generatePollID(pollCreationTime.Unix()), Question: question, Options: options, Creator: creator, Features: features}

	encodedPoll, err := encodePoll(poll)
	if err != nil {
		log.Printf("Error encoding poll: %s", err.Error())
		showErrorToUser(responseURL, ":warning: Error encoding poll. Please report this at https://github.com/alexandre-normand/marcopoller")
		return
	}

	err = mp.storer.PutSiloString(poll.ID, pollInfoKey, encodedPoll)
	if err != nil {
		log.Printf("Error persisting poll [%s]", poll.ID)
		showErrorToUser(responseURL, ":warning: Error persisting poll. Please try again.")
		return
	}

	actionResponse := ActionResponse{ResponseType: "in_channel", Blocks: renderPoll(poll, map[string][]Voter{}, false)}
	resp, err := req.Post(responseURL, req.BodyJSON(&actionResponse))
	if err != nil || resp.Response().StatusCode != 200 {
		if err != nil {
			log.Printf("Error writing new poll [%s] message: %s", poll.ID, err.Error())
			showErrorToUser(responseURL, ":warning: Error writing new poll to slack. Please try again.")
		} else {
			log.Printf("Error writing new poll [%s] message: %s", poll.ID, resp.String())
			showErrorToUser(responseURL, ":warning: Error writing new poll to slack. Please try again.")
		}

		return
	}

	ctx := context.Background()
	mp.instruments.pollCount.Add(ctx, 1)
}

// slackTimestampToTime converts a slack timestamp string (something like "1556928600.008500") to a time.
func slackTimestampToTime(slackTimestamp string) (parsedTime time.Time) {
	timeAsFloat := cast.ToFloat64(slackTimestamp)
	seconds := int64(timeAsFloat)
	nanoSeconds := int64((timeAsFloat - float64(seconds)) * 1000000000)

	return time.Unix(seconds, nanoSeconds)
}

// generatePollID generates a new poll identifier composed of the timestamp of the initial poll content posted to slack
// and a unique identifier
func generatePollID(timestamp int64) (pollID string) {
	return fmt.Sprintf("%d-%s", timestamp, shortuuid.New())
}

// getPollCreationTime returns the poll creation time from the timestamp part of its identifier. To support legacy
// polls that didn't have this format, a zero time is returned if the format doesn't include the creation time. This
// implies that the poll is "old" and therefore might be treated as expired
func getPollCreationTime(pollID string) (creationTime time.Time) {
	idParts := strings.Split(pollID, "-")
	if len(idParts) <= 1 {
		return time.Unix(0, 0)
	}

	creationTimeSeconds := cast.ToInt64(idParts[0])

	return time.Unix(creationTimeSeconds, 0)
}

// parseNewPollRequest parses a new poll request and returns the pollText, the creator and the response url
func parseNewPollRequest(requestBody string) (pollText string, creator string, responseURL string, triggerID string, err error) {
	params, err := parseRequest(requestBody)
	if err != nil {
		return "", "", "", "", err
	}

	return params[textParam], params[creatorParam], params[responseURLParam], params[triggerIDParam], nil
}

// parseRequest parses a slack request parameters. Since slack request parameters have a single value,
// the parsed query parameters are assumed to have a single value
func parseRequest(requestBody string) (params map[string]string, err error) {
	queryParams, err := url.ParseQuery(string(requestBody))
	if err != nil {
		return params, errors.Wrapf(err, "Error decoding params from body [%s]", requestBody)
	}

	params = make(map[string]string)
	for name, vals := range queryParams {
		params[name] = vals[0]
	}

	return params, nil
}

// parseCallback parses an InteractionCallback. If the payload is empty or there's a parsing error
// a zero-value InteractionCallback is returned along with the error
func parseCallback(payload string) (callback InteractionCallback, err error) {
	if payload == "" {
		return callback, fmt.Errorf("Empty payload")
	}

	err = json.Unmarshal([]byte(payload), &callback)
	if err != nil {
		return callback, err
	}

	return callback, nil
}

// encodePoll encodes a poll to a string. If the poll can't be
// encoded, the error is returned
func encodePoll(poll Poll) (encoded string, err error) {
	m, err := json.Marshal(poll)
	if err != nil {
		return "", err
	}

	return string(m), nil
}

// renderPoll renders a poll with its votes to slack blocks
func renderPoll(poll Poll, votes map[string][]Voter, votingActive bool) (blocks []slack.Block) {
	blocks = make([]slack.Block, 0)

	blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*", poll.Question), false, false), nil, nil))
	blocks = append(blocks, slack.NewDividerBlock())
	for i, opt := range poll.Options {
		optionID := fmt.Sprintf("%d", i)

		var accessory *slack.Accessory
		if !votingActive {
			voteButton := slack.NewButtonBlockElement(formatButtonID(poll.ID, voteButtonValue), optionID, slack.NewTextBlockObject("plain_text", "Vote", false, false))
			voteButton.Style = slack.StylePrimary

			accessory = slack.NewAccessory(voteButton)
		}

		blocks = append(blocks, *slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(" • %s", opt), false, false), nil, accessory))
		if voters, ok := votes[optionID]; ok {
			voteBlocks := make([]slack.MixedElement, 0)
			i := 0
			for i = 0; i < len(voters) && i < 9; i++ {
				voter := voters[i]
				voteBlocks = append(voteBlocks, slack.NewImageBlockElement(voter.avatarURL, voter.name))
			}

			if i == 9 {
				voteBlocks = append(voteBlocks, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("`+ %d`", len(voters)-9), false, false))
			}

			if len(voteBlocks) > 0 {
				blocks = append(blocks, slack.NewContextBlock("", voteBlocks...))
			}
		}
	}

	if !votingActive {
		deleteButton := slack.NewButtonBlockElement(formatButtonID(poll.ID, deleteButtonValue), deleteButtonValue, slack.NewTextBlockObject("plain_text", "Delete poll", false, false))
		deleteButton.Style = slack.StyleDanger

		blocks = append(blocks, slack.NewActionBlock(poll.ID, slack.NewButtonBlockElement(formatButtonID(poll.ID, closeButtonValue), closeButtonValue, slack.NewTextBlockObject("plain_text", "Close voting", false, false)), deleteButton))
		blocks = append(blocks, slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Created by <@%s>", poll.Creator), false, false)))
	} else {
		blocks = append(blocks, slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Created by <@%s> (voting closed)", poll.Creator), false, false)))
	}

	return blocks
}

// formatButtonID formats a button action ID
func formatButtonID(pollID string, action string) (buttonID string) {
	return fmt.Sprintf("%s%s%s", pollID, buttonIDPartDelimiter, action)
}

// createInteractivePollPrompt renders the content of a new poll dialog
func createInteractivePollPrompt() (viewRequest slack.ModalViewRequest) {
	blocks := make([]slack.Block, 0)

	conversationSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeConversations, nil, pollConversationSelectActionID)
	conversationSelect.DefaultToCurrentConversation = true
	conversationSelect.ResponseURLEnabled = true

	blocks = append(blocks, slack.NewInputBlock(pollConversationInputBlockID, slack.NewTextBlockObject("plain_text", "Where do you want to send your poll?", false, false), conversationSelect))
	blocks = append(blocks, slack.NewInputBlock(pollQuestionInputBlockID, slack.NewTextBlockObject("plain_text", "What's your poll about?", false, false), slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "What's your favorite color?", false, false), pollQuestionActionID)))

	answerOptionsInput := slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject("plain_text", "All the color options (one per line)", false, false), pollOptionsActionID)
	answerOptionsInput.Multiline = true
	answerOptionsBlock := slack.NewInputBlock(pollOptionsInputBlockID, slack.NewTextBlockObject("plain_text", "Answer Options", false, false), answerOptionsInput)
	answerOptionsBlock.Hint = slack.NewTextBlockObject("plain_text", "Enter the answer options (one per line)", false, false)
	blocks = append(blocks, answerOptionsBlock)

	featuresInputBlock := slack.NewInputBlock(pollFeaturesInputBlockID, slack.NewTextBlockObject("plain_text", "Options", false, false), slack.NewCheckboxGroupsBlockElement(pollFeaturesActionID, slack.NewOptionBlockObject(multiAnswerOptionID, slack.NewTextBlockObject("plain_text", multiAnswerFeatureValue, false, false), nil)))
	featuresInputBlock.Optional = true
	blocks = append(blocks, featuresInputBlock)

	viewRequest.Type = slack.VTModal
	viewRequest.Title = slack.NewTextBlockObject("plain_text", friendlyName, false, false)
	viewRequest.Close = slack.NewTextBlockObject("plain_text", "Cancel", false, false)
	viewRequest.Submit = slack.NewTextBlockObject("plain_text", "Create Poll", false, false)
	viewRequest.CallbackID = interactivePollCallbackID
	viewRequest.Blocks = slack.Blocks{BlockSet: blocks}

	return viewRequest
}

// HandleInteractions handles user interactions callbacks and processes them according to their
// type and content.
// This function is meant to be wrapped by a function that knows how to fetch the
// slackToken and the slackSigningSecret secrets in order to be deployable to gcloud
//
// Example (the companion berglas-backed wrapping implementation):
//
//   var mp *marcopoller.MarcoPoller
//
//   func init() {
// 		mpoller, err := marcopoller.New(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), os.Getenv(marcopoller.GCPProjectIDEnv))
// 		if err != nil {
// 			panic(fmt.Sprintf("Failed to initialize Marco Poller: %s", err.Error()))
// 	 	}
//
// 	 	mp = mpoller
//   }
//
//   func HandleInteractions(w http.ResponseWriter, r *http.Request) {
//   	 mp.HandleInteractions(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), w, r)
//   }
func (mp *MarcoPoller) HandleInteractions(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), 500)

		return
	}

	err = mp.verifier.Verify(r.Header, body)
	if err != nil {
		log.Printf("Error validating request: %v", err)
		http.Error(w, err.Error(), 403)
		return
	}

	params, err := parseRequest(string(body))
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}

	payload := params["payload"]
	callback, err := parseCallback(payload)
	if err != nil {
		log.Printf("Error parsing interaction callback payload [%s]: %v", payload, err)
		http.Error(w, err.Error(), 400)
		return
	}

	// Request accepted so we send back the 200 OK to slack to avoid timeouts
	w.WriteHeader(http.StatusOK)

	if callback.Type == "block_actions" {
		mp.handlePollInteractions(callback, w)
		return
	} else if callback.Type == "view_submission" {
		mp.handleInteractivePollSubmission(callback, w)

		return
	} else {
		errMsg := fmt.Sprintf("Unknown interaction callback type: %s", callback.Type)
		log.Print(errMsg)
		showErrorToUser(callback.ResponseURL, fmt.Sprintf(":warning: %s", errMsg))
		return
	}
}

// handlePollInteractions handles interactions on a poll (via slack voting or action buttons) and processes that by
// updating the state of a poll and reflecting that state on slack.
func (mp *MarcoPoller) handlePollInteractions(callback InteractionCallback, w http.ResponseWriter) {
	pollID, err := pollID(callback)
	if err != nil {
		log.Printf("Error extracting poll identifier from callback [%v]: %s", callback, err.Error())
		showErrorToUser(callback.ResponseURL, ":warning: Error extracting poll identifier from callback. Please report this at https://github.com/alexandre-normand/marcopoller.")
		return
	}

	// Verify the validity of the poll before we proceed with handling the vote
	err = mp.pollVerifier.Verify(pollID, actionTime(callback))

	// Poll is expired/invalid so handle new votes by telling users and poll deletions by deleting the message
	if err != nil {
		mp.debugf("Invalid vote for poll [%s] with action interaction callback [%v]", pollID, callback)

		showErrorToUser(callback.ResponseURL, fmt.Sprintf(":warning: Sorry, %s", err.Error()))
		return
	}

	encodedPoll, err := mp.storer.GetSiloString(pollID, pollInfoKey)
	if err != nil {
		log.Printf("Error getting existing poll info for id [%s]: %v", pollID, err)
		showErrorToUser(callback.ResponseURL, ":warning: Error getting existing poll info. Please try again.")
		return
	}

	poll, err := decodePoll(encodedPoll)
	if err != nil {
		log.Printf("Error parsing existing poll [%s] for id [%s]: %v", encodedPoll, pollID, err)
		showErrorToUser(callback.ResponseURL, ":warning: Error parsing existing poll info. Please report this issue at https://github.com/alexandre-normand/marcopoller")
		return
	}

	vote := voteValue(callback)

	if vote == deleteButtonValue {
		mp.handlePollDeletion(poll, callback, w)
		return
	} else if vote == closeButtonValue {
		mp.handlePollClosure(poll, callback, w)
		return
	}

	// If poll supports multiple answers, read back the existing votes for the user and toggle the vote
	if poll.Features.MultiAnswers {
		userVotes, err := mp.storer.GetSiloString(poll.ID, callback.User.ID)

		if err != nil && err != datastore.ErrNoSuchEntity {
			log.Printf("Error getting existing votes for user [%s] on poll id [%s]: %v", callback.User.ID, pollID, err)
			showErrorToUser(callback.ResponseURL, ":warning: Error loading existing votes. Please try again.")
			return
		}

		vote = toggleVoteForValue(userVotes, vote)
	}

	err = mp.storer.PutSiloString(poll.ID, callback.User.ID, vote)
	if err != nil {
		log.Printf("Error storing vote [%s] for user [%s] for poll [%s]: %v", vote, callback.User.ID, poll.ID, err)
		showErrorToUser(callback.ResponseURL, ":warning: Error persisting vote. Please try again.")
		return
	}

	votes, err := mp.listVotes(poll.ID)
	if err != nil {
		log.Printf("Error listing votes for poll [%s]: %v", poll.ID, err)
		showErrorToUser(callback.ResponseURL, ":warning: Error listing votes for poll. Please try again.")
		return
	}

	resp, err := req.Post(callback.ResponseURL, req.BodyJSON(&UpdateMessage{ActionResponse: ActionResponse{Blocks: renderPoll(poll, votes, false), ReplaceOriginal: true}}))
	if err != nil || resp.Response().StatusCode != 200 {
		if err != nil {
			log.Printf("Error updating poll [%s] message : %v", poll.ID, err)
			showErrorToUser(callback.ResponseURL, ":warning: Error updating slack message for poll. Please try again.")
		} else {
			log.Printf("Error updating poll [%s] message : %s", poll.ID, resp.String())
			showErrorToUser(callback.ResponseURL, ":warning: Error updating slack message for poll. Please try again.")
		}

		return
	}

	ctx := context.Background()
	mp.instruments.votingCount.Add(ctx, 1)
}

// handleInteractivePollSubmission handles a submission of a modal interactive poll dialog
func (mp *MarcoPoller) handleInteractivePollSubmission(callback InteractionCallback, w http.ResponseWriter) {
	if callback.View.CallbackID != interactivePollCallbackID {
		errMsg := fmt.Sprintf("Invalid view submission with unknown callback id: [%s]", callback.CallbackID)
		log.Print(errMsg)
		showErrorToUser(callback.ResponseURL, fmt.Sprintf(":warning: %s. Please report this at https://github.com/alexandre-normand/marcopoller.", errMsg))
	}

	if callback.View.State == nil {
		errMsg := fmt.Sprintf("Invalid view submission with nil state")
		log.Print(errMsg)
		showErrorToUser(callback.ResponseURL, fmt.Sprintf(":warning: %s. Please report this at https://github.com/alexandre-normand/marcopoller.", errMsg))
	}

	values := callback.View.State.Values

	question := values[pollQuestionInputBlockID][pollQuestionActionID].Value
	rawOptions := values[pollOptionsInputBlockID][pollOptionsActionID].Value
	rawSelectedOptions := values[pollFeaturesInputBlockID][pollFeaturesActionID].SelectedOptions

	selectedOptionsAsMap := make(map[string]bool)
	for _, o := range rawSelectedOptions {
		selectedOptionsAsMap[o.Value] = true
	}

	multiAnswer := selectedOptionsAsMap[multiAnswerOptionID]

	if len(callback.ResponseURLs) < 1 {
		errMsg := "Invalid view submission missing response_urls"
		log.Print(errMsg)
		showErrorToUser(callback.ResponseURL, fmt.Sprintf(":warning: %s. Please report this at https://github.com/alexandre-normand/marcopoller.", errMsg))
	}

	pollOptions := strings.Split(rawOptions, "\n")
	validOptions := make([]string, 0)
	for _, o := range pollOptions {
		if o != "" {
			validOptions = append(validOptions, o)
		}
	}

	mp.createNewPoll(question, validOptions, callback.User.ID, PollFeatures{MultiAnswers: multiAnswer}, callback.ResponseURLs[0].ResponseURL, w)
}

// handlePollDeletion handles a request to delete a poll
func (mp *MarcoPoller) handlePollDeletion(poll Poll, callback InteractionCallback, w http.ResponseWriter) {
	pollID, err := pollID(callback)
	if err != nil {
		log.Printf("Error extracting poll identifier from callback [%v]: %s", callback, err.Error())
		showErrorToUser(callback.ResponseURL, ":warning: Error extracting poll identifier from callback. Please report this issue at https://github.com/alexandre-normand/marcopoller.")
		return
	}

	if poll.Creator == callback.User.ID {
		// Delete poll and votes from storage
		err := mp.deletePoll(pollID)
		if err != nil {
			log.Printf("Error deleting poll [%s]: %s", pollID, err.Error())
			showErrorToUser(callback.ResponseURL, ":warning: Error deleting poll. Please try again")
			return
		}

		resp, err := req.Post(callback.ResponseURL, req.BodyJSON(&DeleteMessage{DeleteOriginal: true}))
		if err != nil || resp.Response().StatusCode != 200 {
			if err != nil {
				log.Printf("Error deleting message: %v", err)
				showErrorToUser(callback.ResponseURL, ":warning: Error deleting message from slack")
			} else {
				log.Printf("Error deleting message: %s", resp.String())
				showErrorToUser(callback.ResponseURL, ":warning: Error deleting message from slack")
			}

			return
		}

		return
	}

	showErrorToUser(callback.ResponseURL, fmt.Sprintf(":warning: Only the poll creator (<@%s>) is allowed to delete the poll", poll.Creator))
	return
}

// handlePollClosure handles a request to close a poll
func (mp *MarcoPoller) handlePollClosure(poll Poll, callback InteractionCallback, w http.ResponseWriter) {
	pollID, err := pollID(callback)
	if err != nil {
		log.Printf("Error extracting poll identifier from callback [%v]: %s", callback, err.Error())
		showErrorToUser(callback.ResponseURL, ":warning: Error extracting poll identifier from callback. Please report this issue at https://github.com/alexandre-normand/marcopoller.")
		return
	}

	if poll.Creator == callback.User.ID {
		votes, err := mp.listVotes(poll.ID)
		if err != nil {
			log.Printf("Error listing votes for poll [%s]: %v", poll.ID, err)
			showErrorToUser(callback.ResponseURL, ":warning: Error listing votes for poll. Please try again")
			return
		}

		// Post the final poll update to slack
		resp, err := req.Post(callback.ResponseURL, req.BodyJSON(&UpdateMessage{ActionResponse: ActionResponse{Blocks: renderPoll(poll, votes, true), ReplaceOriginal: true}}))
		if err != nil || resp.Response().StatusCode != 200 {
			if err != nil {
				log.Printf("Error updating poll [%s] message : %v", poll.ID, err)
				showErrorToUser(callback.ResponseURL, ":warning: Error updating poll message. Please try again")
			} else {
				log.Printf("Error updating poll [%s] message : %s", poll.ID, resp.String())
				showErrorToUser(callback.ResponseURL, ":warning: Error updating poll message. Please try again")
			}

			return
		}

		// Delete poll and votes from storage
		err = mp.deletePoll(pollID)
		if err != nil {
			log.Printf("Error deleting poll [%s]: %s", pollID, err.Error())
			return
		}

		return
	}

	showErrorToUser(callback.ResponseURL, fmt.Sprintf(":warning: Only the poll creator (<@%s>) is allowed to close the poll", poll.Creator))
	return
}

// toggleVoteForValue toggles a vote from an existing delimited string of all of a user's votes
func toggleVoteForValue(userVotes string, voteToToggle string) (newUserVotes string) {
	voteMap := make(map[string]bool)

	if userVotes != "" {
		values := strings.Split(userVotes, voteDelimiter)

		for _, v := range values {
			voteMap[v] = true
		}
	}

	if _, exists := voteMap[voteToToggle]; exists {
		delete(voteMap, voteToToggle)
	} else {
		voteMap[voteToToggle] = true
	}

	allVotes := make([]string, 0, len(voteMap))
	for vote := range voteMap {
		allVotes = append(allVotes, vote)
	}

	sort.Strings(allVotes)
	return strings.Join(allVotes, voteDelimiter)
}

// listVotes returns the list of votes: a map of vote values for a poll ID to the array of voters. If an error occurs
// getting the votes or the voter info, that error is returned.
func (mp *MarcoPoller) listVotes(pollID string) (votes map[string][]Voter, err error) {
	values, err := mp.storer.ScanSilo(pollID)
	if err != nil {
		return votes, err
	}

	// Filter out the pollInfoKey
	voteValues := make(map[string]string)
	for k, v := range values {
		if k != pollInfoKey {
			voteValues[k] = v
		}
	}

	votes = make(map[string][]Voter)
	for userID, userVoting := range voteValues {
		user, err := mp.userFinder.GetUserInfo(userID)
		if err != nil {
			return nil, err
		}

		userVotes := strings.Split(userVoting, voteDelimiter)
		for _, value := range userVotes {
			if _, ok := votes[value]; !ok {
				votes[value] = make([]Voter, 0)
			}

			voter := Voter{userID: userID, avatarURL: user.Profile.Image24, name: user.RealName}

			votes[value] = append(votes[value], voter)
		}
	}

	return votes, nil
}

// deletePoll removes a poll and all of its associated data from storage
func (mp *MarcoPoller) deletePoll(pollID string) (err error) {
	values, err := mp.storer.ScanSilo(pollID)
	if err != nil {
		return err
	}

	for k := range values {
		// If we see an error, we keep it but still continue deleting entries
		if err == nil {
			err = mp.storer.DeleteSiloString(pollID, k)
		}
	}

	return err
}

// decodePoll decodes a poll from a encoded string value.
func decodePoll(encoded string) (poll Poll, err error) {
	err = json.Unmarshal([]byte(encoded), &poll)

	return poll, err
}

// voteValue returns the vote value in a given interaction callback
func voteValue(callback InteractionCallback) (vote string) {
	return callback.ActionCallback.BlockActions[0].Value
}

// pollID returns the poll ID in a given interaction callback
func pollID(callback InteractionCallback) (pollID string, err error) {
	actionID := callback.ActionCallback.BlockActions[0].ActionID
	parts := strings.Split(actionID, buttonIDPartDelimiter)
	if len(parts) < 2 {
		return "", fmt.Errorf("Invalid format, expected [string%saction] but got [%s]", buttonIDPartDelimiter, actionID)
	}
	return parts[0], nil
}

// actionTime returns the action time in a given interaction callback
func actionTime(callback InteractionCallback) (actionTime time.Time) {
	timestamp := callback.ActionCallback.BlockActions[0].ActionTs

	return slackTimestampToTime(timestamp)
}

// parsePollParams parses poll parameters. The expected format is: "Some question" "Option 1" "Option 2" "Option 3"
func parsePollParams(rawPoll string) (interactiveReq bool, pollQuestion string, options []string, err error) {
	inQuote := false
	params := make([]string, 0)
	var strBuilder strings.Builder

	// If no parameters provided, this means it's going to be a request for an interactive poll dialog
	if len(strings.TrimSpace(rawPoll)) == 0 {
		return true, "", nil, nil
	}

	// Sacrifice some fidelity for convenience by normalizing smart quotes to standard quotes before parsing so that people
	// having smart quoting enabled don't feel frustrated when the poll doesn't render correctly
	normalizedPollReq := normalizePollRequest(rawPoll)

	for _, r := range normalizedPollReq {
		switch {
		case r == '"' && !inQuote:
			{
				inQuote = true
			}

		case r == '"' && inQuote:
			{
				inQuote = false
				params = append(params, strBuilder.String())
				strBuilder.Reset()
			}

		case unicode.IsSpace(r) && !inQuote:
			{
				param := strBuilder.String()
				if len(param) > 0 {
					params = append(params, param)
				}

				strBuilder.Reset()
			}
		default:
			{
				strBuilder.WriteRune(r)
			}
		}
	}

	param := strBuilder.String()
	if len(param) > 0 {
		params = append(params, param)
	}

	if len(params) < 2 {
		return false, "", nil, fmt.Errorf("Missing parameters in string [%s]", rawPoll)
	}

	return false, params[0], params[1:], nil
}

// normalizePollRequest applies a few operation to normalize a polling request prior to parsing:
//  * Replace opening curly quotes by the standard quote character
//  * Replace closing curly quotes by the standard quote character
func normalizePollRequest(rawRequest string) (normalizedReq string) {
	normalizedPoll := rawRequest
	normalizedPoll = strings.Replace(normalizedPoll, "“", "\"", -1)
	normalizedPoll = strings.Replace(normalizedPoll, "”", "\"", -1)

	return normalizedPoll
}

// debugf logs a debug line after checking if the configuration is in debug mode
func (mp *MarcoPoller) debugf(format string, v ...interface{}) {
	if mp.debug {
		log.Printf(format, v...)
	}
}

// DeleteExpiredPolls removes all poll data (content and associated votes) without deleting
// the slack message holding the most recent snapshot of the poll. The deletionTime should
// be the current time except for synthetic scenarios like tests
func (mp *MarcoPoller) DeleteExpiredPolls(deletionTime time.Time) (count int, err error) {
	count = 0
	polls, err := mp.storer.GlobalScan()
	if err != nil {
		return 0, err
	}

	for pollID := range polls {
		if mp.pollVerifier.Verify(pollID, deletionTime) != nil {
			err := mp.deletePoll(pollID)
			if err != nil {
				return count, err
			}

			count++
		}
	}

	return count, nil
}
