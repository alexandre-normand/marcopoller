package marcopoller

import (
	"encoding/json"
	"fmt"
	"github.com/alexandre-normand/slackscot/store"
	"github.com/alexandre-normand/slackscot/store/datastoredb"
	"github.com/lithammer/shortuuid"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"
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
	name        = "marcoPoller"
	deleteValue = "delete"
)

// Slack slash command parameter names
const (
	textParam    = "text"
	channelParam = "channel_id"
	creatorParam = "user_id"
)

// Callback represents a slack interaction callback payload
type Callback struct {
	Type        slack.InteractionType `json:"type"`
	Team        slack.Team            `json:"team"`
	User        slack.User            `json:"user"`
	ApiAppID    string                `json:"api_app_id"`
	Token       string                `json:"token"`
	TriggerID   string                `json:"trigger_id"`
	ResponseURL string                `json:"response_url"`
	ActionTs    string                `json:"action_ts"`
	Channel     slack.Channel         `json:"channel"`
	Name        string                `json:"name"`
	Value       string                `json:"value"`
	Actions     []Action              `json:"actions"`
}

// Action represents a triggered action value in a slack callback
type Action struct {
	Type     string                `json:"type"`
	BlockID  string                `json:"block_id"`
	ActionID string                `json:"action_id"`
	Text     slack.TextBlockObject `json:"text"`
	Value    string                `json:"value"`
	ActionTs string                `json:"action_ts"`
}

// Poll represents a poll
type Poll struct {
	ID       string        `json:"id"`
	MsgID    MsgIdentifier `json:"msgID"`
	Question string        `json:"question"`
	Options  []string      `json:"options"`
	Creator  string        `json:"creator"`
}

// MsgIdentifier represents a slack message identifier (relative to the workspace the app interacts with)
type MsgIdentifier struct {
	ChannelID string `json:"channelID"`
	Timestamp string `json:"timestamp"`
}

// Voter represents a voting user
type Voter struct {
	userID    string
	avatarURL string
	name      string
}

// MarcoPoller represents a Marco Poller instance
type MarcoPoller struct {
	storer       store.GlobalSiloStringStorer
	messenger    Messenger
	userFinder   UserFinder
	verifier     Verifier
	pollVerifier PollVerifier
	debug        bool
}

// Messenger is implemented by any value that has the PostMessage, DeleteMessage, UpdateMessage methods
type Messenger interface {
	// PostMessage sends a message to a channel. See https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	// PostEphemeral sends an ephemeral message to a user in a channel. See https://godoc.org/github.com/nlopes/slack#Client.PostEphemeral
	PostEphemeral(channelID, userID string, options ...slack.MsgOption) (string, error)
	// DeleteMessage deletes a message in a channel. See https://godoc.org/github.com/nlopes/slack#Client.DeleteMessage
	DeleteMessage(channel, messageTimestamp string) (string, string, error)
	// UpdateMessage updates a message in a channel. See https://godoc.org/github.com/nlopes/slack#Client.UpdateMessage
	UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
}

// UserFinder is implemented by any value that has the GetInfo method
type UserFinder interface {
	// GetUserInfo will retrieve the complete user information. See https://godoc.org/github.com/nlopes/slack#Client.GetUserInfo
	GetUserInfo(user string) (*slack.User, error)
}

// Verifier is implemented by any value that has the Verify method
type Verifier interface {
	Verify(header http.Header, body []byte) (err error)
}

// SlackVerifier represents a slack verifier backed by github.com/nlopes/slack
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

// OptionSlackClient sets a nlopes/slack.Client as the implementation of Messenger
func OptionSlackClient(slackToken string, debug bool) Option {
	return func(mp *MarcoPoller) (err error) {
		sc := slack.New(slackToken, slack.OptionDebug(debug))
		mp.messenger = sc
		mp.userFinder = sc
		return nil
	}
}

// OptionSlackMessenger sets a nlopes/slack.Client as the implementation of Messenger
func OptionSlackMessenger(token string, debug bool) Option {
	return func(mp *MarcoPoller) (err error) {
		sc := slack.New(token, slack.OptionDebug(debug))
		mp.messenger = sc
		return nil
	}
}

// OptionSlackUserFinder sets a nlopes/slack.Client as the implementation of UserFinder
func OptionSlackUserFinder(token string, debug bool) Option {
	return func(mp *MarcoPoller) (err error) {
		sc := slack.New(token, slack.OptionDebug(debug))
		mp.userFinder = sc
		return nil
	}
}

// OptionSlackVerifier sets a nlopes/slack.Client as the implementation of Messenger
func OptionSlackVerifier(slackSigningSecret string) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.verifier = &SlackVerifier{slackSigningSecret: slackSigningSecret}

		return nil
	}
}

// OptionDatastore sets a datastoredb as the implementation of GlobalSiloStringStorer
func OptionDatastore(datastoreProjectID string, gcloudClientOpts ...option.ClientOption) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.storer, err = datastoredb.New(name, datastoreProjectID, gcloudClientOpts...)
		if err != nil {
			return errors.Wrapf(err, "Error initializing datastore persistence on project [%s]", datastoreProjectID)
		}

		return nil
	}
}

// OptionMessenger sets a messenger as the implementation on MarcoPoller
func OptionMessenger(messenger Messenger) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.messenger = messenger
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

// OptionVerifier provides a pollVerifier implementation to MarcoPoller
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
	return NewWithOptions(OptionSlackVerifier(slackSigningSecret), OptionSlackClient(slackToken, cast.ToBool(os.Getenv(DebugEnabledEnv))), OptionDatastore(datastoreProjectID, gcloudClientOpts...), OptionPollVerifier(AlwaysValidPollVerifier{}))
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

	if mp.messenger == nil {
		return nil, fmt.Errorf("Messenger is nil after applying all Options. Did you forget to set one?")
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

	return mp, err
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

	pollText, creator, channel, err := parseNewPollRequest(string(body))
	if err != nil {
		log.Printf("Error parsing poll request: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}

	question, options, err := parsePollParams(pollText)
	if err != nil {
		_, err = mp.messenger.PostEphemeral(channel, creator, slack.MsgOptionText(":warning: Wrong usage. `/poll \"Question\" \"Option 1\" \"Option 2\" ...`", false))
		if err != nil {
			log.Printf("Error sending message: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		return
	}

	pollCreationTime := time.Now()
	poll := Poll{ID: generatePollID(pollCreationTime.Unix()), MsgID: MsgIdentifier{ChannelID: channel, Timestamp: "TBD"}, Question: question, Options: options, Creator: creator}
	_, timestamp, err := mp.messenger.PostMessage(channel, slack.MsgOptionAsUser(false), slack.MsgOptionBlocks(renderPoll(poll, map[string][]Voter{})...))
	if err != nil {
		if err.Error() == "channel_not_found" {
			w.Header().Set("Content-type", "application/json")
			fmt.Fprintf(w, "{ \"response_type\": \"ephemeral\", \"text\": \"I don't have access to this conversation. Try adding me to the apps before creating a poll!\" }")
			return
		} else {
			log.Printf("Error sending poll message: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
	}

	// Tack on the timestamp to the message id and generate the poll identifier
	poll.MsgID.Timestamp = timestamp

	encodedPoll, err := encodePoll(poll)
	if err != nil {
		log.Printf("Error encoding poll: %s", err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	err = mp.storer.PutSiloString(poll.ID, pollInfoKey, encodedPoll)
	if err != nil {
		log.Printf("Error persisting poll [%s]", poll.ID)
		http.Error(w, err.Error(), 500)
		return
	}
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

// parseNewPollRequest parses a new poll request and returns the pollText, the creator and the channel
func parseNewPollRequest(requestBody string) (pollText string, creator string, channel string, err error) {
	params, err := parseRequest(requestBody)
	if err != nil {
		return "", "", "", err
	}

	return params[textParam], params[creatorParam], params[channelParam], nil
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

// parseCallback parses a slack callback. If the payload is empty or there's a parsing error
// a zero-value callback is returned along with the error
func parseCallback(payload string) (callback Callback, err error) {
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
func renderPoll(poll Poll, votes map[string][]Voter) (blocks []slack.Block) {
	blocks = make([]slack.Block, 0)

	blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*", poll.Question), false, false), nil, nil))
	blocks = append(blocks, slack.NewDividerBlock())
	for i, opt := range poll.Options {
		optionID := fmt.Sprintf("%d", i)
		voteButton := slack.NewButtonBlockElement(poll.ID, optionID, slack.NewTextBlockObject("plain_text", "Vote", false, false))
		voteButton.Style = slack.StylePrimary
		blocks = append(blocks, *slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(" • %s", opt), false, false), nil, slack.NewAccessory(voteButton)))
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

	deleteButton := slack.NewButtonBlockElement(poll.ID, deleteValue, slack.NewTextBlockObject("plain_text", "Delete poll", false, false))
	deleteButton.Style = slack.StyleDanger
	blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " ", false, false), nil, slack.NewAccessory(deleteButton)))
	blocks = append(blocks, slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Created by <@%s>", poll.Creator), false, false)))

	return blocks
}

// RegisterVote handles a slack voting button action and processes it by storing the vote and updating the
// poll results (the slack message) accordingly. This function is meant to be wrapped
// by a function that knows how to fetch the slackToken and the slackSigningSecret secrets in order
// to be deployable to gcloud
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
//   func RegisterVote(w http.ResponseWriter, r *http.Request) {
//   	 mp.RegisterVote(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), w, r)
//   }
func (mp *MarcoPoller) RegisterVote(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Error parsing callback payload: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}

	pollID := pollID(callback)

	// Verify the validity of the poll before we proceed with handling the vote
	err = mp.pollVerifier.Verify(pollID, actionTime(callback))

	// Poll is expired/invalid so handle new votes by telling users and poll deletions by deleting the message
	if err != nil {
		mp.debugf("Invalid vote for poll [%s] with action callback [%v]", pollID, callback)

		_, err = mp.messenger.PostEphemeral(callback.Channel.ID, callback.User.ID, slack.MsgOptionText(fmt.Sprintf(":warning: Sorry, %s", err.Error()), false))
		if err != nil {
			log.Printf("Error sending message: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		return
	}

	encodedPoll, err := mp.storer.GetSiloString(pollID, pollInfoKey)
	if err != nil {
		log.Printf("Error getting existing poll info for id [%s]: %v", pollID, err)
		http.Error(w, err.Error(), 500)
		return
	}

	poll, err := decodePoll(encodedPoll)
	if err != nil {
		log.Printf("Error parsing existing poll [%s] for id [%s]: %v", encodedPoll, pollID, err)
		http.Error(w, err.Error(), 500)
		return
	}

	vote := voteValue(callback)
	if vote == deleteValue {
		if poll.Creator == callback.User.ID {
			// Delete poll and votes from storage
			err := mp.deletePoll(pollID)
			if err != nil {
				log.Printf("Error deleting poll [%s]: %s", pollID, err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			_, _, err = mp.messenger.DeleteMessage(poll.MsgID.ChannelID, poll.MsgID.Timestamp)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
				http.Error(w, err.Error(), 500)
				return
			}

			return
		} else {
			_, err = mp.messenger.PostEphemeral(callback.Channel.ID, callback.User.ID, slack.MsgOptionText(fmt.Sprintf(":warning: Only the poll creator (<@%s>) is allowed to delete the poll", poll.Creator), false))
			if err != nil {
				log.Printf("Error sending message: %v", err)
				http.Error(w, err.Error(), 500)
				return
			}

			return
		}
	}

	err = mp.storer.PutSiloString(poll.ID, callback.User.ID, vote)
	if err != nil {
		log.Printf("Error storing vote [%s] for user [%s] for poll [%s]: %v", vote, callback.User.ID, poll.ID, err)
		http.Error(w, err.Error(), 500)
		return
	}

	votes, err := mp.listVotes(poll.ID)
	if err != nil {
		log.Printf("Error listing votes for poll [%s]: %v", poll.ID, err)
		http.Error(w, err.Error(), 500)
		return
	}

	_, _, _, err = mp.messenger.UpdateMessage(poll.MsgID.ChannelID, poll.MsgID.Timestamp, slack.MsgOptionBlocks(renderPoll(poll, votes)...))
	if err != nil {
		log.Printf("Error updating poll [%s] message : %v", poll.ID, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// listVotes returns the list of votes: a map of vote values for a poll ID to the array of voters. If an error occurs
// getting the votes or the voter info, that error is returned.
func (mp *MarcoPoller) listVotes(pollID string) (votes map[string][]Voter, err error) {
	values, err := mp.storer.ScanSilo(pollID)
	if err != nil {
		return votes, err
	}

	// Omit the special pollInfoKey from the scan results since that's the
	// only one not a vote
	delete(values, pollInfoKey)

	votes = make(map[string][]Voter)
	for userID, value := range values {
		user, err := mp.userFinder.GetUserInfo(userID)
		if err != nil {
			return nil, err
		}

		if _, ok := votes[value]; !ok {
			votes[value] = make([]Voter, 0)
		}

		voter := Voter{userID: userID, avatarURL: user.Profile.Image24, name: user.RealName}

		votes[value] = append(votes[value], voter)
	}

	return votes, nil
}

// deletePoll removes a poll and all of its associated data from storage
func (mp *MarcoPoller) deletePoll(pollID string) (err error) {
	values, err := mp.storer.ScanSilo(pollID)
	if err != nil {
		return err
	}

	for k, _ := range values {
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

// voteValue returns the vote value in a given callback
func voteValue(callback Callback) (vote string) {
	return callback.Actions[0].Value
}

// pollID returns the poll ID in a given callback
func pollID(callback Callback) (pollID string) {
	return callback.Actions[0].ActionID
}

// actionTime returns the action time in a given callback
func actionTime(callback Callback) (actionTime time.Time) {
	timestamp := callback.Actions[0].ActionTs

	return slackTimestampToTime(timestamp)
}

// parsePollParams parses poll parameters. The expected format is: "Some question" "Option 1" "Option 2" "Option 3"
func parsePollParams(rawPoll string) (pollQuestion string, options []string, err error) {
	inQuote := false
	params := make([]string, 0)
	var strBuilder strings.Builder

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
		return "", nil, fmt.Errorf("Missing parameters in string [%s]", rawPoll)
	}

	return params[0], params[1:], nil
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

	for pollID, _ := range polls {
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
