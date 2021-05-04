# rest-api

![Gitlab pipeline status (self-hosted)](https://img.shields.io/gitlab/pipeline/monetr/rest-api/main?gitlab_url=https%3A%2F%2Fgitlab.elliotcourant.dev%2Fgithub.com&logo=gitlab)
[![DeepSource](https://deepsource.io/gh/monetr/rest-api.svg/?label=active+issues&show_trend=true&token=4x9L6ApemrQ6x80icvE9cEJl)](https://deepsource.io/gh/monetr/rest-api/?ref=repository-badge)
[![DeepSource](https://deepsource.io/gh/monetr/rest-api.svg/?label=resolved+issues&show_trend=true&token=4x9L6ApemrQ6x80icvE9cEJl)](https://deepsource.io/gh/monetr/rest-api/?ref=repository-badge)

<!-- Test change fasddfasdf -->

This is the REST API behind monetr's budgeting application.

API documentation can be found here: https://docs.monetr.dev/

Documentation is automatically generated with each commit to the main branch.

## Developing Locally

This is still a work in progress, but the entire REST API stack can be run locally and entirely in minikube.
This has been mostly automated and can be initiated by running the following command:

```bash
make init-mini
```

You will be prompted once for your password in a window; this is to add the certificate as trusted to macOS.

This will allow you to access the REST API via `https://api.monetr.mini`.

**NOTE:** This is still being tuned and is subject to change significantly. This also requires the following
to be installed on your computer.
- minikube
- docker
- hyperkit
- kubectl
- openssl (not the LibreSSL version)

To my knowledge this also only currently works on macOS. But I plan on making it work on Linux in the future
as well.
There are some other tools that this make target relies on, but if they are not already present on your computer
then they will be downloaded and installed into you `$API_PROJECT_FOLDER/bin` directory. 

Running init-mini will also create a certificate to be used for TLS locally. On macOS this is added to your
keychain as a trusted certificate so that way TLS will work locally as if it were in a real environment. This
and anything else that is done as part of running `init-mini` can be un-done by running the following command:

```bash
make clean-mini
```

This will remove all of the generated files, trusted certificates, and dependencies that were pulled for running
locally. It will also delete your minikube cluster to make sure the environment is completely clean.

### Seeing Your Changes

As you make changes to the code you can deploy your changes to the local minikube cluster by running:

```bash
make deploy-mini-application
```

This will build a new docker image from your current code as well as evaluate any changes made to the `values.local.yaml`
file in your project directory and push those changes to minikube.

If you want to do step debugging there is a shortcut in the makefile to do so. You will need to create a `config.yaml`
file in your project directory with all the same settings you have specified in your `values.local.yaml` for everything
to work properly. Once you have done that you can run the command below:

```bash
make local-api
```

This will spawn a new tmux window with a minikube tunnel in it. This is needed for the API to talk to the services it
needs within the minikube cluster.
This will also swap out the target for the API service in kubernetes with an endpoint pointed at your computer and at port
4000. This way you can run a REST API instance locally through something like GoLand or VSCode and step debug API
requests directly.

### Testing Webhooks

The API is meant to receive webhooks from both Plaid and Stripe as part of its normal functionality. To help make
development for these APIs easier I have added an ngrok deployment that will allow external services to send 
requests to your local REST API instance running in minikube.

To enable this for local development run:

```bash
make webhooks-mini NGROK_AUTH=${YOUR_NGROK_TOKEN}
```

This will deploy ngrok locally within the minikube cluster and will forward traffic to the REST API service.
You can see the ngrok inspector at `https://ngrok.monetr.mini` after this command has been run.

To disable webhooks if you have enabled them:

```bash
make disable-webhooks-mini
```

This will update the `values.local.yaml` in your project directory and disable the webhook settings, it will then
redeploy the REST API to minikube using the updated values file. Then it will delete the ngrok service that was
previously deployed.
