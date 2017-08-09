package flotilla

import (
	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/logs"
	"github.com/stitchfix/flotilla-os/clients/registry"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/services"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/worker"
	"net/http"
	"time"
)

type App struct {
	address      string
	readTimeout  time.Duration
	writeTimeout time.Duration
	handler      http.Handler
	workers      []worker.Worker
}

func (app *App) Run() error {
	srv := &http.Server{
		Addr:         app.address,
		Handler:      app.handler,
		ReadTimeout:  app.readTimeout,
		WriteTimeout: app.writeTimeout,
	}
	for _, worker := range app.workers {
		go worker.Run()
	}
	return srv.ListenAndServe()
}

func NewApp(conf config.Config,
	log flotillaLog.Logger,
	lc logs.Client,
	ee engine.Engine,
	sm state.Manager,
	qm queue.Manager,
	cc cluster.Client,
	rc registry.Client) (App, error) {

	var app App
	app.configure(conf)

	executionService, err := services.NewExecutionService(conf, ee, sm, qm, cc, rc)
	if err != nil {
		return app, err
	}
	definitionService, err := services.NewDefinitionService(conf, ee, sm)
	if err != nil {
		return app, err
	}
	logService, err := services.NewLogService(conf, sm, lc)
	if err != nil {
		return app, err
	}

	ep := endpoints{
		executionService:  executionService,
		definitionService: definitionService,
		logService:        logService,
	}

	app.configureRoutes(ep)

	return app, app.initializeWorkers(conf, log, ee, qm, sm)
}

func (app *App) configure(conf config.Config) {
	app.address = conf.GetString("http.server.listen_address")
	if len(app.address) == 0 {
		app.address = ":5000"
	}

	readTimeout := conf.GetInt("http.server.read_timeout_seconds")
	if readTimeout == 0 {
		readTimeout = 5
	}
	writeTimeout := conf.GetInt("http.server.write_timeout_seconds")
	if writeTimeout == 0 {
		writeTimeout = 10
	}
	app.readTimeout = time.Duration(readTimeout) * time.Second
	app.writeTimeout = time.Duration(writeTimeout) * time.Second
}

func (app *App) configureRoutes(ep endpoints) {
	app.handler = NewRouter(ep)
}

func (app *App) initializeWorkers(
	conf config.Config,
	log flotillaLog.Logger,
	ee engine.Engine,
	qm queue.Manager,
	sm state.Manager) error {
	for _, workerName := range conf.GetStringSlice("enabled_workers") {
		wk, err := worker.NewWorker(workerName, log, conf, ee, qm, sm)
		if err != nil {
			return err
		}
		app.workers = append(app.workers, wk)
	}
	return nil
}
