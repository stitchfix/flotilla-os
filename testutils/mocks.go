package testutils

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"testing"
)

// ImplementsAllTheThings
type ImplementsAllTheThings struct {
	T           *testing.T
	Calls       []string
	Definitions map[string]state.Definition
	Runs        map[string]state.Run
	Qurls       map[string]string
}

// StateManager
func (iatt *ImplementsAllTheThings) Name() string {
	iatt.Calls = append(iatt.Calls, "Name")
	return "implementer"
}
func (iatt *ImplementsAllTheThings) Initialize(conf config.Config) error {
	iatt.Calls = append(iatt.Calls, "Initialize")
	return nil
}

func (iatt *ImplementsAllTheThings) Cleanup() error {
	iatt.Calls = append(iatt.Calls, "Cleanup")
	return nil
}

func (iatt *ImplementsAllTheThings) ListDefinitions(
	limit int, offset int, sortBy string,
	order string, filters map[string]string,
	envFilters map[string]string) (state.DefinitionList, error) {
	iatt.Calls = append(iatt.Calls, "ListDefinitions")
	dl := state.DefinitionList{Total: len(iatt.Definitions)}
	for _, d := range iatt.Definitions {
		dl.Definitions = append(dl.Definitions, d)
	}
	return dl, nil
}

func (iatt *ImplementsAllTheThings) GetDefinition(definitionID string) (state.Definition, error) {
	iatt.Calls = append(iatt.Calls, "GetDefinition")
	var err error
	d, ok := iatt.Definitions[definitionID]
	if !ok {
		err = fmt.Errorf("No definition %s", definitionID)
	}
	return d, err
}
func (iatt *ImplementsAllTheThings) UpdateDefinition(definitionID string, updates state.Definition) error {
	iatt.Calls = append(iatt.Calls, "UpdateDefinition")
	return nil
}
func (iatt *ImplementsAllTheThings) CreateDefinition(d state.Definition) error {
	iatt.Calls = append(iatt.Calls, "CreateDefinition")
	iatt.Definitions[d.DefinitionID] = d
	return nil
}
func (iatt *ImplementsAllTheThings) DeleteDefinition(definitionID string) error {
	iatt.Calls = append(iatt.Calls, "DeleteDefinition")
	delete(iatt.Definitions, definitionID)
	return nil
}
func (iatt *ImplementsAllTheThings) ListRuns(limit int, offset int, sortBy string,
	order string, filters map[string]string,
	envFilters map[string]string) (state.RunList, error) {
	iatt.Calls = append(iatt.Calls, "ListRuns")
	rl := state.RunList{Total: len(iatt.Runs)}
	for _, r := range iatt.Runs {
		rl.Runs = append(rl.Runs, r)
	}
	return rl, nil
}
func (iatt *ImplementsAllTheThings) GetRun(runID string) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "GetRun")
	var err error
	r, ok := iatt.Runs[runID]
	if !ok {
		err = fmt.Errorf("No run %s", runID)
	}
	return r, err
}
func (iatt *ImplementsAllTheThings) CreateRun(r state.Run) error {
	iatt.Calls = append(iatt.Calls, "CreateRun")
	iatt.Runs[r.RunID] = r
	return nil
}
func (iatt *ImplementsAllTheThings) UpdateRun(runID string, updates state.Run) error {
	iatt.Calls = append(iatt.Calls, "UpdateRun")
	return nil
}

// QueueManager
func (iatt *ImplementsAllTheThings) QurlFor(name string) (string, error) {
	iatt.Calls = append(iatt.Calls, "QurlFor")
	qurl, _ := iatt.Qurls[name]
	return qurl, nil
}
func (iatt *ImplementsAllTheThings) Enqueue(qURL string, run state.Run) error {
	iatt.Calls = append(iatt.Calls, "Enqueue")
	return nil
}
func (iatt *ImplementsAllTheThings) Receive(qURL string) (queue.RunReceipt, error) {
	iatt.Calls = append(iatt.Calls, "Receive")
	receipt := queue.RunReceipt{
		Run: &state.Run{},
	}
	receipt.Done = func() error {
		iatt.Calls = append(iatt.Calls, "RunReceipt.Done")
		return nil
	}
	return receipt, nil
}
func (iatt *ImplementsAllTheThings) List() ([]string, error) {
	iatt.Calls = append(iatt.Calls, "List")
	res := make([]string, len(iatt.Qurls))
	i := 0
	for _, qurl := range iatt.Qurls {
		res[i] = qurl
		i++
	}
	return res, nil
}

// Cluster Client
func (iatt *ImplementsAllTheThings) CanBeRun(clusterName string, definition state.Definition) (bool, error) {
	iatt.Calls = append(iatt.Calls, "CanBeRun")
	if clusterName == "invalidcluster" {
		return false, nil
	}
	return true, nil
}

// Registry Client
func (iatt *ImplementsAllTheThings) IsImageValid(imageRef string) (bool, error) {
	iatt.Calls = append(iatt.Calls, "IsImageValid")
	if imageRef == "invalidimage" {
		return false, nil
	}
	return true, nil
}
