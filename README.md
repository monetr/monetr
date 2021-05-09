# web-ui

![Gitlab pipeline status (self-hosted)](https://img.shields.io/gitlab/pipeline/monetr/web-ui/main?gitlab_url=https%3A%2F%2Fgitlab.elliotcourant.dev%2Fgithub.com&logo=gitlab)
[![Node.js CI](https://github.com/monetr/web-ui/actions/workflows/node.js.yml/badge.svg)](https://github.com/monetr/web-ui/actions/workflows/node.js.yml)
[![DeepSource](https://deepsource.io/gh/monetr/web-ui.svg/?label=active+issues&show_trend=true&token=xHI8Ef6A6rr1C_LlJ_sxzPzR)](https://deepsource.io/gh/monetr/web-ui/?ref=repository-badge)
[![DeepSource](https://deepsource.io/gh/monetr/web-ui.svg/?label=resolved+issues&show_trend=true&token=xHI8Ef6A6rr1C_LlJ_sxzPzR)](https://deepsource.io/gh/monetr/web-ui/?ref=repository-badge)

The web app for the budgeting application monetr.

## Developing Locally

Developing locally on it's own has no additional requirements outside of node. However if you want to run the UI locally
and use its entire functionality then you will need to be running the rest of the application stack in minikube. See the
REST API for details. You will need to run `make init-mini` from the REST API project first to get everything running,
once that is complete you can run the following command from the WEB UI project directory.

```bash
make local-ui
```

This will spawn a new tmux window, install any JS dependencies needed and will start the webpack dev server. You can
then open a browser window to `https://app.monetr.mini` and see the application running. You can make changes to the
application and see the changes reload.
