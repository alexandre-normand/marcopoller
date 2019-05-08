package marcopoller

import (
	"encoding/json"
	"fmt"
	_ "github.com/GoogleCloudPlatform/berglas/pkg/auto"
	"github.com/alexandre-normand/slackscot/store"
	"github.com/alexandre-normand/slackscot/store/datastoredb"
	"github.com/lithammer/shortuuid"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	storer             store.GlobalSiloStringStorer
	messenger          Messenger
	userFinder         UserFinder
	slackSigningSecret string
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

// OptionDatastore sets a datastoredb as the implementation of GlobalSiloStringStorer
func OptionDatastore(datastoreProjectID string) Option {
	return func(mp *MarcoPoller) (err error) {
		mp.storer, err = datastoredb.New(name, datastoreProjectID)
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

// New returns a new MarcoPoller with the default slack client and datastoredb implementations
func New(slackToken string, slackSigningSecret string, datastoreProjectID string) (mp *MarcoPoller, err error) {
	return NewWithOptions(slackSigningSecret, OptionSlackClient(slackToken, cast.ToBool(os.Getenv(DebugEnabledEnv))), OptionDatastore(datastoreProjectID))
}

// NewWithOptions returns a new MarcoPoller with specified options
func NewWithOptions(slackSigningSecret string, opts ...Option) (mp *MarcoPoller, err error) {
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

	return mp, err
}

// StartPoll handles a slash command request to start a new poll. This function is meant to be wrapped
// by a function that knows how to fetch the slackToken and the slackSigningSecret secrets in order
// to be deployable to gcloud
//
// Example (the provided berglas-backed wrapping implementation):
//   func StartPollBerglas(w http.ResponseWriter, r *http.Request) {
//  	 StartPoll(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), w, r)
//   }
func (mp *MarcoPoller) StartPoll(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), 500)
	}

	err = verifyRequest(mp.slackSigningSecret, r.Header, body)
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

	poll := Poll{ID: shortuuid.New(), MsgID: MsgIdentifier{ChannelID: channel, Timestamp: "TBD"}, Question: question, Options: options, Creator: creator}
	_, timestamp, err := mp.messenger.PostMessage(channel, slack.MsgOptionBlocks(renderPoll(poll, map[string][]Voter{})...))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Tack on the timestamp to the message id
	poll.MsgID.Timestamp = timestamp

	encodedPoll, err := encodePoll(poll)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Printf("Created poll [%s]", encodedPoll)
	err = mp.storer.PutSiloString(poll.ID, pollInfoKey, encodedPoll)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
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

// verifyRequest verifies the slack request's authenticity (https://api.slack.com/docs/verifying-requests-from-slack). If the request
// can't be verified or if it fails verification, an error is returned. For a verified valid request, nil is returned
func verifyRequest(slackSigningSecret string, header http.Header, body []byte) (err error) {
	verifier, err := slack.NewSecretsVerifier(header, slackSigningSecret)
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

	blocks = append(blocks, *slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s*", poll.Question), false, false), nil, nil))
	blocks = append(blocks, *slack.NewDividerBlock())
	for i, opt := range poll.Options {
		optionID := fmt.Sprintf("%d", i)
		blocks = append(blocks, *slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(" • %s", opt), false, false), nil, slack.NewButtonBlockElement(poll.ID, optionID, slack.NewTextBlockObject("plain_text", "Vote", false, false))))
		if voters, ok := votes[optionID]; ok {
			voteBlocks := make([]slack.BlockObject, 0)
			i := 0
			for i = 0; i < len(voters) && i < 9; i++ {
				voter := voters[i]
				voteBlocks = append(voteBlocks, slack.NewImageBlockObject(voter.avatarURL, voter.name))
			}

			if i == 9 {
				voteBlocks = append(voteBlocks, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("`+ %d`", len(voters)-9), false, false))
			}

			blocks = append(blocks, *slack.NewContextBlock("", voteBlocks...))
		}
	}

	blocks = append(blocks, *slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " ", false, false), nil, slack.NewButtonBlockElement(poll.ID, deleteValue, slack.NewTextBlockObject("plain_text", "Delete poll", false, false))))
	blocks = append(blocks, *slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Created by <@%s>", poll.Creator), false, false)))

	return blocks
}

// RegisterVote handles a slack voting button action and processes it by storing the vote and updating the
// poll results (the slack message) accordingly. This function is meant to be wrapped
// by a function that knows how to fetch the slackToken and the slackSigningSecret secrets in order
// to be deployable to gcloud
//
// Example (the provided berglas-backed wrapping implementation):
//   func RegisterVoteBerglas(w http.ResponseWriter, r *http.Request) {
//   	 RegisterVote(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), w, r)
//   }
func (mp *MarcoPoller) RegisterVote(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), 500)
	}

	err = verifyRequest(mp.slackSigningSecret, r.Header, body)
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
			_, _, err = mp.messenger.DeleteMessage(poll.MsgID.ChannelID, poll.MsgID.Timestamp)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
				http.Error(w, err.Error(), 500)
				return
			}

			return
		} else {
			_, _, err = mp.messenger.PostMessage(callback.Channel.ID, slack.MsgOptionPostEphemeral(callback.User.ID), slack.MsgOptionText(fmt.Sprintf(":warning: Only the poll creator (<@%s>) is allowed to delete the poll", poll.Creator), false))
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
			return map[string][]Voter{}, err
		}

		if _, ok := votes[value]; !ok {
			votes[value] = make([]Voter, 0)
		}

		voter := Voter{userID: userID, avatarURL: user.Profile.Image24, name: user.RealName}

		votes[value] = append(votes[value], voter)
	}

	return votes, nil
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

// parsePollParams parses poll parameters. The expected format is: "Some question" "Option 1" "Option 2" "Option 3"
func parsePollParams(rawPoll string) (pollQuestion string, options []string, err error) {
	inQuote := false
	params := make([]string, 0)
	var strBuilder strings.Builder

	for _, r := range rawPoll {
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

	if len(params) == 0 {
		return "", nil, fmt.Errorf("No parameters in string [%s]", rawPoll)
	}

	return params[0], params[1:], nil
}
