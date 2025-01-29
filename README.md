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

**monetr is live! [Read the announcement post!](https://monetr.app/blog/2024-12-30-introduction/)**

monetr is a budgeting application that aims to make it easier for people to budget around recurring expenses. While
making it absolutely clear how much you have left over to budget or use for other unplanned spending. It is based off of
the now defunct [Simple](https://web.archive.org/web/20201128231953/https://www.simple.com/). It is also completely free
to [self-host](https://monetr.app/documentation/install/).

![monetr Screenshot](https://github.com/user-attachments/assets/d80847f7-8a99-4813-b15a-29094f5646ad)

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
