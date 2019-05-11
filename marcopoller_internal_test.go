package marcopoller

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderPollNoVotes(t *testing.T) {
	poll := Poll{ID: "un", MsgID: MsgIdentifier{ChannelID: "myLittleChannel", Timestamp: "1120"}, Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{})
	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" \"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un\",\"value\":\"delete\",\"style\":\"danger\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
}

func TestRenderPollOneVote(t *testing.T) {
	poll := Poll{ID: "un", MsgID: MsgIdentifier{ChannelID: "myLittleChannel", Timestamp: "1120"}, Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{"0": []Voter{Voter{userID: "marco", avatarURL: "https://avatar.me", name: "Marco Poller"}}})
	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar.me\",\"alt_text\":\"Marco Poller\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" \"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un\",\"value\":\"delete\",\"style\":\"danger\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
}

func TestRenderPollElevenVoters(t *testing.T) {
	poll := Poll{ID: "un", MsgID: MsgIdentifier{ChannelID: "myLittleChannel", Timestamp: "1120"}, Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
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
	}})

	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar1.me\",\"alt_text\":\"User1\"},{\"type\":\"image\",\"image_url\":\"https://avatar2.me\",\"alt_text\":\"User2\"},{\"type\":\"image\",\"image_url\":\"https://avatar3.me\",\"alt_text\":\"User3\"},{\"type\":\"image\",\"image_url\":\"https://avatar4.me\",\"alt_text\":\"User4\"},{\"type\":\"image\",\"image_url\":\"https://avatar5.me\",\"alt_text\":\"User5\"},{\"type\":\"image\",\"image_url\":\"https://avatar6.me\",\"alt_text\":\"User6\"},{\"type\":\"image\",\"image_url\":\"https://avatar7.me\",\"alt_text\":\"User7\"},{\"type\":\"image\",\"image_url\":\"https://avatar8.me\",\"alt_text\":\"User8\"},{\"type\":\"image\",\"image_url\":\"https://avatar9.me\",\"alt_text\":\"User9\"},{\"type\":\"mrkdwn\",\"text\":\"`+ 2`\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" \"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un\",\"value\":\"delete\",\"style\":\"danger\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
}

func TestRenderPollTenVoters(t *testing.T) {
	poll := Poll{ID: "un", MsgID: MsgIdentifier{ChannelID: "myLittleChannel", Timestamp: "1120"}, Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
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
	}})

	render, err := json.Marshal(blocks)
	require.NoError(t, err)

	assert.Equal(t, "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"*What's your favorite book?*\"}},{\"type\":\"divider\"},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"0\",\"style\":\"primary\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"image\",\"image_url\":\"https://avatar1.me\",\"alt_text\":\"User1\"},{\"type\":\"image\",\"image_url\":\"https://avatar2.me\",\"alt_text\":\"User2\"},{\"type\":\"image\",\"image_url\":\"https://avatar3.me\",\"alt_text\":\"User3\"},{\"type\":\"image\",\"image_url\":\"https://avatar4.me\",\"alt_text\":\"User4\"},{\"type\":\"image\",\"image_url\":\"https://avatar5.me\",\"alt_text\":\"User5\"},{\"type\":\"image\",\"image_url\":\"https://avatar6.me\",\"alt_text\":\"User6\"},{\"type\":\"image\",\"image_url\":\"https://avatar7.me\",\"alt_text\":\"User7\"},{\"type\":\"image\",\"image_url\":\"https://avatar8.me\",\"alt_text\":\"User8\"},{\"type\":\"image\",\"image_url\":\"https://avatar9.me\",\"alt_text\":\"User9\"},{\"type\":\"mrkdwn\",\"text\":\"`+ 1`\"}]},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Story of B\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"1\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • My Ishmael\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"2\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" • Paradise Built in Hell\"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Vote\"},\"action_id\":\"un\",\"value\":\"3\",\"style\":\"primary\"}},{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\" \"},\"accessory\":{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Delete poll\"},\"action_id\":\"un\",\"value\":\"delete\",\"style\":\"danger\"}},{\"type\":\"context\",\"elements\":[{\"type\":\"mrkdwn\",\"text\":\"Created by \\u003c@marco\\u003e\"}]}]", string(render))
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
		question, options, err := parsePollParams(tc.text)
		require.NoError(t, err)

		assert.Equal(t, tc.question, question)
		assert.Equal(t, tc.options, options)
	}
}

func TestParsePollMissingParams(t *testing.T) {
	_, _, err := parsePollParams("")
	assert.EqualError(t, err, "Missing parameters in string []")

	_, _, err = parsePollParams("\"Question but no options?\"")
	assert.EqualError(t, err, "Missing parameters in string [\"Question but no options?\"]")
}
