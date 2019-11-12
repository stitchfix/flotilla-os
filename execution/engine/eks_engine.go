package engine

import (
	"encoding/base64"
	"flag"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/adapter"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//
// EKSExecutionEngine submits runs to EKS.
//
type EKSExecutionEngine struct {
	kClient      *kubernetes.Clientset
	adapter      adapter.EKSAdapter
	qm           queue.Manager
	log          flotillaLog.Logger
	jobQueue     string
	jobNamespace string
	jobTtl       int
}

//
// Initialize configures the EKSExecutionEngine and initializes internal clients
//
func (ee *EKSExecutionEngine) Initialize(conf config.Config) error {
	kStr, err := base64.StdEncoding.DecodeString(conf.GetString("eks.kubeconfig"))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(conf.GetString("eks.kubeconfig_path"), kStr, 0644)
	if err != nil {
		return err
	}

	kubeconfig := flag.String("kubeconfig",
		conf.GetString("eks.kubeconfig_path"),
		"(optional) absolute tmpPath to the kubeconfig file")

	flag.Parse()

	clientConf, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}

	kClient, err := kubernetes.NewForConfig(clientConf)
	if err != nil {
		return err
	}

	ee.jobQueue = conf.GetString("eks.job_queue")
	ee.jobNamespace = conf.GetString("eks.job_namespace")
	ee.jobTtl = conf.GetInt("eks.job_ttl")

	ee.kClient = kClient

	adapt, err := adapter.NewEKSAdapter(conf)

	if err != nil {
		return err
	}

	ee.adapter = adapt
	return nil
}

func (ee *EKSExecutionEngine) Execute(td state.Definition, run state.Run) (state.Run, bool, error) {
	job, err := ee.adapter.AdaptFlotillaDefinitionAndRunToJob(td, run)
	result, err := ee.kClient.BatchV1().Jobs(ee.jobNamespace).Create(&job)
	if err != nil {
		return state.Run{}, false, err
	}

	ee.log.Log("submitted job", run.RunID)

	adaptedRun, err := ee.adapter.AdaptJobToFlotillaRun(result, run)
	if err != nil {
		return state.Run{}, false, err
	}

	return adaptedRun, false, nil
}

func (ee *EKSExecutionEngine) Terminate(run state.Run) error {
	gracePeriod := int64(0)
	_ = ee.log.Log("terminating run=", run.RunID)
	deleteOptions := &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod}
	return ee.kClient.BatchV1().Jobs(ee.jobNamespace).Delete(run.RunID, deleteOptions)
}

func (ee *EKSExecutionEngine) Enqueue(run state.Run) error {
	// Get qurl
	qurl, err := ee.qm.QurlFor(ee.jobQueue, false)
	if err != nil {
		return errors.Wrapf(err, "problem getting queue url for [%s]", run.ClusterName)
	}

	// Queue run
	if err = ee.qm.Enqueue(qurl, run); err != nil {
		return errors.Wrapf(err, "problem enqueing run [%s] to queue [%s]", run.RunID, qurl)
	}
	return nil
}

func (ee *EKSExecutionEngine) PollRuns() ([]RunReceipt, error) {
	qurl, err := ee.qm.QurlFor(ee.jobQueue, false)
	if err != nil {
		return nil, errors.Wrap(err, "problem listing queues to poll")
	}
	queues := []string{qurl}
	var runs []RunReceipt
	for _, qurl := range queues {
		//
		// Get new queued Run
		//
		runReceipt, err := ee.qm.ReceiveRun(qurl)

		if err != nil {
			return runs, errors.Wrapf(err, "problem receiving run from queue url [%s]", qurl)
		}

		if runReceipt.Run == nil {
			continue
		}

		runs = append(runs, RunReceipt{runReceipt})
	}
	return runs, nil
}

//
// PollStatus is a dummy function as EKS does not emit task status
// change events.
//
func (ee *EKSExecutionEngine) PollStatus() (RunReceipt, error) {
	return RunReceipt{}, nil
}

//
// Define returns a blank task definition and an error for the EKS engine.
//
func (ee *EKSExecutionEngine) Define(td state.Definition) (state.Definition, error) {
	updated := td
	// TODO: how to deal w/ ARN?
	updated.Arn = td.DefinitionID
	return updated, nil
}

//
// Deregister returns an error for the EKS engine.
//
func (ee *EKSExecutionEngine) Deregister(definition state.Definition) error {
	return errors.Errorf("EKSExecutionEngine does not allow for deregistering of task definitions.")
}

func (ee *EKSExecutionEngine) Get(run state.Run) (state.Run, error) {
	job, err := ee.kClient.BatchV1().Jobs(ee.jobNamespace).Get(run.RunID, metav1.GetOptions{})

	if err != nil {
		return state.Run{}, errors.Errorf("error getting kubernetes job %s", err)
	}

	updates, err := ee.adapter.AdaptJobToFlotillaRun(job, run)

	if err != nil {
		return state.Run{}, errors.Errorf("error adapting kubernetes job to flotilla run %s", err)
	}

	return updates, nil
}
