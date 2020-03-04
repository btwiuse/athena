package actions

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	mw "github.com/gomods/athens/pkg/middleware"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/secure"
	"go.opencensus.io/plugin/ochttp"
)

// Service is the name of the service that we want to tag our processes with
const Service = "proxy"

// App is where all routes and middleware for the proxy
// should be defined. This is the nerve center of your
// application.
func App(conf *config.Config) (http.Handler, error) {
	// ENV is used to help switch settings based on where the
	// application is being run. Default is "development".
	ENV := conf.GoEnv
	store, err := GetStorage(conf.StorageType, conf.Storage, conf.TimeoutDuration())
	if err != nil {
		err = fmt.Errorf("error getting storage configuration (%s)", err)
		return nil, err
	}

	if conf.GithubToken != "" {
		if conf.NETRCPath != "" {
			fmt.Println("Cannot provide both GithubToken and NETRCPath. Only provide one.")
			os.Exit(1)
		}

		netrcFromToken(conf.GithubToken)
	}

	// mount .netrc to home dir
	// to have access to private repos.
	initializeAuthFile(conf.NETRCPath)

	// mount .hgrc to home dir
	// to have access to private repos.
	initializeAuthFile(conf.HGRCPath)

	logLvl, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		return nil, err
	}
	lggr := log.New(conf.CloudRuntime, logLvl)

	r := mux.NewRouter()
	r.Use(mw.LogEntryMiddleware(lggr))
	r.Use(mw.RequestLogger)
	r.Use(secure.New(secure.Options{
		SSLRedirect:     conf.ForceSSL,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	}).Handler)
	r.Use(mw.ContentType)

	// RegisterExporter will register an exporter where we will export our traces to.
	// The error from the RegisterExporter would be nil if the tracer was specified by
	// the user and the trace exporter was created successfully.
	// RegisterExporter returns the function that all traces are flushed to the exporter
	// and the exporter needs to be stopped. The function should be called when the exporter
	// is no longer needed.
	flushTraces, err := observ.RegisterExporter(
		conf.TraceExporter,
		conf.TraceExporterURL,
		Service,
		ENV,
	)
	if err != nil {
		lggr.Infof("%s", err)
	} else {
		defer flushTraces()
	}

	// RegisterStatsExporter will register an exporter where we will collect our stats.
	// The error from the RegisterStatsExporter would be nil if the proper stats exporter
	// was specified by the user.
	flushStats, err := observ.RegisterStatsExporter(r, conf.StatsExporter, Service)
	if err != nil {
		lggr.Infof("%s", err)
	} else {
		defer flushStats()
	}

	user, pass, ok := conf.BasicAuth()
	if ok {
		r.Use(basicAuth(user, pass))
	}

	if !conf.FilterOff() {
		mf, err := module.NewFilter(conf.FilterFile)
		if err != nil {
			lggr.Fatal(err)
		}
		r.Use(mw.NewFilterMiddleware(mf, conf.GlobalEndpoint))
	}

	// Having the hook set means we want to use it
	if vHook := conf.ValidatorHook; vHook != "" {
		r.Use(mw.NewValidationMiddleware(vHook))
	}

	proxyRouter := r
	if err := addProxyRoutes(
		proxyRouter,
		store,
		lggr,
		conf,
	); err != nil {
		err = fmt.Errorf("error adding proxy routes (%s)", err)
		return nil, err
	}

	h := &ochttp.Handler{
		Handler: r,
	}

	return h, nil
}
