[![Build Status](https://travis-ci.org/alexandre-normand/marcopoller.svg?branch=master)](https://travis-ci.org/alexandre-normand/marcopoller)

# Marco Poller

## Requirements

*   A bot slack token 
    *   With the following `scopes`:
    	*   `chat:write:bot`
    	*   `chat:write:user`
    	*   `bot`
    	*   `commands`
    	*   `users.profile:read`

	*   The following `slash` commands:
    	*   poll

    *   The following `interactive` components:
    	*   poll

    *	A `bot user` with the recommended info:
		*   *Display name*: `Marco Poller`
		*   *Display username*: `marcopoller`

*   A gcloud project ID with the datastore API enabled
*   A way to store secrets (needed for the slack token and the slack signing secret)

## Integration

## Deploy
`gcloud functions deploy startPoll --entry-point StartPoll --runtime go111 --trigger-http --project $PROJECT_ID --service-account ${SA_EMAIL} --set-env-vars "PROJECT_ID=${PROJECT_ID},SLACK_TOKEN=berglas://${BUCKET_ID}/slacktoken,SIGNING_SECRET=berglas://${BUCKET_ID}/signingsecret"`
`gcloud functions deploy registerVote --entry-point RegisterVote --runtime go111 --trigger-http --project $PROJECT_ID --service-account ${SA_EMAIL} --set-env-vars "PROJECT_ID=${PROJECT_ID},SLACK_TOKEN=berglas://${BUCKET_ID}/slacktoken,SIGNING_SECRET=berglas://${BUCKET_ID}/signingsecret"`
