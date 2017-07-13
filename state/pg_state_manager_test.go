package state

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
	"time"
)

func getDB() *sqlx.DB {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func setUp() StateManager {
	db := getDB()
	//
	// Implicit testing - this will create tables
	//
	sm, _ := NewStateManager("postgres")
	//
	//
	//
	insertDefinitions(db)

	return sm
}

func insertDefinitions(db *sqlx.DB) {
	defsql := `
    INSERT INTO task_def (definition_id, image, group_name, container_name, alias, memory, command)
      VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	envsql := `
    INSERT INTO task_def_environments(task_def_id, name, value)
      VALUES ($1, $2, $3)
    `

	taskEnvSql := `
    INSERT INTO task_environments(task_id, name, value)
      VALUES ($1, $2, $3)
    `

	portsql := `
    INSERT INTO task_def_ports(task_def_id, port) VALUES ($1, $2)
    `

	taskSql := `
    INSERT INTO task (
      run_id, definition_id, cluster_name, exit_code, status,
      started_at, finished_at, instance_id, instance_dns_name, group_name
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
    )
    `

	db.MustExec(defsql, "A", "imageA", "groupZ", "containerA", "aliasA", 1024, "echo 'hi'")
	db.MustExec(defsql, "B", "imageB", "groupY", "containerB", "aliasB", 1024, "echo 'hi'")
	db.MustExec(defsql, "C", "imageC", "groupX", "containerC", "aliasC", 1024, "echo 'hi'")
	db.MustExec(defsql, "D", "imageD", "groupW", "containerD", "aliasD", 1024, "echo 'hi'")
	db.MustExec(defsql, "E", "imageE", "groupV", "containerE", "aliasE", 1024, "echo 'hi'")

	db.MustExec(envsql, "A", "E_A1", "V_A1")
	db.MustExec(envsql, "B", "E_B1", "V_B1")
	db.MustExec(envsql, "B", "E_B2", "V_B2")
	db.MustExec(envsql, "B", "E_B3", "V_B3")

	db.MustExec(portsql, "A", 10000)
	db.MustExec(portsql, "C", 10001)
	db.MustExec(portsql, "D", 10002)
	db.MustExec(portsql, "E", 10003)
	db.MustExec(portsql, "E", 10004)

	t1, _ := time.Parse(time.RFC3339, "2017-07-04T00:01:00+00:00")
	t2, _ := time.Parse(time.RFC3339, "2017-07-04T00:02:00+00:00")
	t3, _ := time.Parse(time.RFC3339, "2017-07-04T00:03:00+00:00")
	t4, _ := time.Parse(time.RFC3339, "2017-07-04T00:04:00+00:00")

	db.MustExec(taskSql, "run0", "A", "clusta", nil, "RUNNING", t1, nil, "id1", "dns1", "groupZ")
	db.MustExec(taskSql, "run1", "B", "clusta", nil, "RUNNING", t2, nil, "id1", "dns1", "groupY")
	db.MustExec(taskSql, "run2", "B", "clusta", 1, "STOPPED", t2, t3, "id1", "dns1", "groupY")
	db.MustExec(taskSql, "run3", "C", "clusta", nil, "QUEUED", nil, nil, "", "", "groupX")
	db.MustExec(taskSql, "run4", "C", "clusta", 0, "STOPPED", t3, t4, "id1", "dns1", "groupX")
	db.MustExec(taskSql, "run5", "D", "clustb", nil, "PENDING", nil, nil, "", "", "groupW")

	db.MustExec(taskEnvSql, "run0", "E0", "V0")
	db.MustExec(taskEnvSql, "run1", "E1", "V1")
	db.MustExec(taskEnvSql, "run2", "E2", "V2")
	db.MustExec(taskEnvSql, "run3", "E3_1", "V3_1")
	db.MustExec(taskEnvSql, "run3", "E3_2", "V3_2")
	db.MustExec(taskEnvSql, "run3", "E3_3", "V3_3")
}

func tearDown() {
	db := getDB()
	db.MustExec(`
    drop table if exists
      task, task_def,
      task_def_environments, task_def_ports,
      task_environments, task_status
    cascade;
    drop sequence if exists task_status_status_id_seq;
    `)
}

func TestSQLStateManager_ListDefinitions(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	var dl DefinitionList
	// Test limiting
	expectedTotal := 5
	dl, _ = sm.ListDefinitions(1, 0, "alias", "asc", nil, nil)
	if dl.Total != expectedTotal {
		t.Errorf("Expected %v total definitions, got %v", expectedTotal, dl.Total)
	}

	if len(dl.Definitions) != 1 {
		t.Errorf("Expected 1 definition returned, got %v", len(dl.Definitions))
	}

	dA := dl.Definitions[0]
	if dA.DefinitionID != "A" {
		t.Errorf("Listing returned incorrect definition, expected A but got %s", dA.DefinitionID)
	}

	if len(*dA.Env) != 1 {
		t.Errorf("Expected returned definitions to have correctly attached env vars, was %s", dA.Env)
	}

	if len(*dA.Ports) != 1 {
		t.Errorf("Expected returned definitions to have correctly attached ports, was %s", dA.Ports)
	}

	// Test ordering and offset
	dl, _ = sm.ListDefinitions(1, 1, "group_name", "asc", nil, nil)
	if dl.Definitions[0].GroupName != "groupW" {
		t.Errorf("Error ordering with offset - expected groupW but got %s", dl.Definitions[0].GroupName)
	}

	// Test order validation
	dl, err = sm.ListDefinitions(1, 0, "nonexistent_field", "asc", nil, nil)
	if err == nil {
		t.Errorf("Sorting by [nonexistent_field] did not produce an error")
	}
	dl, err = sm.ListDefinitions(1, 0, "alias", "nooop", nil, nil)
	if err == nil {
		t.Errorf("Sort order [nooop] is not valid but did not produce an error")
	}

	// Test filtering on fields
	dl, _ = sm.ListDefinitions(1, 0, "alias", "asc", map[string]string{"image": "imageC"}, nil)
	if dl.Definitions[0].Image != "imageC" {
		t.Errorf("Error filtering by field - expected imageC but got %s", dl.Definitions[0].Image)
	}

	// Test filtering on environment variables
	dl, _ = sm.ListDefinitions(1, 0, "alias", "desc", nil, map[string]string{"E_B1": "V_B1", "E_B2": "V_B2"})
	if dl.Definitions[0].DefinitionID != "B" {
		t.Errorf(
			`Expected environment variable filters (E_B1:V_B1 AND E_B2:V_B2) to yield
            definition B, but was %s`, dl.Definitions[0].DefinitionID)
	}
}

func TestSQLStateManager_GetDefinition(t *testing.T) {
	defer tearDown()
	sm := setUp()

	dE, _ := sm.GetDefinition("E")
	if dE.DefinitionID != "E" {
		t.Errorf("Expected definition E to be fetched, got %s", dE.DefinitionID)
	}

	if dE.Env != nil {
		t.Errorf("Expected empty environment but got %s", *dE.Env)
	}

	if len(*dE.Ports) != 2 {
		t.Errorf("Expected 2 ports but got %s", *dE.Ports)
	}

	_, err := sm.GetDefinition("Z")
	if err == nil {
		t.Errorf("Expected get for non-existent definition Z to return error, was nil")
	}
}

func TestSQLStateManager_CreateDefinition(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	memory := 512
	d := Definition{
		Arn:           "arn:cupcake",
		DefinitionID:  "id:cupcake",
		GroupName:     "group:cupcake",
		ContainerName: "container:cupcake",
		User:          "noone",
		Memory:        &memory,
		Alias:         "cupcake",
		Image:         "image:cupcake",
		Command:       "echo 'hi'",
		Env: &EnvList{
			{Name: "E1", Value: "V1"},
		},
		Ports: &PortsList{12345, 6789},
	}

	sm.CreateDefinition(d)

	f, err := sm.GetDefinition("id:cupcake")
	if err != nil {
		t.Errorf("Expected create definition to create definition with id [id:cupcake]")
		t.Error(err)
	}

	if f.Alias != d.Alias ||
		len(*f.Env) != len(*d.Env) ||
		len(*f.Ports) != len(*d.Ports) ||
		*f.Memory != *d.Memory {
		t.Errorf("Expected created definition to match the one passed in for creation")
	}
}

func TestSQLStateManager_UpdateDefinition(t *testing.T) {
	defer tearDown()
	sm := setUp()

	env := EnvList{
		{Name: "NEW1", Value: "NEWVAL1"},
		{Name: "NEW2", Value: "NEWVAL2"},
	}

	updates := Definition{
		Image: "updated",
		Env:   &env,
		Ports: &PortsList{}, // <---- empty, set ports to empty list
	}
	sm.UpdateDefinition("A", updates)

	d, _ := sm.GetDefinition("A")
	if d.Image != "updated" {
		t.Errorf("Expected image to be updated to [updated] but is %s", d.Image)
	}

	if d.Ports != nil {
		t.Errorf("Expected no ports after update")
	}

	if len(*d.Env) != 2 {
		t.Errorf("Expected new env to have length 2, was %v", len(*d.Env))
	}

	updatedEnv := *d.Env
	matches := 0
	for i, _ := range updatedEnv {
		updatedVar := updatedEnv[i]
		for j, _ := range env {
			expectedVar := env[j]
			if updatedVar.Name == expectedVar.Name &&
				updatedVar.Value == expectedVar.Value {
				matches++
			}
		}
	}
	if matches != len(env) {
		t.Errorf("Not all updated env vars match")
	}
}

func TestSQLStateManager_DeleteDefinition(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	sm.DeleteDefinition("A")

	_, err = sm.GetDefinition("A")
	if err == nil {
		t.Errorf("Expected querying definition after delete would return error")
	}
}

func TestSQLStateManager_ListRuns(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	expectedTotal := 6
	rl, _ := sm.ListRuns(1, 0, "started_at", "asc", nil, nil)
	if rl.Total != expectedTotal {
		t.Errorf("Expected total to be %v but was %v", expectedTotal, rl.Total)
	}

	if len(rl.Runs) != 1 {
		t.Errorf("Expected limit query to limit to 1 but was %v", len(rl.Runs))
	}

	r0 := rl.Runs[0]
	if r0.RunID != "run0" {
		t.Errorf("Listing with order returned incorrect run, expected run0 but got %s", r0.RunID)
	}

	if len(*r0.Env) != 1 {
		t.Errorf("Expected returned runs to have correctly attached env vars, was %s", r0.Env)
	}

	// Test ordering and offset
	// - there's only two, so offset 1 should return second one
	rl, err = sm.ListRuns(1, 1, "cluster_name", "desc", nil, nil)
	if rl.Runs[0].ClusterName != "clusta" {
		t.Errorf("Error ordering with offset - expected clusta but got %s", rl.Runs[0].ClusterName)
	}

	// Test order validation
	rl, err = sm.ListRuns(1, 0, "nonexistent_field", "asc", nil, nil)
	if err == nil {
		t.Errorf("Sorting by [nonexistent_field] did not produce an error")
	}
	rl, err = sm.ListRuns(1, 0, "started_at", "nooop", nil, nil)
	if err == nil {
		t.Errorf("Sort order [nooop] is not valid but did not produce an error")
	}

	// Test filtering on fields
	rl, err = sm.ListRuns(1, 0, "started_at", "asc", map[string]string{"cluster_name": "clustb"}, nil)
	if rl.Runs[0].ClusterName != "clustb" {
		t.Errorf("Error filtering by field - expected clustb but got %s", rl.Runs[0].ClusterName)
	}

	// Test filtering on environment variables
	rl, _ = sm.ListRuns(1, 0, "started_at", "desc", nil, map[string]string{"E2": "V2"})
	if rl.Runs[0].RunID != "run2" {
		t.Errorf(
			`Expected environment variable filters (E2:V2) to yield
            run run2, but was %s`, rl.Runs[0].RunID)
	}
}

func TestSQLStateManager_GetRun(t *testing.T) {
	defer tearDown()
	sm := setUp()

	r2, _ := sm.GetRun("run2")
	if r2.RunID != "run2" {
		t.Errorf("Expected run 2 to be fetched, got %s", r2.RunID)
	}

	if len(*r2.Env) != 1 {
		t.Errorf("Expected environment to have exactly one entry, but was %s", len(*r2.Env))
	}

	_, err := sm.GetRun("run100")
	if err == nil {
		t.Errorf("Expected get for non-existent run100 to return error, was nil")
	}
}

func TestSQLStateManager_CreateRun(t *testing.T) {
	defer tearDown()
	sm := setUp()

	r1 := Run{
		RunID:        "run:17",
		GroupName:    "group:cupcake",
		DefinitionID: "A",
		ClusterName:  "clusta",
		Status:       "QUEUED",
		Env: &EnvList{
			{Name: "RUN_PARAM", Value: "VAL"},
		},
	}

	ec := 137
	t1, _ := time.Parse(time.RFC3339, "2017-07-04T00:01:00+00:00")
	t2, _ := time.Parse(time.RFC3339, "2017-07-04T00:02:00+00:00")
	t1 = t1.UTC()
	t2 = t2.UTC()
	r2 := Run{
		TaskArn:      "arn1",
		RunID:        "run:18",
		GroupName:    "group:cupcake",
		DefinitionID: "A",
		ExitCode:     &ec,
		StartedAt:    &t1,
		FinishedAt:   &t2,
		ClusterName:  "clusta",
		Status:       "STOPPED",
		Env: &EnvList{
			{Name: "RUN_PARAM", Value: "VAL"},
		},
	}
	sm.CreateRun(r1)
	sm.CreateRun(r2)

	f1, _ := sm.GetRun("run:17")
	f2, _ := sm.GetRun("run:18")

	if f1.RunID != "run:17" {
		t.Errorf("Expected to fetch inserted run:17, but got %s", f1.RunID)
	}

	// Check null handling
	if f1.ExitCode != nil || f1.StartedAt != nil || f1.FinishedAt != nil {
		t.Errorf("Expected run:17 to have null exit code, started_at, and finished_at")
	}

	if f2.ExitCode == nil || f2.StartedAt == nil || f2.FinishedAt == nil {
		t.Errorf("Expected run:18 to have non null exit code, started_at, and finished_at")
	}

	if *f2.ExitCode != *r2.ExitCode {
		t.Errorf("Expected exit code %v but was %v", *r2.ExitCode, *f2.ExitCode)
	}

	if (*f2.StartedAt).UTC().String() != (*r2.StartedAt).String() {
		t.Errorf("Expected started_at %s but was %s", *r2.StartedAt, *f2.StartedAt)
	}

	if (*f2.FinishedAt).UTC().String() != (*r2.FinishedAt).String() {
		t.Errorf("Expected finished_at %s but was %s", *r2.FinishedAt, *f2.FinishedAt)
	}

}

func TestSQLStateManager_UpdateRun(t *testing.T) {
	defer tearDown()
	sm := setUp()

	ec := 1
	env := EnvList{
		{Name: "NEW1", Value: "NEWVAL1"},
		{Name: "NEW2", Value: "NEWVAL2"},
	}
	t1, _ := time.Parse(time.RFC3339, "2017-07-04T00:01:00+00:00")
	t2, _ := time.Parse(time.RFC3339, "2017-07-04T00:02:00+00:00")
	t1 = t1.UTC()
	t2 = t2.UTC()
	u := Run{
		TaskArn:    "arn1",
		ExitCode:   &ec,
		Status:     "STOPPED",
		StartedAt:  &t1,
		FinishedAt: &t2,
		Env:        &env,
	}
	sm.UpdateRun("run3", u)

	r, _ := sm.GetRun("run3")
	if *r.ExitCode != ec {
		t.Errorf("Expected update to set exit code to %v but was %v", ec, *r.ExitCode)
	}

	if (*r.StartedAt).UTC().String() != t1.String() {
		t.Errorf("Expected update to started_at to %s but was %s", t1, *r.StartedAt)
	}

	if (*r.FinishedAt).UTC().String() != t2.String() {
		t.Errorf("Expected update to set finished_at to %s but was %s", t1, *r.FinishedAt)
	}

	if r.Status != u.Status {
		t.Errorf("Expected update to set status to %s but was %s", u.Status, r.Status)
	}

	updatedEnv := *r.Env
	matches := 0
	for i, _ := range updatedEnv {
		updatedVar := updatedEnv[i]
		for j, _ := range env {
			expectedVar := env[j]
			if updatedVar.Name == expectedVar.Name &&
				updatedVar.Value == expectedVar.Value {
				matches++
			}
		}
	}
	if matches != len(env) {
		t.Errorf("Not all updated env vars match")
	}
}
