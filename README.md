# web-ui

[![Gitlab pipeline status (self-hosted)](https://img.shields.io/gitlab/pipeline/monetr/web-ui/main?gitlab_url=https%3A%2F%2Fgitlab.elliotcourant.dev%2Fgithub.com&logo=gitlab)](https://gitlab.elliotcourant.dev/github.com/monetr/web-ui/-/pipelines)
[![Node.js CI](https://github.com/monetr/web-ui/actions/workflows/node.js.yml/badge.svg)](https://github.com/monetr/web-ui/actions/workflows/node.js.yml)
[![DeepSource](https://deepsource.io/gh/monetr/web-ui.svg/?label=active+issues&show_trend=true&token=xHI8Ef6A6rr1C_LlJ_sxzPzR)](https://deepsource.io/gh/monetr/web-ui/?ref=repository-badge)
[![DeepSource](https://deepsource.io/gh/monetr/web-ui.svg/?label=resolved+issues&show_trend=true&token=xHI8Ef6A6rr1C_LlJ_sxzPzR)](https://deepsource.io/gh/monetr/web-ui/?ref=repository-badge)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fweb-ui.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fweb-ui?ref=badge_shield)

The web app for the budgeting application monetr.

## Developing Locally

Developing locally on it's own has no additional requirements outside of node. However if you want to run the UI locally
and use its entire functionality then you will need to be running the rest of the application stack in minikube. See the
REST API for details. You will need to run `make init-mini` from the REST API project first to get everything running,
once that is complete you can run the following command from the WEB UI project directory.

```bash
make deploy-web-ui # Will get the ingress and service setup as well as a dummy pod.
make local-ui # Will tweak the service to forward to your local webpack dev server.
```

This will spawn a new tmux window, install any JS dependencies needed and will start the webpack dev server. You can
then open a browser window to `https://app.monetr.mini` and see the application running. You can make changes to the
application and see the changes reload.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fweb-ui.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fweb-ui?ref=badge_large)