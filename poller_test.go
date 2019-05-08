package marcopoller_test

import (
	"github.com/alexandre-normand/marcopoller"
	"github.com/alexandre-normand/slackscot/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidNewPoll(t *testing.T) {
	messenger := &Messenger{}
	defer messenger.AssertExpectations(t)

	userFinder := &UserFinder{}
	defer userFinder.AssertExpectations(t)

	storer := &mocks.Storer{}
	defer storer.AssertExpectations(t)

	mp, err := marcopoller.NewWithOptions("myLittleSecret", marcopoller.OptionMessenger(messenger), marcopoller.OptionUserFinder(userFinder), marcopoller.OptionStorer(storer))
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("token=sometoken&team_id=TEAMID3&team_domain=test-workspace&channel_id=CID&channel_name=testchannel&user_id=UID&user_name=marco&command=%2Fpoll&text=%22To%20do%20or%20not%20to%20do%3F%22%20%22Do%22%20%22Not%20Do%22&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2Fbla%2Fbleh%2Fblo&trigger_id=someTriggerID"))
	w := httptest.NewRecorder()

	mp.StartPoll(w, r)

	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode)
}
