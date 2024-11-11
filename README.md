# monetr

[![GitHub](https://github.com/monetr/monetr/actions/workflows/main.yaml/badge.svg?event=push)](https://github.com/monetr/monetr/actions/workflows/main.yaml)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr?ref=badge_shield)
[![codecov](https://codecov.io/gh/monetr/monetr/branch/main/graph/badge.svg?token=4BRVTD3VSJ)](https://codecov.io/gh/monetr/monetr)
[![Go Report Card](https://goreportcard.com/badge/github.com/monetr/monetr)](https://goreportcard.com/report/github.com/monetr/monetr)
[![DeepSource](https://deepsource.io/gh/monetr/monetr.svg/?label=active+issues&show_trend=true&token=aGbSggz8nyhTexdqi1AK1ByR)](https://deepsource.io/gh/monetr/monetr/?ref=repository-badge)
[![Discord](https://discordapp.com/api/guilds/1006270466123636836/widget.png)](https://discord.gg/68wTCXrhuq)

<p align="center">
  <img width="460" height="300" src="https://raw.githubusercontent.com/monetr/monetr/main/docs/src/assets/logo.svg">
</p>

monetr is a budgeting application that aims to allow people to more easily plan their recurring expenses. It is
completely free to self-host (documentation to come), but requires a Plaid account in order to communicate with banks.

monetr is currently still in heavy development, but working towards a v1.0.0 release. You can see what items are still
outstanding before a v1.0.0 release for monetr will be published here: [v1.0.0
Milestone](https://github.com/monetr/monetr/milestone/3)

![image](https://github.com/user-attachments/assets/010648f1-829f-47a2-a408-c1d8759221ab)

## Contributing

In order to run monetr locally you will need sandbox credentials for Plaid, you can obtain your own credentials here:
[Plaid Sign Up](https://dashboard.plaid.com/signup). Once you have a Client ID and a Client Secret you can run the
following command in the monetr project directory to start a local environment.

To start working on monetr, you can follow the [Local
Development](https://monetr.app/documentation/development/local_development/) documentation. Or if you want to just dive
in you can run the following command in the project directory:

```shell
make develop
```

This will set pretty much everything up that you need to work on monetr.

When you want to tear it all down, you can run the following command:

```shell
make clean
```

---

Contributions are more than welcome!

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr?ref=badge_large)
