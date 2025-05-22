package services

import (
	"context"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
	"testing"
)

func setUpDefinitionServiceTest(t *testing.T) (DefinitionService, *testutils.ImplementsAllTheThings) {
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"A": {DefinitionID: "A"},
			"B": {DefinitionID: "B"},
			"C": {DefinitionID: "C", ExecutableResources: state.ExecutableResources{Image: "invalidimage"}},
		},
		Runs: map[string]state.Run{
			"runA": {DefinitionID: "A", ClusterName: "A", GroupName: "A", RunID: "runA"},
			"runB": {DefinitionID: "B", ClusterName: "B", GroupName: "B", RunID: "runB"},
		},
		Qurls: map[string]string{
			"A": "a/",
			"B": "b/",
		},
	}
	ds, _ := NewDefinitionService(&imp)
	return ds, &imp
}

func TestDefinitionService_Create(t *testing.T) {
	ds, imp := setUpDefinitionServiceTest(t)
	// Check that new definition id
	// Check that define is called
	// Check that save is called and has the new definition id
	memory := int64(512)
	newValidDef := state.Definition{
		Alias:     "cupcake",
		GroupName: "group-cupcake",
		Command:   "echo 'hi'",
		ExecutableResources: state.ExecutableResources{
			Image:  "image:cupcake",
			Memory: &memory,
		},
	}

	created, _ := ds.Create(context.Background(), &newValidDef)
	if len(created.DefinitionID) == 0 {
		t.Errorf("Expected non-empty definition id")
	}

	// order matters
	expected := []string{"ListDefinitions", "CreateDefinition"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of create calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}

	// Check that the saved definition is the one with the id
	_, ok := imp.Definitions[created.DefinitionID]
	if !ok {
		t.Errorf("Expected that definition with id %s would be saved in state manager", created.DefinitionID)
	}
}

func TestDefinitionService_Create2(t *testing.T) {
	// Check that invalid definitions return errors
	ds, _ := setUpDefinitionServiceTest(t)
	var err error
	memory := int64(512)
	invalid4 := state.Definition{
		Alias:               "cupcake",
		GroupName:           "group-cupcake",
		ExecutableResources: state.ExecutableResources{Memory: &memory},
	}
	_, err = ds.Create(context.Background(), &invalid4)
	if err == nil {
		t.Errorf("Expected invalid definition with no image to result in error")
	}
}

func TestDefinitionService_Update(t *testing.T) {
	ds, imp := setUpDefinitionServiceTest(t)
	memory := int64(512)
	d := state.Definition{
		ExecutableResources: state.ExecutableResources{Memory: &memory},
	}
	ds.Update(context.Background(), "A", d)

	// order matters
	expected := []string{"GetDefinition", "UpdateDefinition"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of create calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}
}

func TestDefinitionService_Delete(t *testing.T) {
	ds, imp := setUpDefinitionServiceTest(t)
	ds.Delete(context.Background(), "A")

	// order matters
	expected := []string{"DeleteDefinition"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of create calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}
}
