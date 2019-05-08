package marcopoller

import (
	_ "github.com/GoogleCloudPlatform/berglas/pkg/auto"
	"log"
	"net/http"
	"os"
)

// Berglas secret environment variables
const (
	slackTokenEnv    = "SLACK_TOKEN"
	signingSecretEnv = "SIGNING_SECRET"
)

// StartPollBerglas handles a slash command request to start a new poll using berglas backed secrets
func StartPollBerglas(w http.ResponseWriter, r *http.Request) {
	mp, err := New(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), os.Getenv(GCPProjectIDEnv))
	if err != nil {
		log.Printf("Error initializing Marco Poller: %v", err)
		http.Error(w, err.Error(), 500)
	}

	mp.StartPoll(w, r)
}

// RegisterVoteBerglas handles a slash command request to register a poll vote using berglas backed secrets
func RegisterVoteBerglas(w http.ResponseWriter, r *http.Request) {
	mp, err := New(os.Getenv(slackTokenEnv), os.Getenv(signingSecretEnv), os.Getenv(GCPProjectIDEnv))
	if err != nil {
		log.Printf("Error initializing Marco Poller: %v", err)
		http.Error(w, err.Error(), 500)
	}

	mp.RegisterVote(w, r)
}
