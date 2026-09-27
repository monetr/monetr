package commands

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/monetr/monetr/server/certs"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func StatusCommand(parent *cobra.Command) {
	var arguments struct {
		BaseURL                 string
		TLSCertificateAuthority string
		TLSCertificate          string
		TLSKey                  string
		ServerName              string
		InsecureSkipVerify      bool
	}

	command := &cobra.Command{
		Use:   "status",
		Short: "Check the status of the monetr server",
		Long:  "Checks the status of the monetr server by making a request to the health check endpoint using the specified URL and/or TLS certificates. Appends /api/health to the end of the base URL. Exits with a non-zero exit code if the server is not healthy.",
		// Don't print the usage when the server is unhealthy, this is meant to be
		// run by health checks and the usage is just noise in the output.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.Debug("config file loaded", "config", configFileName)
			}

			{ // Allow overrides from the flags
				configuration.Server.TLSClientCertificateAuthority = myownsanity.CoalesceStrings(
					arguments.TLSCertificateAuthority,
					configuration.Server.TLSClientCertificateAuthority,
				)
				configuration.Server.TLSCertificate = myownsanity.CoalesceStrings(
					arguments.TLSCertificate,
					configuration.Server.TLSCertificate,
				)
				configuration.Server.TLSKey = myownsanity.CoalesceStrings(
					arguments.TLSKey,
					configuration.Server.TLSKey,
				)
			}

			baseUrl := arguments.BaseURL
			serverName := arguments.ServerName
			if baseUrl == "" {
				// If no base URL is provided then make the request to the local server
				// the same way that serve would be listening.
				protocol := "http"
				if configuration.Server.TLSCertificate != "" && configuration.Server.TLSKey != "" {
					protocol = "https"
				}
				baseUrl = fmt.Sprintf(
					"%s://%s",
					protocol,
					net.JoinHostPort(
						"localhost",
						strconv.Itoa(configuration.Server.ListenPort),
					),
				)
				// The request is made to localhost but the server's certificate is
				// most likely issued for the external hostname, so verify against that
				// instead.
				serverName = myownsanity.CoalesceStrings(
					serverName,
					configuration.Server.GetHostname(),
				)
			}

			requestUrl, err := url.Parse(baseUrl)
			if err != nil {
				log.ErrorContext(
					cmd.Context(),
					"failed to parse base url",
					"err", err,
				)
				return errors.WithStack(err)
			}

			requestUrl = requestUrl.JoinPath("/api/health")

			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			if requestUrl.Scheme == "https" {
				certificates, err := certs.NewFileSource(log, certs.Options{
					CACertificatePath:  configuration.Server.TLSClientCertificateAuthority,
					CertificatePath:    configuration.Server.TLSCertificate,
					KeyPath:            configuration.Server.TLSKey,
					ServerName:         serverName,
					InsecureSkipVerify: arguments.InsecureSkipVerify,
				})
				if err != nil {
					log.ErrorContext(
						cmd.Context(),
						"failed to load TLS certificates",
						"err", err,
					)
					return errors.Wrap(err, "failed to load TLS certificates")
				}
				// The certificates are not started because this is a single request,
				// but stop still needs to be called to close the watcher.
				defer certificates.Stop()

				client.Transport = &http.Transport{
					TLSClientConfig: certificates.ClientConfig(),
				}
			}

			return checkStatus(cmd.Context(), log, client, requestUrl)
		},
	}

	command.PersistentFlags().StringVarP(&arguments.BaseURL, "base-url", "b", "", "Base URL for which the status request will be made, should not include the health path but should include any base path modifications. Defaults to: localhost on the configured listen port, using https if a TLS certificate and key are configured")
	command.PersistentFlags().StringVarP(&arguments.TLSCertificateAuthority, "tls-ca", "a", "", "Path to the TLS Certificate Authority used to verify the server's certificate. Defaults to: the configured TLS client certificate authority, or the system certificate pool if none is configured")
	command.PersistentFlags().StringVarP(&arguments.TLSCertificate, "tls-crt", "t", "", "Path to the TLS Certificate presented to the server as the client certificate. If the server requires client certificates then this must include the client auth extended key usage. Defaults to: the configured TLS certificate")
	command.PersistentFlags().StringVarP(&arguments.TLSKey, "tls-key", "k", "", "Path to the TLS Key for the client certificate. Defaults to: the configured TLS key")
	command.PersistentFlags().StringVarP(&arguments.ServerName, "server-name", "s", "", "Hostname the server's certificate is verified against. Defaults to: the hostname of the external URL when no base URL is provided, otherwise the hostname of the base URL")
	command.PersistentFlags().BoolVarP(&arguments.InsecureSkipVerify, "insecure", "i", false, "Skip verifying the server's certificate. Defaults to: false")

	parent.AddCommand(command)
}

// checkStatus makes a request to the health endpoint at the provided URL and
// returns an error if the server is not healthy.
func checkStatus(
	ctx context.Context,
	log *slog.Logger,
	client *http.Client,
	requestUrl *url.URL,
) error {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestUrl.String(),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create status request")
	}

	response, err := client.Do(request)
	if err != nil {
		log.ErrorContext(
			ctx,
			"failed to make status request",
			"url", requestUrl.String(),
			"err", err,
		)
		return errors.Wrap(err, "failed to make status request")
	}
	defer response.Body.Close()

	// The health endpoint returns a small JSON body, but limit how much is read
	// in case the URL is pointed at something else.
	body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
	if err != nil {
		return errors.Wrap(err, "failed to read status response")
	}

	if response.StatusCode != http.StatusOK {
		log.ErrorContext(
			ctx,
			"monetr is not healthy",
			"url", requestUrl.String(),
			"status", response.StatusCode,
			"body", string(body),
		)
		return errors.Errorf("monetr is not healthy, status code: %d", response.StatusCode)
	}

	log.InfoContext(
		ctx,
		"monetr is healthy",
		"url", requestUrl.String(),
		"body", string(body),
	)

	return nil
}
