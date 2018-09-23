package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/evalphobia/logrus_sentry"
	"github.com/getsentry/raven-go"
	"github.com/mlowicki/rhythm/conf"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.AddHook(filename.NewHook())
}

func initLogging(c *conf.Logging) {
	switch c.Backend {
	case conf.LoggingBackendSentry:
		err := initSentryLogging(&c.Sentry)
		if err != nil {
			log.Fatalf("Error initializing Sentry logging: %s", err)
		}
	case conf.LoggingBackendNone:
	default:
		log.Fatalf("Unknown logging backend: %s", c.Backend)
	}
	log.Infof("Logging backend: %s", c.Backend)
}

func initSentryLogging(c *conf.LoggingSentry) error {
	cli, err := raven.New(c.DSN)
	if err != nil {
		return err
	}
	if c.RootCA != "" {
		rootCAs := x509.NewCertPool()
		certs, err := ioutil.ReadFile(c.RootCA)
		if err != nil {
			return err
		}
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			return errors.New("No certs appended")
		}
		cli.Transport = &raven.HTTPTransport{
			Client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{RootCAs: rootCAs},
				},
			},
		}
	}
	hook, err := logrus_sentry.NewWithClientSentryHook(cli, []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
	})
	if err != nil {
		return err
	}
	hook.Timeout = 0 // Do not wait for a reply.
	log.AddHook(hook)
	return nil
}
