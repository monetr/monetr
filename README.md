# monetr

[![GitHub](https://github.com/monetr/monetr/actions/workflows/main.yaml/badge.svg?event=push)](https://github.com/monetr/monetr/actions/workflows/main.yaml)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr?ref=badge_shield)
[![codecov](https://codecov.io/gh/monetr/monetr/branch/main/graph/badge.svg?token=4BRVTD3VSJ)](https://codecov.io/gh/monetr/monetr)
[![Go Report Card](https://goreportcard.com/badge/github.com/monetr/monetr)](https://goreportcard.com/report/github.com/monetr/monetr)
[![wakatime](https://wakatime.com/badge/user/e7d2c225-af72-41dc-bf39-f4a8108dc790/project/30965d1c-e425-4da3-9a31-7b1ca82dfaef.svg)](https://wakatime.com/badge/user/e7d2c225-af72-41dc-bf39-f4a8108dc790/project/30965d1c-e425-4da3-9a31-7b1ca82dfaef)

<p align="center">
  <img width="460" height="300" src="https://raw.githubusercontent.com/monetr/monetr/main/ui/assets/logo.svg">
</p>

monetr is a budgeting application that aims to allow people to more easily plan their recurring expenses. It is
completely free to self-host (documentation to come), but requires a Plaid account in order to communicate with banks.

monetr is currently still in heavy development; but is being alpha-tested with the latest release version constantly.

![image](https://user-images.githubusercontent.com/37967690/168505075-5f5f11d4-0546-4594-a0b3-4ee860a9938e.png)

## Status

| Environment   | Uptime                                                                                                        | Status Page                                                       |
|---------------|---------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------|
| Alpha Testing | ![Uptime Robot ratio (30 days)](https://img.shields.io/uptimerobot/ratio/m789641931-ce8fe24a641913b47027297d) | [Status Page](https://stats.uptimerobot.com/zAjyOcGm7E/789641931) |

## Contributing

In order to run monetr locally you will need sandbox credentials for Plaid, you can obtain your own credentials here:
[Plaid Sign Up](https://dashboard.plaid.com/signup). Once you have a Client ID and a Client Secret you can run the
following command in the monetr project directory to start a local environment.

If you run into any missing commands you can run

```shell
brew bundle
```

To install all of the tools needed to develop monetr.

You'll want to install node dependencies before starting the containers, as yarn install is extremely slow inside
docker.

```shell
yarn install
PLAID_CLIENT_ID="Your Client ID" PLAID_CLIENT_SECRET="Your Secret" docker compose up
```

You can certainly develop monetr without Plaid credentials, but at the moment most functionality will not work without
them.

This will start all of the services needed to do basic development on monetr. Some services are not currently included
in this workflow though, like vault. I am still working on an easier way to develop locally that will not require
providing secrets each time, but will also allow for config customizations.

The UI and API will "hot reload" while running in docker compose, any changes made to them will be auto built to make
development easier.

The API is run inside Docker using [delve](https://github.com/go-delve/delve), so if you want to or prefer step
debugging for development you can connect your editor to `localhost:2345` for remote debugging.

It may take a few moments for the API and the UI to get running. But once they are you can access the application by
navigating to `http://localhost` in a browser.

---

If you want to/need to develop using webhooks from Plaid then you can run the following command to enable ngrok.

```shell
docker compose up ngrok -d
PLAID_CLIENT_ID="Your Client ID" PLAID_CLIENT_SECRET="Your Secret" docker compose restart monetr
```

This will restart the `monetr` API container as well as bring up the ngrok container. If you have ngrok API credentials
you can specify them using the `NGROK_AUTH` environment variable. If you do not have credentials you will be given a
temporary URL where webhooks will be directed. The hot-reload container for the API will also grab this URL
automatically when it is restarted as well as enable Plaid webhooks. Note: Right now it is required to specify the Plaid
credentials again, otherwise they would be overwritten when the monetr container is restarted.

Once ngrok is running you can access it by going to `http://localhost:4040`.

---

Once you have finished any work and you want to tear down the local development environment you can run:

```shell
docker compose down --remove-orphans -v
```

Note: If you have created any Plaid links, especially ones with webhooks enabled; it is recommended to remove them
before tearing everything down. This way Plaid isn't left with useless stuff in their sandbox and doesn't continue to
send webhooks to an ngrok instance that isn't being used anymore.

Once everything is down you can also run:

```shell
make clean
```

To clean up any extra things like node modules or temp files that may have been created in the project directory.

---

Most of these can be run directly through `make` but I cannot guarantee that it will support Windows at the moment. To
get started you can run `make develop` to start the development environment. `make logs` will tail all the logs for
everything running. `make webhooks` will provision and setup containers for developing webhooks, and `make shutdown`
will take down the development environment.

Contributions are more than welcome!

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fmonetr%2Fmonetr?ref=badge_large)
