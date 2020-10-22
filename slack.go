package marcopoller

import (
	"encoding/json"
	"github.com/slack-go/slack"
)

// InteractionCallback embeds the full slack interaction callback and adds response urls
type InteractionCallback struct {
	Type            slack.InteractionType `json:"type"`
	Token           string                `json:"token"`
	CallbackID      string                `json:"callback_id"`
	ResponseURL     string                `json:"response_url"`
	TriggerID       string                `json:"trigger_id"`
	ActionTs        string                `json:"action_ts"`
	Team            slack.Team            `json:"team"`
	Channel         slack.Channel         `json:"channel"`
	User            slack.User            `json:"user"`
	OriginalMessage slack.Message         `json:"original_message"`
	Message         slack.Message         `json:"message"`
	Name            string                `json:"name"`
	Value           string                `json:"value"`
	MessageTs       string                `json:"message_ts"`
	AttachmentID    string                `json:"attachment_id"`
	ActionCallback  slack.ActionCallbacks `json:"actions"`
	View            slack.View            `json:"view"`
	ActionID        string                `json:"action_id"`
	APIAppID        string                `json:"api_app_id"`
	BlockID         string                `json:"block_id"`
	Container       slack.Container       `json:"container"`
	ResponseURLs    []ResponseURL         `json:"response_urls"`
	State           json.RawMessage       `json:"state,omitempty"`
	slack.DialogSubmissionCallback
	slack.ViewSubmissionCallback
	slack.ViewClosedCallback
}

// ResponseURL represents a response URL with associated data
type ResponseURL struct {
	BlockID     string `json:"block_id"`
	ActionID    string `json:"action_id"`
	ChannelID   string `json:"channel_id"`
	ResponseURL string `json:"response_url"`
}
