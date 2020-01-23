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
	ecsLogsClient logs.Client,
	eksLogsClient logs.Client,
	ecsExecutionEngine engine.Engine,
	eksExecutionEngine engine.Engine,
	stateManager state.Manager,
	ecsClusterClient cluster.Client,
	eksClusterClient cluster.Client,
	registryClient registry.Client) (App, error) {
	var app App
	app.logger = log
	app.configure(conf)

	executionService, err := services.NewExecutionService(conf, ecsExecutionEngine, eksExecutionEngine, stateManager, ecsClusterClient, eksClusterClient, registryClient, log)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing execution service")
	}
	definitionService, err := services.NewDefinitionService(conf, ecsExecutionEngine, stateManager)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing definition service")
	}
	ecsLogService, err := services.NewLogService(conf, stateManager, ecsLogsClient)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing ecs log service")
	}

	eksLogService, err := services.NewLogService(conf, stateManager, eksLogsClient)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing eks log service")
	}

	workerService, err := services.NewWorkerService(conf, stateManager)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing worker service")
	}

	templateService, err := services.NewDefinitionTemplateService(conf, stateManager)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing worker service")
	}

	ep := endpoints{
		executionService:  executionService,
		definitionService: definitionService,
		ecsLogService:     ecsLogService,
		eksLogService:     eksLogService,
		workerService:     workerService,
		templateService: templateService,
		logger:            log,
	}

	app.configureRoutes(ep)
	if err = app.initializeECSWorkers(conf, log, ecsExecutionEngine, stateManager); err != nil {
		return app, errors.Wrap(err, "problem ecs initializing workers")
	}

	if err = app.initializeEKSWorkers(conf, log, eksExecutionEngine, stateManager); err != nil {
		return app, errors.Wrap(err, "problem eks initializing workers")
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
	engine := state.ECSEngine
	workerManager, err := worker.NewWorker("worker_manager", log, conf, ee, sm, &engine)
	app.logger.Log("message", "Starting worker", "name", "worker_manager")
	if err != nil {
		return errors.Wrapf(err, "problem initializing worker with name [%s]", "worker_manager")
	}
	app.workerManager = workerManager
	return nil
}

func (app *App) initializeEKSWorkers(
	conf config.Config,
	log flotillaLog.Logger,
	ee engine.Engine,
	sm state.Manager) error {
	engine := state.EKSEngine
	workerManager, err := worker.NewWorker("worker_manager", log, conf, ee, sm, &engine)
	app.logger.Log("message", "Starting worker", "name", "worker_manager")
	if err != nil {
		return errors.Wrapf(err, "problem initializing worker with name [%s]", "worker_manager")
	}
	app.workerManager = workerManager
	return nil
}
