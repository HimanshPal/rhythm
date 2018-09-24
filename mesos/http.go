package mesos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/extras/scheduler/callrules"
	"github.com/mesos/mesos-go/api/v1/lib/extras/store"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli"
	"github.com/mesos/mesos-go/api/v1/lib/httpcli/httpsched"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler"
	"github.com/mesos/mesos-go/api/v1/lib/scheduler/calls"
	"github.com/mlowicki/rhythm/conf"
	log "github.com/sirupsen/logrus"
)

func newClient(c *conf.Mesos, frameworkID store.Singleton) (calls.Caller, error) {
	var authConf httpcli.ConfigOpt
	if c.Auth.Type == conf.MesosAuthTypeBasic {
		authConf = httpcli.BasicAuth(c.Auth.Basic.Username, c.Auth.Basic.Password)
	} else if c.Auth.Type != conf.MesosAuthTypeNone {
		return nil, fmt.Errorf("Unknown authentication type: %s", c.Auth.Type)
	}
	tc := &tls.Config{}
	if c.RootCA != "" {
		rootCAs := x509.NewCertPool()
		certs, err := ioutil.ReadFile(c.RootCA)
		if err != nil {
			return nil, err
		}
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			return nil, errors.New("No certs appended")
		}
		tc.RootCAs = rootCAs
	}
	cli := httpcli.New(
		httpcli.Endpoint(c.BaseURL+"/api/v1/scheduler"),
		httpcli.Do(httpcli.With(
			authConf,
			httpcli.Timeout(time.Second*10),
			httpcli.TLSConfig(tc),
		)))
	return callrules.New(
		logCalls(map[scheduler.Call_Type]string{scheduler.Call_SUBSCRIBE: "Connecting..."}),
		callrules.WithFrameworkID(store.GetIgnoreErrors(frameworkID)),
	).Caller(httpsched.NewCaller(cli, httpsched.AllowReconnection(true))), nil
}

func logCalls(messages map[scheduler.Call_Type]string) callrules.Rule {
	return func(ctx context.Context, c *scheduler.Call, r mesos.Response, err error, ch callrules.Chain) (context.Context, *scheduler.Call, mesos.Response, error) {
		if message, ok := messages[c.GetType()]; ok {
			log.Println(message)
		}
		return ch(ctx, c, r, err)
	}
}
