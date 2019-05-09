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
	storer.On("ScanSilo", "poll1").Return(map[string]string{"pollInfo": "{\"id\":\"poll1\",\"msgID\":{\"channelID\":\"myLittleChannel\",\"timestamp\":\"ts\"},\"question\":\"To do or not to do?\",\"options\":[\"Do\",\"Not Do\"],\"creator\":\"UID\"}", "marco": "{\"userID\":\"marco\",\"avatarURL\":\"http://image.me\",\"name\":\"Marco Poller\"}"}, nil)
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

func TestValidNewUpdate(t *testing.T) {
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
