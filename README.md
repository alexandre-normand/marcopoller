# Poller

## Deploy
`gcloud functions deploy startPoll --entry-point StartPoll --runtime go111 --trigger-http --project $PROJECT_ID --service-account ${SA_EMAIL} --set-env-vars "PROJECT_ID=${PROJECT_ID},SLACK_TOKEN=berglas://${BUCKET_ID}/slacktoken,SIGNING_SECRET=berglas://${BUCKET_ID}/signingsecret"`
`gcloud functions deploy registerVote --entry-point RegisterVote --runtime go111 --trigger-http --project $PROJECT_ID --service-account ${SA_EMAIL} --set-env-vars "PROJECT_ID=${PROJECT_ID},SLACK_TOKEN=berglas://${BUCKET_ID}/slacktoken,SIGNING_SECRET=berglas://${BUCKET_ID}/signingsecret"`

## Required slack app access
In order to deploy and run `Marco Poller`, you'll need to create a new `slack` app in your workspace with the following:
