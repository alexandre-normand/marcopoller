package poller

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
	slackTokenEnv    = "SLACK_TOKEN"
	projectIDEnv     = "PROJECT_ID"
	signingSecretEnv = "SIGNING_SECRET"
	debugEnabledEnv  = "DEBUG"
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

// StartPoll handles a slash command request to start a new poll
func StartPoll(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), 500)
	}

	err = verifyRequest(r.Header, body)
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

	storer, err := datastoredb.New(name, os.Getenv(projectIDEnv))
	if err != nil {
		log.Printf("Error creating persistence: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	sc := slack.New(os.Getenv(slackTokenEnv), slack.OptionDebug(cast.ToBool(os.Getenv(debugEnabledEnv))))

	question, options, err := parsePollParams(pollText)
	if err != nil {
		_, _, err = sc.PostMessage(channel, slack.MsgOptionPostEphemeral(creator), slack.MsgOptionText(":warning: Wrong usage. `/poll \"Question\" \"Option 1\" \"Option 2\" ...`", false))
		if err != nil {
			log.Printf("Error sending message: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		return
	}

	poll := Poll{ID: shortuuid.New(), MsgID: MsgIdentifier{ChannelID: channel, Timestamp: "TBD"}, Question: question, Options: options, Creator: creator}
	_, timestamp, err := sc.PostMessage(channel, slack.MsgOptionBlocks(renderPoll(poll, map[string][]Voter{})...))
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
	err = storer.PutSiloString(poll.ID, pollInfoKey, encodedPoll)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func parseNewPollRequest(requestBody string) (pollText string, creator string, channel string, err error) {
	params, err := parseRequest(requestBody)
	if err != nil {
		return "", "", "", err
	}

	return params[textParam], params[creatorParam], params[channelParam], nil
}

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
func verifyRequest(header http.Header, body []byte) (err error) {
	verifier, err := slack.NewSecretsVerifier(header, os.Getenv(signingSecretEnv))
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
// poll results (the slack message) accordingly
func RegisterVote(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), 500)
	}

	err = verifyRequest(r.Header, body)
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

	sc := slack.New(os.Getenv(slackTokenEnv), slack.OptionDebug(cast.ToBool(os.Getenv(debugEnabledEnv))))

	payload := params["payload"]
	callback, err := parseCallback(payload)
	if err != nil {
		log.Printf("Error parsing payload: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}

	storer, err := datastoredb.New(name, os.Getenv(projectIDEnv))
	if err != nil {
		log.Printf("Error creating persistence: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	pollID := pollID(callback)
	encodedPoll, err := storer.GetSiloString(pollID, pollInfoKey)
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
			_, _, err = sc.DeleteMessage(poll.MsgID.ChannelID, poll.MsgID.Timestamp)
			if err != nil {
				log.Printf("Error deleting message: %v", err)
				http.Error(w, err.Error(), 500)
				return
			}

			return
		} else {
			_, _, err = sc.PostMessage(callback.Channel.ID, slack.MsgOptionPostEphemeral(callback.User.ID), slack.MsgOptionText(fmt.Sprintf(":warning: Only the poll creator (<@%s>) is allowed to delete the poll", poll.Creator), false))
			if err != nil {
				log.Printf("Error sending message: %v", err)
				http.Error(w, err.Error(), 500)
				return
			}

			return
		}
	}

	err = storer.PutSiloString(poll.ID, callback.User.ID, vote)
	if err != nil {
		log.Printf("Error storing vote [%s] for user [%s] for poll [%s]: %v", vote, callback.User.ID, poll.ID, err)
		http.Error(w, err.Error(), 500)
		return
	}

	votes, err := listVotes(sc, storer, poll.ID)
	if err != nil {
		log.Printf("Error listing votes for poll [%s]: %v", poll.ID, err)
		http.Error(w, err.Error(), 500)
		return
	}

	_, _, _, err = sc.UpdateMessage(poll.MsgID.ChannelID, poll.MsgID.Timestamp, slack.MsgOptionBlocks(renderPoll(poll, votes)...))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// listVotes returns the list of votes: a map of vote values for a poll ID to the array of voters. If an error occurs
// getting the votes or the voter info, that error is returned.
func listVotes(slackClient *slack.Client, storer store.GlobalSiloStringStorer, pollID string) (votes map[string][]Voter, err error) {
	values, err := storer.ScanSilo(pollID)
	if err != nil {
		return votes, err
	}

	// Omit the special pollInfoKey from the scan results since that's the
	// only one not a vote
	delete(values, pollInfoKey)

	votes = make(map[string][]Voter)
	for userID, value := range values {
		user, err := slackClient.GetUserInfo(userID)
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
