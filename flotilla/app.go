package flotilla

import (
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/logs"
	"github.com/stitchfix/flotilla-os/clients/registry"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/services"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/worker"
)

type App struct {
	address            string
	mode               string
	corsAllowedOrigins []string
	logger             flotillaLog.Logger
	readTimeout        time.Duration
	writeTimeout       time.Duration
	handler            http.Handler
	workerManager      worker.Worker
}

func (app *App) Run() error {
	srv := &http.Server{
		Addr:         app.address,
		Handler:      app.handler,
		ReadTimeout:  app.readTimeout,
		WriteTimeout: app.writeTimeout,
	}
	// Start worker manager's run goroutine.
	app.workerManager.GetTomb().Go(app.workerManager.Run)
	return srv.ListenAndServe()
}

func NewApp(conf config.Config,
	log flotillaLog.Logger,
	lc logs.Client,
	ee engine.Engine,
	sm state.Manager,
	cc cluster.Client,
	rc registry.Client) (App, error) {

	var app App
	app.logger = log
	app.configure(conf)

	executionService, err := services.NewExecutionService(conf, ee, sm, cc, rc)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing execution service")
	}
	definitionService, err := services.NewDefinitionService(conf, ee, sm)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing definition service")
	}
	logService, err := services.NewLogService(conf, sm, lc)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing log service")
	}
	workerService, err := services.NewWorkerService(conf, sm)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing worker service")
	}

	ep := endpoints{
		executionService:  executionService,
		definitionService: definitionService,
		logService:        logService,
		workerService:     workerService,
		logger:            log,
	}

	app.configureRoutes(ep)
	if err = app.initializeECSWorkers(conf, log, ee, sm); err != nil {
		return app, errors.Wrap(err, "problem initializing workers")
	}
	return app, nil
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

	app.mode = conf.GetString("flotilla_mode")
	app.corsAllowedOrigins = conf.GetStringSlice("http.server.cors_allowed_origins")
}

func (app *App) configureRoutes(ep endpoints) {
	if app.mode == "dev" || app.mode == "test" {
		app.logger.Log(
			"message", "WARNING - enabling CORS",
			"origins", strings.Join(app.corsAllowedOrigins, ","))
		router := NewRouter(ep)
		c := cors.New(cors.Options{
			AllowedOrigins: app.corsAllowedOrigins,
			AllowedMethods: []string{"GET", "DELETE", "POST", "PUT"},
		})
		app.handler = c.Handler(router)
	} else {
		app.handler = NewRouter(ep)
	}
}

func (app *App) initializeECSWorkers(
	conf config.Config,
	log flotillaLog.Logger,
	ee engine.Engine,
	sm state.Manager) error {
	engine := state.DefaultEngine
	workerManager, err := worker.NewWorker("worker_manager", log, conf, ee, sm, &engine)
	app.logger.Log("message", "Starting worker", "name", "worker_manager")
	if err != nil {
		return errors.Wrapf(err, "problem initializing worker with name [%s]", "worker_manager")
	}
	app.workerManager = workerManager
	return nil
}
