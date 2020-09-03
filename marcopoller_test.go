package marcopoller_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/alexandre-normand/marcopoller"
	mmocks "github.com/alexandre-normand/marcopoller/mocks"
	"github.com/alexandre-normand/slackscot/store/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidNewPoll(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	body := fmt.Sprintf("token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%%2Fpoll&text=%%22To%%20do%%20or%%20not%%20to%%20do%%3F%%22%%20%%22Do%%22%%20%%22Not%%20Do%%22&response_url=%s&trigger_id=someTriggerID", server.URL)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("PutSiloString", mock.Anything, "pollInfo", mock.MatchedBy(func(val string) bool {
		match, _ := regexp.MatchString("{\"id\":\".*\",\"question\":\"To do or not to do\\?\",\"options\":\\[\"Do\",\"Not Do\"\\],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", val)
		return match
	})).Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Regexp(t, regexp.MustCompile("\\{\"response_type\":\"in_channel\",\"blocks\":.*,\"replace_original\":false}"), slackRequest)
}

func TestPollWithCurlyQuotes(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	body := fmt.Sprintf("token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%%2Fpoll&text=%%22To%%20do%%20or%%20not%%20to%%20do%%3F%%22%%20%%22Do%%22%%20%%22Not%%20Do%%22&response_url=%s&trigger_id=someTriggerID", server.URL)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("PutSiloString", mock.Anything, "pollInfo", mock.MatchedBy(func(val string) bool {
		match, _ := regexp.MatchString("{\"id\":\".*\",\"question\":\"To do or not to do\\?\",\"options\":\\[\"Do\",\"Not Do\"\\],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", val)
		return match
	})).Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()

	assert.Regexp(t, regexp.MustCompile("\\{\"response_type\":\"in_channel\",\"blocks\":.*,\"replace_original\":false}"), slackRequest)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestValidNewPollFailureToPersistPollInfo(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22%20%22Do%22%20%22Not%20Do%22&response_url=" + server.URL + "&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("PutSiloString", mock.Anything, "pollInfo", mock.MatchedBy(func(val string) bool {
		match, _ := regexp.MatchString("{\"id\":\".*\",\"question\":\"To do or not to do\\?\",\"options\":\\[\"Do\",\"Not Do\"\\],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", val)
		return match
	})).Return(fmt.Errorf("failed to persist"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()

	// Validate the immediate slack acknowledgement response is sent
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error persisting poll. Please try again.\",\"replace_original\":false}", slackRequest)
}

func TestInvalidNewPollPayload(t *testing.T) {
	body := "%gh&%ij"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Error decoding params from body [%gh&%ij]: invalid URL escape \"%gh\"\n", string(rbody))
	assert.Equal(t, 400, resp.StatusCode)
}

func TestInvalidNewVotePayload(t *testing.T) {
	body := "%gh&%ij"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Error decoding params from body [%gh&%ij]: invalid URL escape \"%gh\"\n", string(rbody))
	assert.Equal(t, 400, resp.StatusCode)
}

func TestNewPollWithWrongUsage(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	body := fmt.Sprintf("token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=marco&user_name=marcopoller&command=%%2Fpoll&text=%%22To%%20do%%20or%%20not%%20to%%20do%%3F%%22&response_url=%s&trigger_id=someTriggerID", server.URL)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Wrong usage. `/poll \\\"Question\\\" \\\"Option 1\\\" \\\"Option 2\\\" ...`\",\"replace_original\":false}", slackRequest)
}

func TestErrorReadingBodyOnNewPoll(t *testing.T) {
	reader := &Reader{}
	reader.On("Read", mock.Anything).Return(0, fmt.Errorf("IO Error"))
	defer reader.AssertExpectations(t)

	r := httptest.NewRequest(http.MethodPost, "/", reader)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "IO Error\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestErrorReadingBodyOnNewVote(t *testing.T) {
	reader := &Reader{}
	reader.On("Read", mock.Anything).Return(0, fmt.Errorf("IO Error"))
	defer reader.AssertExpectations(t)

	r := httptest.NewRequest(http.MethodPost, "/", reader)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "IO Error\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestNewPollWithInvalidSlackSignature(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22%20%22Do%22%20%22Not%20Do%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionSlackVerifier("badSecret"), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Error creating slack secrets verifier: timestamp is too old\n", string(rbody))
	assert.Equal(t, 403, resp.StatusCode)
}

func TestNewVoteWithInvalidSlackSignature(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22%20%22Do%22%20%22Not%20Do%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionSlackVerifier("badSecret"), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Error creating slack secrets verifier: timestamp is too old\n", string(rbody))
	assert.Equal(t, 403, resp.StatusCode)
}

func TestValidVoteUpdate(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	userFinder.On("GetUserInfo", "marco").Return(&slack.User{ID: "marco", Profile: slack.UserProfile{Image24: "http://image.me", RealName: "Marco Poller"}}, nil)
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", "marco": "0"}, nil)
	storer.On("PutSiloString", "1566576557-poll1", "marco", "1").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)

	assert.Regexp(t, regexp.MustCompile("\\{\"blocks\".*,\"replace_original\":true}"), slackRequest)
}

func TestVoteOnExpiredPoll(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1", ActionTs: "1566580158"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.ExpirationPollVerifier{ValidityPeriod: time.Duration(1) * time.Hour}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Sorry, the poll is expired and is now read-only\",\"replace_original\":false}", slackRequest)
}

func TestVoteOnPollUsingOldIdentifierFormat(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "poll1,vote", Value: "1", ActionTs: "1566580158"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.ExpirationPollVerifier{ValidityPeriod: time.Duration(1) * time.Hour}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Sorry, the poll is expired and is now read-only\",\"replace_original\":false}", slackRequest)
}

func TestErrorLoadingUserInfoOnVoteRegistration(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	userFinder.On("GetUserInfo", "marco").Return(nil, fmt.Errorf("can't get user info"))
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", nil)
	storer.On("PutSiloString", "1566576557-poll1", "marco", "1").Return(nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", "marco": "0"}, nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error listing votes for poll. Please try again.\",\"replace_original\":false}", slackRequest)
}

func TestValidNewVote(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"}, nil)
	storer.On("PutSiloString", "1566576557-poll1", "marco", "1").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)

	assert.Regexp(t, regexp.MustCompile("\\{\"blocks\".*,\"replace_original\":true}"), slackRequest)
}

func TestValidNewVoteFailureToLoadPoll(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("", fmt.Errorf("failed to load"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error getting existing poll info. Please try again.\",\"replace_original\":false}", slackRequest)
}

func TestNewVoteInvalidStoredPollInfo(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("corrupted", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error parsing existing poll info. Please report this issue at https://github.com/alexandre-normand/marcopoller\",\"replace_original\":false}", slackRequest)
}

func TestErrorStoringNewVote(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", nil)
	storer.On("PutSiloString", "1566576557-poll1", "marco", "1").Return(fmt.Errorf("failed to put new vote"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error persisting vote. Please try again.\",\"replace_original\":false}", slackRequest)
}

func TestErrorListingVotesOnNewVote(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", nil)
	storer.On("PutSiloString", "1566576557-poll1", "marco", "1").Return(nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{}, fmt.Errorf("failed to load votes"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error listing votes for poll. Please try again.\",\"replace_original\":false}", slackRequest)
}

func TestErrorUpdatingMessageOnNewVote(t *testing.T) {
	lastRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		lastRequest = string(reqBody)
		// Reject the ephemeral user message requests but let the others succeed
		if !strings.Contains(lastRequest, "ephemeral") {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("can't update message"))
		} else {
			fmt.Fprintln(w, "OK")
		}
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,vote", Value: "1"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"}, nil)
	storer.On("PutSiloString", "1566576557-poll1", "marco", "1").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error updating slack message for poll. Please try again.\",\"replace_original\":false}", lastRequest)
}

func TestNewVoteEmptyPayload(t *testing.T) {
	body := "payload="
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Empty payload\n", string(rbody))
	assert.Equal(t, 400, resp.StatusCode)
}

func TestNewVoteInvalidCallback(t *testing.T) {
	body := "payload=invalidCallback"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "invalid character 'i' looking for beginning of value\n", string(rbody))
	assert.Equal(t, 400, resp.StatusCode)
}

func TestDeletePoll(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,delete", Value: "delete"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"marco\"}", nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", "marco": "0"}, nil)
	storer.On("DeleteSiloString", "1566576557-poll1", "pollInfo").Return(nil)
	storer.On("DeleteSiloString", "1566576557-poll1", "marco").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, "{\"delete_original\":true}", slackRequest)
}

func TestDeletePollFailedToDeleteMsg(t *testing.T) {
	lastRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		lastRequest = string(reqBody)
		// Reject the ephemeral user message requests but let the others succeed
		if !strings.Contains(lastRequest, "ephemeral") {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("failed to delete"))
		} else {
			fmt.Fprintln(w, "OK")
		}
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,delete", Value: "delete"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"marco\"}", nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"}, nil)
	storer.On("DeleteSiloString", "1566576557-poll1", "pollInfo").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error deleting message from slack\",\"replace_original\":false}", lastRequest)
}

func TestDeletePollFailedToDeleteStoredData(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,delete", Value: "delete"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"marco\"}", nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"}, nil)
	storer.On("DeleteSiloString", "1566576557-poll1", "pollInfo").Return(fmt.Errorf("failed to delete data from db"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error deleting poll. Please try again\",\"replace_original\":false}", slackRequest)
}

func TestUnauthorizedDeletePoll(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "thatguy"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,delete", Value: "delete"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"marco\"}", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Only the poll creator (\\u003c@marco\\u003e) is allowed to delete the poll\",\"replace_original\":false}", slackRequest)
}

func TestUnauthorizedDeletePollFailureToSendSlackMsg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("failed to send"))
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "thatguy"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,delete", Value: "delete"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"marco\"}", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestNoVerifierOnCreation(t *testing.T) {
	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "Verifier is nil after applying all Options. Did you forget to set one?")
}

func TestNoUserFinderOnCreation(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "UserFinder is nil after applying all Options. Did you forget to set one?")
}

func TestNoStorerOnCreation(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}), marcopoller.OptionDialoguer(dialoguer))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "Storer is nil after applying all Options. Did you forget to set one?")
}

func TestNoPollVerifierOnCreation(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionStorer(storer), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionDialoguer(dialoguer))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "PollVerifier is nil after applying all Options. Did you forget to set one?")
}

func TestNoDialoguerOnCreation(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionStorer(storer), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "Dialoguer is nil after applying all Options. Did you forget to set one?")
}

// A New Slack Client with a bad token doesn't immediately fail on instantiation because it doesn't connect until
// a RTM is created or the API is used
func TestNewWithSlackClientWithBadToken(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionSlackUserFinder("badToken", false), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))

	assert.NotNil(t, mp)
	assert.NoError(t, err)
}

func TestNewWithDatastoreWithoutCredentialsAndInvalidProjectID(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionDatastore("invalidProjectID"), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.Contains(t, err.Error(), "Error initializing datastore persistence on project [invalidProjectID]")
}

func TestDeleteExpiredPolls(t *testing.T) {
	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	// Set up 2 expired polls and 1 still valid
	storer.On("GlobalScan").Return(map[string]map[string]string{"1566576557-expiredPoll1": {"pollInfo": "{\"id\":\"1566576557-expiredPoll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"},
		"1566574991-expiredPoll2": {"pollInfo": "{\"id\":\"1566574991-expiredPoll2\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"},
		"1566580148-freshPoll":    {"pollInfo": "{\"id\":\"1566580148-freshPoll\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"}}, nil)
	storer.On("ScanSilo", "1566576557-expiredPoll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-expiredPoll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"}, nil)
	storer.On("DeleteSiloString", "1566576557-expiredPoll1", "pollInfo").Return(nil)
	storer.On("ScanSilo", "1566574991-expiredPoll2").Return(map[string]string{"pollInfo": "{\"id\":\"1566574991-expiredPoll2\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}"}, nil)
	storer.On("DeleteSiloString", "1566574991-expiredPoll2", "pollInfo").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.ExpirationPollVerifier{ValidityPeriod: time.Duration(1) * time.Hour}))
	require.NoError(t, err)

	deleted, err := mp.DeleteExpiredPolls(time.Unix(1566580158, 0))
	require.NoError(t, err)

	assert.Equal(t, 2, deleted)
}

func TestInteractivePrompt(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	dialoguer.On("OpenView", "someTriggerID", mock.Anything).Return(nil, nil)
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestErrorCreatingInteractivePrompt(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=&response_url=" + server.URL + "&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	dialoguer.On("OpenView", "someTriggerID", mock.Anything).Return(nil, fmt.Errorf("error opening view"))
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error opening up interactive prompt. Try again, maybe?\",\"replace_original\":false}", slackRequest)
}

func TestErrorCreatingNewPoll(t *testing.T) {
	lastRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		lastRequest = string(reqBody)
		// Reject the ephemeral user message requests but let the others succeed
		if !strings.Contains(lastRequest, "ephemeral") {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("can't update message"))
		} else {
			fmt.Fprintln(w, "OK")
		}
	}))
	defer server.Close()

	body := fmt.Sprintf("token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%%2Fpoll&text=%%22To%%20do%%20or%%20not%%20to%%20do%%3F%%22%%20%%22Do%%22%%20%%22Not%%20Do%%22&response_url=%s&trigger_id=someTriggerID", server.URL)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("PutSiloString", mock.Anything, "pollInfo", mock.MatchedBy(func(val string) bool {
		match, _ := regexp.MatchString("{\"id\":\".*\",\"question\":\"To do or not to do\\?\",\"options\":\\[\"Do\",\"Not Do\"\\],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", val)
		return match
	})).Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Error writing new poll to slack. Please try again.\",\"replace_original\":false}", lastRequest)
}

func TestInteractivePollSubmission(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := marcopoller.InteractionCallback{Type: "view_submission",
		User:         slack.User{ID: "marco"},
		ResponseURLs: []marcopoller.ResponseURL{marcopoller.ResponseURL{ResponseURL: server.URL}},
		View: slack.View{CallbackID: "interactive-poll-create",
			State: &slack.ViewState{Values: map[string]map[string]slack.BlockAction{
				"poll_question":       map[string]slack.BlockAction{"poll_question": slack.BlockAction{Value: "To do or not to do?"}},
				"poll_answer_options": map[string]slack.BlockAction{"poll_answer_options": slack.BlockAction{Value: "Do\nNot Do\n"}},
			}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("PutSiloString", mock.Anything, "pollInfo", mock.MatchedBy(func(val string) bool {
		match, _ := regexp.MatchString("{\"id\":\".*\",\"question\":\"To do or not to do\\?\",\"options\":\\[\"Do\",\"Not Do\"\\],\"features\":{\"multianswers\":false},\"creator\":\"marco\"}", val)
		return match
	})).Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Regexp(t, regexp.MustCompile("\\{\"response_type\":\"in_channel\",\"blocks\":.*,\"replace_original\":false}"), slackRequest)
}

func TestInvalidCallbackType(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := marcopoller.InteractionCallback{Type: "invalidType", ResponseURL: server.URL}

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"response_type\":\"ephemeral\",\"text\":\":warning: Unknown interaction callback type: invalidType\",\"replace_original\":false}", slackRequest)
}

func TestCloseVoting(t *testing.T) {
	slackRequest := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		slackRequest = string(reqBody)
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	callback := slack.InteractionCallback{Type: "block_actions", User: slack.User{ID: "marco"}, ResponseURL: server.URL, ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{&slack.BlockAction{ActionID: "1566576557-poll1,close", Value: "close"}}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	userFinder := &UserFinder{}
	userFinder.On("GetUserInfo", "marco").Return(&slack.User{ID: "marco", Profile: slack.UserProfile{Image24: "http://image.me", RealName: "Marco Poller"}}, nil)
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "1566576557-poll1", "pollInfo").Return("{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"marco\"}", nil)
	storer.On("ScanSilo", "1566576557-poll1").Return(map[string]string{"pollInfo": "{\"id\":\"1566576557-poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"1566576557.354007\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"features\":{\"multianswers\":false},\"creator\":\"UID\"}", "marco": "0"}, nil)
	storer.On("DeleteSiloString", "1566576557-poll1", "pollInfo").Return(nil)
	storer.On("DeleteSiloString", "1566576557-poll1", "marco").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	dialoguer := &mmocks.Dialoguer{}
	defer dialoguer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer), marcopoller.OptionDialoguer(dialoguer), marcopoller.OptionPollVerifier(marcopoller.AlwaysValidPollVerifier{}))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.HandleInteractions(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, "{\"blocks\":[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*To do or not to do?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Do\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"http://image.me\",\"alt_text\":\"\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Not Do\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e (voting closed)\"}]}],\"replace_original\":true}", slackRequest)
}
