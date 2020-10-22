package marcopoller

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderPollNoVotes(t *testing.T) {
	poll := Poll{ID: "un", Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{}, false)
	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"actions\",\"block_id\":\"un\",\"elements\":[{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Close voting\"},\"action_id\":\"un,close\",\"value\":\"close\"},{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un,delete\",\"value\":\"delete\",\"style\":\"danger\"}]},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
}

func TestRenderPollOneVote(t *testing.T) {
	poll := Poll{ID: "un", Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{"0": []Voter{Voter{userID: "marco", avatarURL: "https://avatar.me", name: "Marco Poller"}}}, false)
	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar.me\",\"alt_text\":\"Marco Poller\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"actions\",\"block_id\":\"un\",\"elements\":[{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Close voting\"},\"action_id\":\"un,close\",\"value\":\"close\"},{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un,delete\",\"value\":\"delete\",\"style\":\"danger\"}]},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
}

func TestRenderPollElevenVoters(t *testing.T) {
	poll := Poll{ID: "un", Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{"0": []Voter{Voter{userID: "user1", avatarURL: "https://avatar1.me", name: "User1"},
		Voter{userID: "user2", avatarURL: "https://avatar2.me", name: "User2"},
		Voter{userID: "user3", avatarURL: "https://avatar3.me", name: "User3"},
		Voter{userID: "user4", avatarURL: "https://avatar4.me", name: "User4"},
		Voter{userID: "user5", avatarURL: "https://avatar5.me", name: "User5"},
		Voter{userID: "user6", avatarURL: "https://avatar6.me", name: "User6"},
		Voter{userID: "user7", avatarURL: "https://avatar7.me", name: "User7"},
		Voter{userID: "user8", avatarURL: "https://avatar8.me", name: "User8"},
		Voter{userID: "user9", avatarURL: "https://avatar9.me", name: "User9"},
		Voter{userID: "user10", avatarURL: "https://avatar10.me", name: "User10"},
		Voter{userID: "user11", avatarURL: "https://avatar11.me", name: "User11"},
	}}, false)

	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar1.me\",\"alt_text\":\"User1\"},{\"type\":\"image\",\"image_url\":\"https://avatar2.me\",\"alt_text\":\"User2\"},{\"type\":\"image\",\"image_url\":\"https://avatar3.me\",\"alt_text\":\"User3\"},{\"type\":\"image\",\"image_url\":\"https://avatar4.me\",\"alt_text\":\"User4\"},{\"type\":\"image\",\"image_url\":\"https://avatar5.me\",\"alt_text\":\"User5\"},{\"type\":\"image\",\"image_url\":\"https://avatar6.me\",\"alt_text\":\"User6\"},{\"type\":\"image\",\"image_url\":\"https://avatar7.me\",\"alt_text\":\"User7\"},{\"type\":\"image\",\"image_url\":\"https://avatar8.me\",\"alt_text\":\"User8\"},{\"type\":\"image\",\"image_url\":\"https://avatar9.me\",\"alt_text\":\"User9\"},{\"type\":\"mrkdwn\",\"text\":\"`+ 2`\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"actions\",\"block_id\":\"un\",\"elements\":[{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Close voting\"},\"action_id\":\"un,close\",\"value\":\"close\"},{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un,delete\",\"value\":\"delete\",\"style\":\"danger\"}]},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
}

func TestRenderPollTenVoters(t *testing.T) {
	poll := Poll{ID: "un", Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{"0": []Voter{Voter{userID: "user1", avatarURL: "https://avatar1.me", name: "User1"},
		Voter{userID: "user2", avatarURL: "https://avatar2.me", name: "User2"},
		Voter{userID: "user3", avatarURL: "https://avatar3.me", name: "User3"},
		Voter{userID: "user4", avatarURL: "https://avatar4.me", name: "User4"},
		Voter{userID: "user5", avatarURL: "https://avatar5.me", name: "User5"},
		Voter{userID: "user6", avatarURL: "https://avatar6.me", name: "User6"},
		Voter{userID: "user7", avatarURL: "https://avatar7.me", name: "User7"},
		Voter{userID: "user8", avatarURL: "https://avatar8.me", name: "User8"},
		Voter{userID: "user9", avatarURL: "https://avatar9.me", name: "User9"},
		Voter{userID: "user10", avatarURL: "https://avatar10.me", name: "User10"},
	}}, false)

	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar1.me\",\"alt_text\":\"User1\"},{\"type\":\"image\",\"image_url\":\"https://avatar2.me\",\"alt_text\":\"User2\"},{\"type\":\"image\",\"image_url\":\"https://avatar3.me\",\"alt_text\":\"User3\"},{\"type\":\"image\",\"image_url\":\"https://avatar4.me\",\"alt_text\":\"User4\"},{\"type\":\"image\",\"image_url\":\"https://avatar5.me\",\"alt_text\":\"User5\"},{\"type\":\"image\",\"image_url\":\"https://avatar6.me\",\"alt_text\":\"User6\"},{\"type\":\"image\",\"image_url\":\"https://avatar7.me\",\"alt_text\":\"User7\"},{\"type\":\"image\",\"image_url\":\"https://avatar8.me\",\"alt_text\":\"User8\"},{\"type\":\"image\",\"image_url\":\"https://avatar9.me\",\"alt_text\":\"User9\"},{\"type\":\"mrkdwn\",\"text\":\"`+ 1`\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un,vote\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"actions\",\"block_id\":\"un\",\"elements\":[{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Close voting\"},\"action_id\":\"un,close\",\"value\":\"close\"},{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un,delete\",\"value\":\"delete\",\"style\":\"danger\"}]},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
}

func TestRenderClosedPoll(t *testing.T) {
	poll := Poll{ID: "un", Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{"0": []Voter{Voter{userID: "user1", avatarURL: "https://avatar1.me", name: "User1"},
		Voter{userID: "user2", avatarURL: "https://avatar2.me", name: "User2"},
		Voter{userID: "user3", avatarURL: "https://avatar3.me", name: "User3"},
		Voter{userID: "user4", avatarURL: "https://avatar4.me", name: "User4"},
		Voter{userID: "user5", avatarURL: "https://avatar5.me", name: "User5"},
		Voter{userID: "user6", avatarURL: "https://avatar6.me", name: "User6"},
		Voter{userID: "user7", avatarURL: "https://avatar7.me", name: "User7"},
		Voter{userID: "user8", avatarURL: "https://avatar8.me", name: "User8"},
		Voter{userID: "user9", avatarURL: "https://avatar9.me", name: "User9"},
		Voter{userID: "user10", avatarURL: "https://avatar10.me", name: "User10"},
		Voter{userID: "user11", avatarURL: "https://avatar11.me", name: "User11"},
	}}, true)

	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar1.me\",\"alt_text\":\"User1\"},{\"type\":\"image\",\"image_url\":\"https://avatar2.me\",\"alt_text\":\"User2\"},{\"type\":\"image\",\"image_url\":\"https://avatar3.me\",\"alt_text\":\"User3\"},{\"type\":\"image\",\"image_url\":\"https://avatar4.me\",\"alt_text\":\"User4\"},{\"type\":\"image\",\"image_url\":\"https://avatar5.me\",\"alt_text\":\"User5\"},{\"type\":\"image\",\"image_url\":\"https://avatar6.me\",\"alt_text\":\"User6\"},{\"type\":\"image\",\"image_url\":\"https://avatar7.me\",\"alt_text\":\"User7\"},{\"type\":\"image\",\"image_url\":\"https://avatar8.me\",\"alt_text\":\"User8\"},{\"type\":\"image\",\"image_url\":\"https://avatar9.me\",\"alt_text\":\"User9\"},{\"type\":\"mrkdwn\",\"text\":\"`+ 2`\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e (voting closed)\"}]}]", string(render))
}

func TestRenderClosedPollWithMultiVoting(t *testing.T) {
	poll := Poll{ID: "un", Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{"0": []Voter{Voter{userID: "user1", avatarURL: "https://avatar1.me", name: "User1"},
		Voter{userID: "user2", avatarURL: "https://avatar2.me", name: "User2"},
		Voter{userID: "user3", avatarURL: "https://avatar3.me", name: "User3"},
		Voter{userID: "user4", avatarURL: "https://avatar4.me", name: "User4"},
		Voter{userID: "user5", avatarURL: "https://avatar5.me", name: "User5"},
		Voter{userID: "user6", avatarURL: "https://avatar6.me", name: "User6"},
		Voter{userID: "user7", avatarURL: "https://avatar7.me", name: "User7"},
		Voter{userID: "user8", avatarURL: "https://avatar8.me", name: "User8"},
		Voter{userID: "user9", avatarURL: "https://avatar9.me", name: "User9"},
		Voter{userID: "user10", avatarURL: "https://avatar10.me", name: "User10"},
		Voter{userID: "user11", avatarURL: "https://avatar11.me", name: "User11"},
	},
		"1": []Voter{Voter{userID: "user1", avatarURL: "https://avatar1.me", name: "User1"},
			Voter{userID: "user2", avatarURL: "https://avatar2.me", name: "User2"},
			Voter{userID: "user3", avatarURL: "https://avatar3.me", name: "User3"},
			Voter{userID: "user4", avatarURL: "https://avatar4.me", name: "User4"},
		}}, true)

	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar1.me\",\"alt_text\":\"User1\"},{\"type\":\"image\",\"image_url\":\"https://avatar2.me\",\"alt_text\":\"User2\"},{\"type\":\"image\",\"image_url\":\"https://avatar3.me\",\"alt_text\":\"User3\"},{\"type\":\"image\",\"image_url\":\"https://avatar4.me\",\"alt_text\":\"User4\"},{\"type\":\"image\",\"image_url\":\"https://avatar5.me\",\"alt_text\":\"User5\"},{\"type\":\"image\",\"image_url\":\"https://avatar6.me\",\"alt_text\":\"User6\"},{\"type\":\"image\",\"image_url\":\"https://avatar7.me\",\"alt_text\":\"User7\"},{\"type\":\"image\",\"image_url\":\"https://avatar8.me\",\"alt_text\":\"User8\"},{\"type\":\"image\",\"image_url\":\"https://avatar9.me\",\"alt_text\":\"User9\"},{\"type\":\"mrkdwn\",\"text\":\"`+ 2`\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar1.me\",\"alt_text\":\"User1\"},{\"type\":\"image\",\"image_url\":\"https://avatar2.me\",\"alt_text\":\"User2\"},{\"type\":\"image\",\"image_url\":\"https://avatar3.me\",\"alt_text\":\"User3\"},{\"type\":\"image\",\"image_url\":\"https://avatar4.me\",\"alt_text\":\"User4\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e (voting closed)\"}]}]", string(render))
}

func TestParsePollParams(t *testing.T) {
	testCases := []struct {
		text     string
		question string
		options  []string
	}{
		{"\"Favorite thing?\" \"Reading\"", "Favorite thing?", []string{"Reading"}},
		{"\"Favorite thing?\" Reading Running", "Favorite thing?", []string{"Reading", "Running"}},
		{"Agree? Yes No", "Agree?", []string{"Yes", "No"}},
		{"Agree? Yes \"Don't care for trailing double-quotes", "Agree?", []string{"Yes", "Don't care for trailing double-quotes"}},
	}

	for _, tc := range testCases {
		_, question, options, err := parsePollParams(tc.text)
		require.NoError(t, err)

		assert.Equal(t, tc.question, question)
		assert.Equal(t, tc.options, options)
	}
}

func TestParsePollMissingParams(t *testing.T) {
	_, _, _, err := parsePollParams("\"Question but no options?\"")
	assert.EqualError(t, err, "Missing parameters in string [\"Question but no options?\"]")
}

func TestInteractivePollRequestRendering(t *testing.T) {
	viewRequest := createInteractivePollPrompt()

	render, err := json.Marshal(viewRequest)
	require.NoError(t, err)

	assert.Equal(t, "{\"type\":\"modal\",\"title\":{\"type\":\"plain_text\",\"text\":\"Marco Poller\"},\"blocks\":[{\"type\":\"input\",\"block_id\":\"poll_conversation_select\",\"label\":{\"type\":\"plain_text\",\"text\":\"Where do you want to send your poll?\"},\"element\":{\"type\":\"conversations_select\",\"action_id\":\"poll_conversation_select\",\"default_to_current_conversation\":true,\"response_url_enabled\":true}},{\"type\":\"input\",\"block_id\":\"poll_question\",\"label\":{\"type\":\"plain_text\",\"text\":\"What's your poll about?\"},\"element\":{\"type\":\"plain_text_input\",\"action_id\":\"poll_question\",\"placeholder\":{\"type\":\"plain_text\",\"text\":\"What's your favorite color?\"}}},{\"type\":\"input\",\"block_id\":\"poll_answer_options\",\"label\":{\"type\":\"plain_text\",\"text\":\"Answer Options\"},\"element\":{\"type\":\"plain_text_input\",\"action_id\":\"poll_answer_options\",\"placeholder\":{\"type\":\"plain_text\",\"text\":\"All the color options (one per line)\"},\"multiline\":true},\"hint\":{\"type\":\"plain_text\",\"text\":\"Enter the answer options (one per line)\"}},{\"type\":\"input\",\"block_id\":\"poll_features\",\"label\":{\"type\":\"plain_text\",\"text\":\"Options\"},\"element\":{\"type\":\"checkboxes\",\"action_id\":\"poll_features\",\"options\":[{\"text\":{\"type\":\"plain_text\",\"text\":\"Allow voters to vote for many options\"},\"value\":\"multivoting\"}]},\"optional\":true}],\"close\":{\"type\":\"plain_text\",\"text\":\"Cancel\"},\"submit\":{\"type\":\"plain_text\",\"text\":\"Create Poll\"},\"callback_id\":\"interactive-poll-create\"}", string(render))
}

func TestToggleVoteForValue(t *testing.T) {
	testCases := []struct {
		name           string
		existingVotes  string
		voteToToggle   string
		expectedOutput string
	}{
		{"First time vote", "", "2", "2"},
		{"Add second vote", "1", "2", "1,2"},
		{"Remove existing vote from single", "1", "1", ""},
		{"Remove existing vote from many", "1,2", "1", "2"},
		{"Toggle existing vote from many", "1,2,4", "2", "1,4"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := toggleVoteForValue(tc.existingVotes, tc.voteToToggle)
			assert.Equal(t, tc.expectedOutput, output)
		})
	}
}

func TestParseCallbackParsesResponseURLs(t *testing.T) {
	input := `
	{
		"type": "view_submission",
		"team": {
		  "id": "redacted",
		  "domain": "redacted"
		},
		"user": {
		  "id": "redacted",
		  "username": "redacted",
		  "name": "redacted",
		  "team_id": "redacted"
		},
		"api_app_id": "redacted",
		"token": "redacted",
		"trigger_id": "1301903545095.469250665127.6c5b73e004ccc34fd7c6b8e489796606",
		"view": {
		  "id": "redacted",
		  "team_id": "redacted",
		  "type": "modal",
		  "blocks": [
			{
			  "type": "input",
			  "block_id": "poll_conversation_select",
			  "label": {
				"type": "plain_text",
				"text": "Where do you want to send your poll?",
				"emoji": true
			  },
			  "optional": false,
			  "element": {
				"type": "conversations_select",
				"action_id": "poll_conversation_select",
				"default_to_current_conversation": true,
				"response_url_enabled": true,
				"initial_conversation": "redacted"
			  }
			},
			{
			  "type": "input",
			  "block_id": "poll_question",
			  "label": {
				"type": "plain_text",
				"text": "What's your poll about?",
				"emoji": true
			  },
			  "optional": false,
			  "element": {
				"type": "plain_text_input",
				"action_id": "poll_question",
				"placeholder": {
				  "type": "plain_text",
				  "text": "What's your favorite color?",
				  "emoji": true
				}
			  }
			},
			{
			  "type": "input",
			  "block_id": "poll_answer_options",
			  "label": {
				"type": "plain_text",
				"text": "Answer Options",
				"emoji": true
			  },
			  "hint": {
				"type": "plain_text",
				"text": "Enter the answer options (one per line)",
				"emoji": true
			  },
			  "optional": false,
			  "element": {
				"type": "plain_text_input",
				"action_id": "poll_answer_options",
				"placeholder": {
				  "type": "plain_text",
				  "text": "All the color options (one per line)",
				  "emoji": true
				},
				"multiline": true
			  }
			},
			{
			  "type": "input",
			  "block_id": "poll_features",
			  "label": {
				"type": "plain_text",
				"text": "Options",
				"emoji": true
			  },
			  "optional": false,
			  "element": {
				"type": "checkboxes",
				"action_id": "poll_features",
				"options": [
				  {
					"text": {
					  "type": "plain_text",
					  "text": "Allow voters to vote for many options",
					  "emoji": true
					},
					"value": "multivoting"
				  }
				]
			  }
			}
		  ],
		  "private_metadata": "",
		  "callback_id": "interactive-poll-create",
		  "state": {
			"values": {
			  "poll_conversation_select": {
				"poll_conversation_select": {
				  "type": "conversations_select",
				  "selected_conversation": "redacted"
				}
			  },
			  "poll_question": {
				"poll_question": {
				  "type": "plain_text_input",
				  "value": "What's your favorite?"
				}
			  },
			  "poll_answer_options": {
				"poll_answer_options": {
				  "type": "plain_text_input",
				  "value": "Blue\nRed"
				}
			  },
			  "poll_features": {
				"poll_features": {
				  "type": "checkboxes",
				  "selected_options": [
					{
					  "text": {
						"type": "plain_text",
						"text": "Allow voters to vote for many options",
						"emoji": true
					  },
					  "value": "multivoting"
					}
				  ]
				}
			  }
			}
		  },
		  "hash": "1598063908.QDG9s6ce",
		  "title": {
			"type": "plain_text",
			"text": "Marco Poller",
			"emoji": true
		  },
		  "clear_on_close": false,
		  "notify_on_close": false,
		  "close": {
			"type": "plain_text",
			"text": "Cancel",
			"emoji": true
		  },
		  "submit": {
			"type": "plain_text",
			"text": "Create Poll",
			"emoji": true
		  },
		  "previous_view_id": null,
		  "root_view_id": "redacted",
		  "app_id": "redacted",
		  "external_id": "",
		  "app_installed_team_id": "TDT7CKK3R",
		  "bot_id": "redacted"
		},
		"response_urls": [
		  {
			"block_id": "poll_conversation_select",
			"action_id": "poll_conversation_select",
			"channel_id": "redacted",
			"response_url": "https://hooks.slack.com/app/redacted/redacted/redacted"
		  }
		]
	  }	  
	`
	callback, err := parseCallback(input)
	assert.NoError(t, err)

	assert.Equal(t, "https://hooks.slack.com/app/redacted/redacted/redacted", callback.ResponseURLs[0].ResponseURL)
}

func TestParseCallbackParsesState(t *testing.T) {
	input := `
	{
  "type": "block_actions",
  "user": {
    "id": "redacted",
    "username": "redacted",
    "name": "redacted",
    "team_id": "redacted"
  },
  "api_app_id": "redacted",
  "token": "redacted",
  "container": {
    "type": "message",
    "message_ts": "1603396789.000300",
    "channel_id": "redacted",
    "is_ephemeral": false
  },
  "trigger_id": "1471323389664.469250665127.ee979b5671d1bcfdf3d73ec6c602879e",
  "team": {
    "id": "redacted",
    "domain": "redacted"
  },
  "channel": {
    "id": "redacted",
    "name": "privategroup"
  },
  "message": {
    "type": "message",
    "subtype": "bot_message",
    "text": "This content can't be displayed.",
    "ts": "1603396789.000300",
    "bot_id": "redacted",
    "blocks": [
      {
        "type": "section",
        "block_id": "iQ0a",
        "text": {
          "type": "mrkdwn",
          "text": "*Test*",
          "verbatim": false
        }
      },
      {
        "type": "divider",
        "block_id": "cnLo"
      },
      {
        "type": "section",
        "block_id": "eu/AJ",
        "text": {
          "type": "mrkdwn",
          "text": " • Yes",
          "verbatim": false
        },
        "accessory": {
          "type": "button",
          "action_id": "1603396789-8NUAND5diAybK5nm3QjHJM,vote",
          "style": "primary",
          "text": {
            "type": "plain_text",
            "text": "Vote",
            "emoji": true
          },
          "value": "0"
        }
      },
      {
        "type": "context",
        "block_id": "tVTg",
        "elements": [
          {
            "type": "image",
            "image_url": "redacted",
            "alt_text": "redacted"
          }
        ]
      },
      {
        "type": "section",
        "block_id": "vTXj+",
        "text": {
          "type": "mrkdwn",
          "text": " • No",
          "verbatim": false
        },
        "accessory": {
          "type": "button",
          "action_id": "1603396789-8NUAND5diAybK5nm3QjHJM,vote",
          "style": "primary",
          "text": {
            "type": "plain_text",
            "text": "Vote",
            "emoji": true
          },
          "value": "1"
        }
      },
      {
        "type": "actions",
        "block_id": "1603396789-8NUAND5diAybK5nm3QjHJM",
        "elements": [
          {
            "type": "button",
            "action_id": "1603396789-8NUAND5diAybK5nm3QjHJM,close",
            "text": {
              "type": "plain_text",
              "text": "Close voting",
              "emoji": true
            },
            "value": "close"
          },
          {
            "type": "button",
            "action_id": "1603396789-8NUAND5diAybK5nm3QjHJM,delete",
            "text": {
              "type": "plain_text",
              "text": "Delete poll",
              "emoji": true
            },
            "style": "danger",
            "value": "delete"
          }
        ]
      },
      {
        "type": "context",
        "block_id": "9s3uD",
        "elements": [
          {
            "type": "mrkdwn",
            "text": "Created by redactd",
            "verbatim": false
          }
        ]
      }
    ],
    "edited": {
      "user": "redacted",
      "ts": "1603396791.000000"
    }
  },
  "state": {
    "values": {}
  },
  "response_url": "redacted",
  "actions": [
    {
      "action_id": "1603396789-8NUAND5diAybK5nm3QjHJM,vote",
      "block_id": "eu/AJ",
      "text": {
        "type": "plain_text",
        "text": "Vote",
        "emoji": true
      },
      "value": "0",
      "style": "primary",
      "type": "button",
      "action_ts": "1603396798.112344"
    }
  ]
}`

	callback, err := parseCallback(input)
	assert.NoError(t, err)

	assert.NotNil(t, callback.State)
}
