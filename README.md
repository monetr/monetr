# monetr

[![GitHub](https://github.com/monetr/monetr/actions/workflows/main.yaml/badge.svg?event=push)](https://github.com/monetr/monetr/actions/workflows/main.yaml)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr?ref=badge_shield)
[![codecov](https://codecov.io/gh/monetr/monetr/branch/main/graph/badge.svg?token=4BRVTD3VSJ)](https://codecov.io/gh/monetr/monetr)
[![Go Report Card](https://goreportcard.com/badge/github.com/monetr/monetr)](https://goreportcard.com/report/github.com/monetr/monetr)
[![DeepSource](https://deepsource.io/gh/monetr/monetr.svg/?label=active+issues&show_trend=true&token=aGbSggz8nyhTexdqi1AK1ByR)](https://deepsource.io/gh/monetr/monetr/?ref=repository-badge)
[![Discord](https://discordapp.com/api/guilds/1006270466123636836/widget.png)](https://discord.gg/68wTCXrhuq)

<p align="center">
  <img width="460" height="300" src="https://raw.githubusercontent.com/monetr/monetr/main/ui/assets/logo.svg">
</p>

monetr is a budgeting application that aims to allow people to more easily plan their recurring expenses. It is
completely free to self-host (documentation to come), but requires a Plaid account in order to communicate with banks.

monetr is currently still in heavy development; but is being alpha-tested with the latest release version constantly.

![image](https://user-images.githubusercontent.com/37967690/201261822-0103fdc6-f8a9-4cdf-9c21-0ed72bc080e3.png)

## Contributing

In order to run monetr locally you will need sandbox credentials for Plaid, you can obtain your own credentials here:
[Plaid Sign Up](https://dashboard.plaid.com/signup). Once you have a Client ID and a Client Secret you can run the
following command in the monetr project directory to start a local environment.

If you run into any missing commands you can run

```shell
brew bundle
```

To install all of the tools needed to develop monetr.

To start working on monetr, you can follow the [Local Development](https://monetr.app/developing/local/) documentation.
Or if you want to just dive in you can run the following command in the project directory:

```shell
make develop
```

This will set pretty much everything up that you need to work on monetr. This also works out of the box inside GitPod or
GitHub Codespaces.

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/from-referrer)

---

Contributions are more than welcome!

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr?ref=badge_large)
