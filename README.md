# rest-api

![Gitlab pipeline status (self-hosted)](https://img.shields.io/gitlab/pipeline/monetr/rest-api/main?gitlab_url=https%3A%2F%2Fgitlab.elliotcourant.dev%2Fgithub.com&logo=gitlab)
[![DeepSource](https://deepsource.io/gh/monetr/rest-api.svg/?label=active+issues&show_trend=true&token=4x9L6ApemrQ6x80icvE9cEJl)](https://deepsource.io/gh/monetr/rest-api/?ref=repository-badge)
[![DeepSource](https://deepsource.io/gh/monetr/rest-api.svg/?label=resolved+issues&show_trend=true&token=4x9L6ApemrQ6x80icvE9cEJl)](https://deepsource.io/gh/monetr/rest-api/?ref=repository-badge)

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

If you want to do step debugging you can use a tool called telepresence. The shorthand make target below will spawn a new
tmux session with both `minikube tunnel` and `telepresence` running side by side. This is because on their own I've found
that telepresence does not always tunnel traffic properly. And minikube has no way to send traffic from its cluster to
a local target. This will do both.

```bash
make debug-api-mini
```

Which will let you run the REST API on your actual computer and serve requests from minikube there. This means you can
run the REST API in GoLand or VSCode and use a step debugger to debug requests. There is a caveat with doing this though;
configuration is primarily provided to the REST API service in Kubernetes via environment variables. So you may need to
make a `config.yaml` file in your project directory and add any configuration options there in order to run the service
properly. More documentation on that file will be added at a later date.

You'll need to install macFUSE and SSHFS from the installer packages here: https://osxfuse.github.io/


You may see an error like the following in the tmux session that is started:
```
T: Warning: kubectl 1.21.0 may not work correctly with cluster version 1.18.15 due to the version discrepancy. See https://kubernetes.io/docs/setup/version-skew-policy/ for more information.

T: Using a Pod instead of a Deployment for the Telepresence proxy. If you experience problems, please file an issue!
T: Set the environment variable TELEPRESENCE_USE_DEPLOYMENT to any non-empty value to force the old behavior, e.g.,
T:     env TELEPRESENCE_USE_DEPLOYMENT=1 telepresence --run curl hello

T: Starting proxy with method 'inject-tcp', which has the following limitations: Go programs, static binaries, suid programs, and custom DNS implementations are not supported. For a full list of method limitations see
T: https://telepresence.io/reference/methods.html
T: Volumes are rooted at $TELEPRESENCE_ROOT. See https://telepresence.io/howto/volumes.html for details.
T: Starting network proxy to cluster by swapping out Deployment rest-api with a proxy Pod
T: Forwarding remote port 9000 to local port 9000.
T: Forwarding remote port 4000 to local port 4000.


Looks like there's a bug in our code. Sorry about that!

Traceback (most recent call last):
  File "/Users/elliotcourant/monetr/rest-api/bin/telepresence/telepresence/cli.py", line 135, in crash_reporting
    yield
  File "/Users/elliotcourant/monetr/rest-api/bin/telepresence/telepresence/main.py", line 81, in main
    user_process = launch(
  File "/Users/elliotcourant/monetr/rest-api/bin/telepresence/telepresence/outbound/setup.py", line 64, in launch
    return launch_inject(runner_, command, socks_port, env)
  File "/Users/elliotcourant/monetr/rest-api/bin/telepresence/telepresence/outbound/local.py", line 120, in launch_inject
    torsocks_env = set_up_torsocks(runner, socks_port)
  File "/Users/elliotcourant/monetr/rest-api/bin/telepresence/telepresence/outbound/local.py", line 71, in set_up_torsocks
    raise RuntimeError("SOCKS network proxying failed to start...")
RuntimeError: SOCKS network proxying failed to start...


Here are the last few lines of the logfile (see /Users/elliotcourant/monetr/rest-api/telepresence.log for the complete logs):

  17.8 TEL | [114] exit -11 in 0.05 secs.
  17.9 TEL | [115] Running: torsocks python3 -c 'import socket; socket.socket().connect(('"'"'kubernetes.default'"'"', 443))'
  17.9 TEL | [115] exit -11 in 0.05 secs.
  18.0 TEL | [116] Running: torsocks python3 -c 'import socket; socket.socket().connect(('"'"'kubernetes.default'"'"', 443))'
  18.0 TEL | [116] exit -11 in 0.05 secs.
  18.2 TEL | [117] Running: torsocks python3 -c 'import socket; socket.socket().connect(('"'"'kubernetes.default'"'"', 443))'
  18.2 TEL | [117] exit -11 in 0.05 secs.
  18.3 TEL | [118] Running: torsocks python3 -c 'import socket; socket.socket().connect(('"'"'kubernetes.default'"'"', 443))'
  18.3 TEL | [118] exit -11 in 0.05 secs.
  18.5 TEL | [119] Running: torsocks python3 -c 'import socket; socket.socket().connect(('"'"'kubernetes.default'"'"', 443))'
  18.5 TEL | [119] exit -11 in 0.04 secs.
  18.5 TEL | END SPAN local.py:42(set_up_torsocks)   15.0s

Would you like to file an issue in our issue tracker? You'll be able to review and edit before anything is posted to the public. We'd really appreciate the help improving our product. [Y/n]:
```

If you do, **do not enter `n` or exit telepresence**. Chances are it actually did work and is tunneling TCP traffic like
we want it to.

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