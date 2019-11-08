package engine

import (
	"flag"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/adapter"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	 metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

//
// EKSExecutionEngine submits runs to EKS.
//
type EKSExecutionEngine struct {
	kClient    *kubernetes.Clientset
	adapter    adapter.EKSAdapter
	qm         queue.Manager
	statusQurl string
	log        flotillaLog.Logger
}

//
// Initialize configures the EKSExecutionEngine and initializes internal clients
//
func (ee *EKSExecutionEngine) Initialize(conf config.Config) error {
	// TODO: this section should be set with whatever config is necessary to connect to EKS.
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	kubeConf, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	kClient, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		panic(err.Error())
	}

	ee.kClient = kClient

	adapt, err := adapter.NewEKSAdapter(conf)

	if err != nil {
		return err
	}

	ee.adapter = adapt

	statusQueue := conf.GetString("queue.status")
	ee.statusQurl, err = ee.qm.QurlFor(statusQueue, false)
	if err != nil {
		return errors.Wrapf(err, "problem getting queue url for status queue with name [%s]", statusQueue)
	}

	return nil
}

func (ee *EKSExecutionEngine) Execute(td state.Definition, run state.Run) (state.Run, bool, error) {
	job, err := ee.adapter.AdaptFlotillaDefinitionAndRunToJob(td, run)

	result, err := ee.kClient.BatchV1().Jobs("default").Create(&job)
	if err != nil {
		return state.Run{}, false, err
	}

	adaptedRun, err := ee.adapter.AdaptJobToFlotillaRun(result, run)
	if err != nil {
		return state.Run{}, false, err
	}

	return adaptedRun, false, nil
}

func (ee *EKSExecutionEngine) Terminate(run state.Run) error {
	return nil
}

func (ee *EKSExecutionEngine) Enqueue(run state.Run) error {
	// Get qurl
	qurl, err := ee.qm.QurlFor(run.ClusterName, true)
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
	queues, err := ee.qm.List()
	if err != nil {
		return nil, errors.Wrap(err, "problem listing queues to poll")
	}

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
	job, err := ee.kClient.BatchV1().Jobs("default").Get(run.RunID, metav1.GetOptions{})

	if err != nil {
		return state.Run{}, errors.Errorf("error getting kubernetes job", err)
	}

	updates, err := ee.adapter.AdaptJobToFlotillaRun(job, run)

	if err != nil {
		return state.Run{}, errors.Errorf("error adapting kubernetes job to flotilla run", err)
	}

	return updates, nil
}

// TODO: this section should be set with whatever config is necessary to connect to EKS. This is currently used for local dev.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
