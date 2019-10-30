package engine

import (
	"errors"
	"flag"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/adapter"
	"github.com/stitchfix/flotilla-os/state"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

//
// EKSExecutionEngine submits runs to EKS.
//
type EKSExecutionEngine struct {
	eksClient  *kubernetes.Clientset
	adapter    adapter.EKSAdapter
}

//
// Initialize configures the EKSExecutionEngine and initializes internal clients
//
func (ee *EKSExecutionEngine) Initialize(conf config.Config) error {
	// @TODO: this section should be set with whatever config is necessary to
	// connect to EKS. This is currently used for local dev.
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

	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		panic(err.Error())
	}

	ee.eksClient = clientset

	adapt, err := adapter.NewEKSAdapter(conf)

	if err != nil {
		return err
	}

	ee.adapter = adapt
	return nil
}

func (ee *EKSExecutionEngine) Execute(_ state.Definition, run state.Run) (state.Run, bool, error) {
	return state.Run{}, false, errors.New("EKSExecutionEngine is only allowed to execute stateless jobs. Please use the ExecuteStateless method.")
}

func (ee *EKSExecutionEngine) ExecuteStateless(sr state.StatelessRun) (state.Run, error) {
	job, err := ee.adapter.AdaptFlotillaRunToJob(&sr)
	submittedJob, err := ee.eksClient.BatchV1().Jobs("default").Create(&job)

	if err != nil {
		return state.Run{}, err
	}

	adaptedRun, err := ee.adapter.AdaptJobToFlotillaRun(submittedJob)

	if err != nil {
		return state.Run{}, err
	}

	return adaptedRun, nil
}

// @TODO: this is a placeholder. Remove later.
func (ee *EKSExecutionEngine) Define(definition state.Definition) (state.Definition, error) {
	return definition, nil
}

// @TODO: this is a placeholder. Remove later.
func (ee *EKSExecutionEngine) Deregister(definition state.Definition) error {
	return nil
}

// @TODO: this is a placeholder. Remove later.
func (ee *EKSExecutionEngine) Terminate(run state.Run) error {
	return nil
}

// @TODO: this is a placeholder. Remove later.
func (ee *EKSExecutionEngine) Enqueue(run state.Run) error {
	return nil
}

// @TODO: this is a placeholder. Remove later.
func (ee *EKSExecutionEngine) PollRuns() ([]RunReceipt, error) {
	rr := []RunReceipt{}
	return rr, nil
}

// @TODO: this is a placeholder. Remove later.
func (ee *EKSExecutionEngine) PollStatus() (RunReceipt, error) {
	rr := RunReceipt{}
	return rr, nil
}

// TODO: this section should be set with whatever config is necessary to connect to EKS. This is currently used for local dev.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
