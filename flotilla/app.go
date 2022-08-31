package flotilla

import (
	"github.com/stitchfix/flotilla-os/queue"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/stitchfix/flotilla-os/clients/cluster"
	"github.com/stitchfix/flotilla-os/clients/logs"
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

// Start the Application.
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

// Function to initialize a new Flotilla app.
func NewApp(conf config.Config,
	log flotillaLog.Logger,
	eksLogsClient logs.Client,
	eksExecutionEngine engine.Engine,
	stateManager state.Manager,
	eksClusterClient cluster.Client,
	eksQueueManager queue.Manager,
	emrExecutionEngine engine.Engine,
	emrQueueManager queue.Manager,
) (App, error) {
	var app App
	app.logger = log
	app.configure(conf)

	executionService, err := services.NewExecutionService(conf, eksExecutionEngine, stateManager, eksClusterClient, emrExecutionEngine)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing execution service")
	}
	templateService, err := services.NewTemplateService(conf, stateManager)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing template service")
	}
	eksLogService, err := services.NewLogService(stateManager, eksLogsClient)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing eks log service")
	}
	workerService, err := services.NewWorkerService(conf, stateManager)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing worker service")
	}
	definitionService, err := services.NewDefinitionService(stateManager)
	if err != nil {
		return app, errors.Wrap(err, "problem initializing definition service")
	}

	ep := endpoints{
		executionService:  executionService,
		eksLogService:     eksLogService,
		workerService:     workerService,
		templateService:   templateService,
		logger:            log,
		definitionService: definitionService,
	}

	app.configureRoutes(ep)
	if err = app.initializeEKSWorkers(conf, log, eksExecutionEngine, emrExecutionEngine, stateManager, eksQueueManager); err != nil {
		return app, errors.Wrap(err, "problem eks initializing workers")
	}

	return app, nil
}

func (app *App) configure(conf config.Config) {
	app.address = conf.GetString("http_server_listen_address")
	if len(app.address) == 0 {
		app.address = ":5000"
	}

	readTimeout := conf.GetInt("http_server_read_timeout_seconds")
	if readTimeout == 0 {
		readTimeout = 5
	}
	writeTimeout := conf.GetInt("http_server_write_timeout_seconds")
	if writeTimeout == 0 {
		writeTimeout = 10
	}
	app.readTimeout = time.Duration(readTimeout) * time.Second
	app.writeTimeout = time.Duration(writeTimeout) * time.Second

	app.mode = conf.GetString("flotilla_mode")
	app.corsAllowedOrigins = strings.Split(conf.GetString("http_server_cors_allowed_origins"), ",")
}

func (app *App) configureRoutes(ep endpoints) {
	router := NewRouter(ep)
	c := cors.New(cors.Options{
		AllowedOrigins: app.corsAllowedOrigins,
		AllowedMethods: []string{"GET", "DELETE", "POST", "PUT"},
	})
	app.handler = c.Handler(router)
}

func (app *App) initializeEKSWorkers(
	conf config.Config,
	log flotillaLog.Logger,
	ee engine.Engine,
	emr engine.Engine,
	sm state.Manager,
	qm queue.Manager) error {
	workerManager, err := worker.NewWorker("worker_manager", log, conf, ee, emr, sm, qm)
	_ = app.logger.Log("message", "Starting worker", "name", "worker_manager")
	if err != nil {
		return errors.Wrapf(err, "problem initializing worker with name [%s]", "worker_manager")
	}
	app.workerManager = workerManager
	return nil
}

func (app *App) initializeEMRWorkers(
	conf config.Config,
	log flotillaLog.Logger,
	ee engine.Engine,
	emr engine.Engine,
	sm state.Manager,
	qm queue.Manager) error {
	workerManager, err := worker.NewWorker("worker_manager", log, conf, ee, emr, sm, qm)
	_ = app.logger.Log("message", "Starting worker", "name", "worker_manager")
	if err != nil {
		return errors.Wrapf(err, "problem initializing worker with name [%s]", "worker_manager")
	}
	app.workerManager = workerManager
	return nil
}
