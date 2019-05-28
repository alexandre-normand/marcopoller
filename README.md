[![GoDoc](https://godoc.org/github.com/alexandre-normand/marcopoller?status.svg)](https://godoc.org/github.com/alexandre-normand/marcopoller)
[![Build Status](https://travis-ci.org/alexandre-normand/marcopoller.svg?branch=master)](https://travis-ci.org/alexandre-normand/marcopoller)
[![Test Coverage](https://api.codeclimate.com/v1/badges/8659e71d13e1df4118a2/test_coverage)](https://codeclimate.com/github/alexandre-normand/marcopoller/test_coverage)
[![Latest Version](https://img.shields.io/github/tag/alexandre-normand/marcopoller.svg?label=version)](https://github.com/alexandre-normand/marcopoller/releases)

# Marco Poller
![Marco Poller](https://www.dropbox.com/s/ddocj6175iofy2a/anchorman2.jpg?raw=1)

Marco Poller is a [slack](https://slack.com) app that does one thing: manage slack polls. Yes, there are many slack polling applications at the moment. Chances are
that you don't need this. But if you're part of an organisation that doesn't allow 3rd party slack apps or if you just feel like having your own version that does its
thing a little differently, go ahead and look at Marco Poller.

## Backing Technology
*   [Go](https://golang.org): `>= 1.12` 
*   [Google Cloud Functions](https://cloud.google.com/functions/docs/quickstart#functions-deploy-command-go)
*   [Google Cloud Datastore](https://cloud.google.com/datastore/docs/)

## Requirements

*   A bot slack token 
    *   With the following `scopes`:
    	*   `chat:write:bot`
    	*   `chat:write:user`
    	*   `bot`
    	*   `commands`
    	*   `users.profile:read`

	*   The following `slash` commands:
    	*   `/poll`:
    	    * *Command*: `/poll`
    	    * *Request URL*: `<url of the startPoll gcloud function>`
    	    * *Short Description*: `Starts a new poll`
    	    * *Usage Hint*: `"Question?" "Option1" "Option 2"`

    *   The following `interactive` components (should be toggled to `on`):
    	*   `registerVote` action URL: This is going to show up in the `gcloud functions deploy` output for the `registerVote` function. You only have to do this when
    	     you first deploy the `registerVote` function but you'll have to enter the `URL` of the `registerVote` gcloud function in _Request URL_. 

    *	A `bot user` with the recommended info:
		*   *Display name*: `Marco Poller`
		*   *Display username*: `marcopoller`

*   A gcloud project ID with the datastore API enabled
*   A way to store secrets (needed for the slack token and the slack signing secret)

## Integration
The ready-to-deploy vanilla version of `Marco Poller` uses [berglas](https://github.com/GoogleCloudPlatform/berglas) to manage secrets and lives at [github.com/alexandre-normand/marcopoller-vanilla](github.com/alexandre-normand/marcopoller-vanilla). 
Refer to the [berglas gcloud functions example and documentation](https://github.com/GoogleCloudPlatform/berglas/tree/master/examples/cloudfunctions/go) on how to set 
up berglas with gcloud functions. 

## Deploy
### Vanilla deployment using berglas managed secrets

* Make sure you've written the `slacktoken` and slack `signingsecret` using `berglas` and that you've granted the service account access to those. If you haven't done so already, 
refer to the [berglas gcloud functions example and documentation](https://github.com/GoogleCloudPlatform/berglas/tree/master/examples/cloudfunctions/go) for how to do this.

* Deploy the vanilla/berglas version using the [gcloud cli](https://cloud.google.com/sdk/gcloud/) commands from [github.com/alexandre-normand/marcopoller/berglas](berglas):

```
gcloud functions deploy startPoll --entry-point StartPoll --runtime go111 --trigger-http --project $PROJECT_ID --service-account ${SA_EMAIL} --set-env-vars "PROJECT_ID=${PROJECT_ID},SLACK_TOKEN=berglas://${BUCKET_ID}/slacktoken,SIGNING_SECRET=berglas://${BUCKET_ID}/signingsecret"
```

```
gcloud functions deploy registerVote --entry-point RegisterVote --runtime go111 --trigger-http --project $PROJECT_ID --service-account ${SA_EMAIL} --set-env-vars "PROJECT_ID=${PROJECT_ID},SLACK_TOKEN=berglas://${BUCKET_ID}/slacktoken,SIGNING_SECRET=berglas://${BUCKET_ID}/signingsecret"
```