package marcopoller

import (
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderPollNoVotes(t *testing.T) {
	poll := Poll{ID: "un", MsgID: MsgIdentifier{ChannelID: "myLittleChannel", Timestamp: "1120"}, Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{})
	assert.Equal(t, []slack.Block{*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "*What's your favorite book?*", false, false), nil, nil),
		*slack.NewDividerBlock(),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "0", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Story of B", false, false), nil, *slack.NewButtonBlockElement("un", "1", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • My Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "2", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Paradise Built in Hell", false, false), nil, *slack.NewButtonBlockElement("un", "3", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " ", false, false), nil, *slack.NewButtonBlockElement("un", "delete", slack.NewTextBlockObject("plain_text", "Delete poll", false, false))),
		*slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "Created by <@marco>", false, false)),
	}, blocks)
}

func TestRenderPollOneVote(t *testing.T) {
	poll := Poll{ID: "un", MsgID: MsgIdentifier{ChannelID: "myLittleChannel", Timestamp: "1120"}, Question: "What's your favorite book?", Options: []string{"Ishmael", "Story of B", "My Ishmael", "Paradise Built in Hell"}, Creator: "marco"}
	blocks := renderPoll(poll, map[string][]Voter{"0": []Voter{Voter{userID: "marco", avatarURL: "https://avatar.me", name: "Marco Poller"}}})
	assert.Equal(t, []slack.Block{*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "*What's your favorite book?*", false, false), nil, nil),
		*slack.NewDividerBlock(),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "0", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewContextBlock("", *slack.NewImageBlockObject("https://avatar.me", "Marco Poller")),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Story of B", false, false), nil, *slack.NewButtonBlockElement("un", "1", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • My Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "2", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Paradise Built in Hell", false, false), nil, *slack.NewButtonBlockElement("un", "3", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " ", false, false), nil, *slack.NewButtonBlockElement("un", "delete", slack.NewTextBlockObject("plain_text", "Delete poll", false, false))),
		*slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "Created by <@marco>", false, false)),
	}, blocks)
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

	assert.Equal(t, []slack.Block{*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "*What's your favorite book?*", false, false), nil, nil),
		*slack.NewDividerBlock(),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "0", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewContextBlock("", *slack.NewImageBlockObject("https://avatar1.me", "User1"),
			*slack.NewImageBlockObject("https://avatar2.me", "User2"),
			*slack.NewImageBlockObject("https://avatar3.me", "User3"),
			*slack.NewImageBlockObject("https://avatar4.me", "User4"),
			*slack.NewImageBlockObject("https://avatar5.me", "User5"),
			*slack.NewImageBlockObject("https://avatar6.me", "User6"),
			*slack.NewImageBlockObject("https://avatar7.me", "User7"),
			*slack.NewImageBlockObject("https://avatar8.me", "User8"),
			*slack.NewImageBlockObject("https://avatar9.me", "User9"),
			*slack.NewTextBlockObject("mrkdwn", "`+ 2`", false, false),
		),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Story of B", false, false), nil, *slack.NewButtonBlockElement("un", "1", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • My Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "2", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Paradise Built in Hell", false, false), nil, *slack.NewButtonBlockElement("un", "3", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " ", false, false), nil, *slack.NewButtonBlockElement("un", "delete", slack.NewTextBlockObject("plain_text", "Delete poll", false, false))),
		*slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "Created by <@marco>", false, false)),
	}, blocks)
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

	assert.Equal(t, []slack.Block{*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "*What's your favorite book?*", false, false), nil, nil),
		*slack.NewDividerBlock(),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "0", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewContextBlock("", *slack.NewImageBlockObject("https://avatar1.me", "User1"),
			*slack.NewImageBlockObject("https://avatar2.me", "User2"),
			*slack.NewImageBlockObject("https://avatar3.me", "User3"),
			*slack.NewImageBlockObject("https://avatar4.me", "User4"),
			*slack.NewImageBlockObject("https://avatar5.me", "User5"),
			*slack.NewImageBlockObject("https://avatar6.me", "User6"),
			*slack.NewImageBlockObject("https://avatar7.me", "User7"),
			*slack.NewImageBlockObject("https://avatar8.me", "User8"),
			*slack.NewImageBlockObject("https://avatar9.me", "User9"),
			*slack.NewTextBlockObject("mrkdwn", "`+ 1`", false, false),
		),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Story of B", false, false), nil, *slack.NewButtonBlockElement("un", "1", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • My Ishmael", false, false), nil, *slack.NewButtonBlockElement("un", "2", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " • Paradise Built in Hell", false, false), nil, *slack.NewButtonBlockElement("un", "3", slack.NewTextBlockObject("plain_text", "Vote", false, false))),
		*slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", " ", false, false), nil, *slack.NewButtonBlockElement("un", "delete", slack.NewTextBlockObject("plain_text", "Delete poll", false, false))),
		*slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "Created by <@marco>", false, false)),
	}, blocks)
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
