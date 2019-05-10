package marcopoller_test

import (
	"encoding/json"
	"fmt"
	"github.com/alexandre-normand/marcopoller"
	"github.com/alexandre-normand/slackscot/store/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestValidNewPoll(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22%20%22Do%22%20%22Not%20Do%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("PostMessage", "CID", mock.Anything).Return("CID", "ts", nil)
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("PutSiloString", mock.Anything, "pollInfo", mock.MatchedBy(func(val string) bool {
		match, _ := regexp.MatchString("{\"id\":\".*\",\"msgID\":{\"channelID\":\"CID\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do\\?\",\"options\":\\[\"Do\",\"Not Do\"\\],\"creator\":\"UID\"}", val)
		return match
	})).Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestValidNewPollFailureToSendSlackMsg(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22%20%22Do%22%20%22Not%20Do%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("PostMessage", "CID", mock.Anything).Return("", "", fmt.Errorf("failed to send"))
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to send\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestValidNewPollFailureToPersistPollInfo(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22%20%22Do%22%20%22Not%20Do%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("PostMessage", "CID", mock.Anything).Return("CID", "ts", nil)
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("PutSiloString", mock.Anything, "pollInfo", mock.MatchedBy(func(val string) bool {
		match, _ := regexp.MatchString("{\"id\":\".*\",\"msgID\":{\"channelID\":\"CID\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do\\?\",\"options\":\\[\"Do\",\"Not Do\"\\],\"creator\":\"UID\"}", val)
		return match
	})).Return(fmt.Errorf("failed to persist"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to persist\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestInvalidNewPollPayload(t *testing.T) {
	body := "%gh&%ij"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
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

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Error decoding params from body [%gh&%ij]: invalid URL escape \"%gh\"\n", string(rbody))
	assert.Equal(t, 400, resp.StatusCode)
}

func TestNewPollWithWrongUsage(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=marco&user_name=marcopoller&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("PostEphemeral", "CID", "marco", mock.Anything).Return("", nil)
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNewPollWithWrongUsageFailureSendingMsgToSlack(t *testing.T) {
	body := "token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=marco&user_name=marcopoller&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("PostEphemeral", "CID", "marco", mock.Anything).Return("", fmt.Errorf("failed to connect"))
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to connect\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestErrorReadingBodyOnNewPoll(t *testing.T) {
	reader := &Reader{}
	reader.On("Read", mock.Anything).Return(0, fmt.Errorf("IO Error"))
	defer reader.AssertExpectations(t)

	r := httptest.NewRequest(http.MethodPost, "/", reader)

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
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

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

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

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionSlackVerifier("badSecret"), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
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

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionSlackVerifier("badSecret"), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Error creating slack secrets verifier: timestamp is too old\n", string(rbody))
	assert.Equal(t, 403, resp.StatusCode)
}

func TestValidVoteUpdate(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("UpdateMessage", "myLittleChannel", "ts", mock.Anything).Return("CID", "ts", "", nil)
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	userFinder.On("GetUserInfo", "marco").Return(&slack.User{ID: "marco", Profile: slack.UserProfile{Image24: "http://image.me", RealName: "Marco Poller"}}, nil)
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", nil)
	storer.On("ScanSilo", "poll1").Return(map[string]string{"pollInfo": "{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", "marco": "0"}, nil)
	storer.On("PutSiloString", "poll1", "marco", "1").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestErrorLoadingUserInfoOnVoteRegistration(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	userFinder.On("GetUserInfo", "marco").Return(nil, fmt.Errorf("can't get user info"))
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", nil)
	storer.On("PutSiloString", "poll1", "marco", "1").Return(nil)
	storer.On("ScanSilo", "poll1").Return(map[string]string{"pollInfo": "{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", "marco": "0"}, nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "can't get user info\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestValidNewVote(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("UpdateMessage", "myLittleChannel", "ts", mock.Anything).Return("CID", "ts", "", nil)
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", nil)
	storer.On("ScanSilo", "poll1").Return(map[string]string{"pollInfo": "{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}"}, nil)
	storer.On("PutSiloString", "poll1", "marco", "1").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestValidNewVoteFailureToLoadPoll(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("", fmt.Errorf("failed to load"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to load\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestNewVoteInvalidStoredPollInfo(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("corrupted", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "invalid character 'c' looking for beginning of value\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestErrorStoringNewVote(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", nil)
	storer.On("PutSiloString", "poll1", "marco", "1").Return(fmt.Errorf("failed to put new vote"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to put new vote\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestErrorListingVotesOnNewVote(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", nil)
	storer.On("PutSiloString", "poll1", "marco", "1").Return(nil)
	storer.On("ScanSilo", "poll1").Return(map[string]string{}, fmt.Errorf("failed to load votes"))
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to load votes\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestErrorUpdatingMessageOnNewVote(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "1"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("UpdateMessage", "myLittleChannel", "ts", mock.Anything).Return("", "", "", fmt.Errorf("can't update message"))
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", nil)
	storer.On("ScanSilo", "poll1").Return(map[string]string{"pollInfo": "{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}"}, nil)
	storer.On("PutSiloString", "poll1", "marco", "1").Return(nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "can't update message\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestNewVoteEmptyPayload(t *testing.T) {
	body := "payload="
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

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

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "invalid character 'i' looking for beginning of value\n", string(rbody))
	assert.Equal(t, 400, resp.StatusCode)
}

func TestDeletePoll(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "delete"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("DeleteMessage", "myLittleChannel", "ts").Return("CID", "ts", nil)
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"marco\"}", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDeletePollFailedToDelete(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "marco"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "delete"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("DeleteMessage", "myLittleChannel", "ts").Return("", "", fmt.Errorf("failed to delete"))
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"marco\"}", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to delete\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestUnauthorizedDeletePoll(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "thatguy"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "delete"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("PostEphemeral", "myLittleChannel", "thatguy", mock.Anything).Return("CID", nil)
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"marco\"}", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "", string(rbody))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUnauthorizedDeletePollFailureToSendSlackMsg(t *testing.T) {
	callback := marcopoller.Callback{User: slack.User{ID: "thatguy"}, Actions: []marcopoller.Action{marcopoller.Action{ActionID: "poll1", Value: "delete"}}}
	callback.Channel.ID = "myLittleChannel"

	payload, _ := json.Marshal(callback)
	body := fmt.Sprintf("payload=%s", payload)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Add("X-Slack-Signature", "8e9fe980e2b36c7a7accab28bd8e315667cf9122c3f01c3b7230bb9587627ccb")
	r.Header.Add("X-Slack-Request-Timestamp", "1531431954")

	messenger := &Messenger{}
	messenger.On("PostEphemeral", "myLittleChannel", "thatguy", mock.Anything).Return("CID", fmt.Errorf("failed to send"))
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	storer.On("GetSiloString", "poll1", "pollInfo").Return("{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"marco\"}", nil)
	defer storer.AssertExpectations(t)

	verifier := &Verifier{}
	verifier.On("Verify", r.Header, []byte(body)).Return(nil)
	defer verifier.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	mp.RegisterVote(w, r)

	resp := w.Result()
	rbody, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "failed to send\n", string(rbody))
	assert.Equal(t, 500, resp.StatusCode)
}

func TestNoVerifierOnCreation(t *testing.T) {
	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "Verifier is nil after applying all Options. Did you forget to set one?")
}

func TestNoMessengerOnCreation(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "Messenger is nil after applying all Options. Did you forget to set one?")
}

func TestNoUserFinderOnCreation(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionStorer(storer))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "UserFinder is nil after applying all Options. Did you forget to set one?")
}

func TestNoStorerOnCreation(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.EqualError(t, err, "Storer is nil after applying all Options. Did you forget to set one?")
}

// A New Slack Client with a bad token doesn't immediately fail on instantiation because it doesn't connect until
// a RTM is created or the API is used
func TestNewWithSlackClientWithBadToken(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionSlackClient("badToken", false), marcopoller.OptionStorer(storer))

	assert.NotNil(t, mp)
	assert.NoError(t, err)
}

func TestNewWithDatastoreWithoutCredentialsAndInvalidProjectID(t *testing.T) {
	verifier := &Verifier{}
	defer verifier.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions(marcopoller.OptionVerifier(verifier), marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionDatastore("invalidProjectID"))
	require.Error(t, err)

	assert.Nil(t, mp)
	assert.Contains(t, err.Error(), "Error initializing datastore persistence on project [invalidProjectID]")
}
