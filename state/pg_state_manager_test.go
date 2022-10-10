package state

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stitchfix/flotilla-os/config"
)

func getDB(conf config.Config) *sqlx.DB {
	db, err := sqlx.Connect("postgres", conf.GetString("database_url"))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func setUp() Manager {
	conf, _ := config.NewConfig(nil)

	db := getDB(conf)
	os.Setenv("STATE_MANAGER", "postgres")
	os.Setenv("CREATE_DATABASE_SCHEMA", "true")
	sm, err := NewStateManager(conf)
	fmt.Println(err)

	insertDefinitions(db)

	return sm
}

func insertDefinitions(db *sqlx.DB) {
	defsql := `
    INSERT INTO task_def (definition_id, image, group_name, alias, memory, command, env)
      VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	portsql := `
    INSERT INTO task_def_ports(task_def_id, port) VALUES ($1, $2)
    `

	taskDefTagsSQL := `
	INSERT INTO task_def_tags(task_def_id, tag_id) VALUES($1, $2)
	`
	tagSQL := `
	INSERT INTO tags(text) VALUES($1)
	`

	taskSQL := `
    INSERT INTO task (
      run_id, definition_id, cluster_name, alias, image, exit_code, status,
      started_at, finished_at, instance_id, instance_dns_name, group_name, env, engine, "user"
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'eks', 'foo'
    )
    `

	db.MustExec(defsql,
		"A", "imageA", "groupZ", "aliasA", 1024, "echo 'hi'", `[{"name":"E_A1","value":"V_A1"}]`)
	db.MustExec(defsql,
		"B", "imageB", "groupY", "aliasB", 1024, "echo 'hi'",
		`[{"name":"E_B1","value":"V_B1"},{"name":"E_B2","value":"V_B2"},{"name":"E_B3","value":"V_B3"}]`)
	db.MustExec(defsql, "C", "imageC", "groupX", "aliasC", 1024, "echo 'hi'", nil)
	db.MustExec(defsql, "D", "imageD", "groupW", "aliasD", 1024, "echo 'hi'", nil)
	db.MustExec(defsql, "E", "imageE", "groupV", "aliasE", 1024, "echo 'hi'", nil)

	db.MustExec(portsql, "A", 10000)
	db.MustExec(portsql, "C", 10001)
	db.MustExec(portsql, "D", 10002)
	db.MustExec(portsql, "E", 10003)
	db.MustExec(portsql, "E", 10004)

	db.MustExec(tagSQL, "tagA")
	db.MustExec(tagSQL, "tagB")
	db.MustExec(tagSQL, "tagC")

	db.MustExec(taskDefTagsSQL, "A", "tagA")
	db.MustExec(taskDefTagsSQL, "A", "tagC")
	db.MustExec(taskDefTagsSQL, "B", "tagB")

	t1, _ := time.Parse(time.RFC3339, "2017-07-04T00:01:00+00:00")
	t2, _ := time.Parse(time.RFC3339, "2017-07-04T00:02:00+00:00")
	t3, _ := time.Parse(time.RFC3339, "2017-07-04T00:03:00+00:00")
	t4, _ := time.Parse(time.RFC3339, "2017-07-04T00:04:00+00:00")

	db.MustExec(taskSQL,
		"run0", "A", "clusta", "aliasA", "imgA", nil, StatusRunning, t1, nil, "id1", "dns1", "groupZ", `[{"name":"E0","value":"V0"}]`)
	db.MustExec(
		taskSQL, "run1", "B", "clusta", "aliasB", "imgB", nil, StatusRunning, t2, nil, "id1", "dns1", "groupY", `[{"name":"E1","value":"V1"}]`)

	db.MustExec(
		taskSQL, "run2", "B", "clusta", "aliasB", "imgB", 1, StatusStopped, t2, t3, "id1", "dns1", "groupY", `[{"name":"E2","value":"V2"}]`)

	db.MustExec(taskSQL,
		"run3", "C", "clusta", "aliasC", "imgC", nil, StatusQueued, nil, nil, "", "", "groupX",
		`[{"name":"E3_1","value":"V3_1"},{"name":"E3_2","value":"v3_2"},{"name":"E3_3","value":"V3_3"}]`)

	db.MustExec(taskSQL, "run4", "C", "clusta", "aliasC", "imgC", 0, StatusStopped, t3, t4, "id1", "dns1", "groupX", nil)
	db.MustExec(taskSQL, "run5", "D", "clustb", "aliasD", "imgD", nil, StatusPending, nil, nil, "", "", "groupW", nil)
}

func tearDown() {
	conf, _ := config.NewConfig(nil)
	db := getDB(conf)
	db.MustExec(`
		DELETE FROM task_def_ports;
		DELETE FROM task_def_tags;
		DELETE FROM task_status;
		DELETE FROM task;
		DELETE FROM task_def;
		DELETE FROM tags;
  `)
}

func TestSQLStateManager_ListDefinitions(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	var dl DefinitionList
	// Test limiting
	expectedTotal := 5
	dl, err = sm.ListDefinitions(1, 0, "alias", "asc", nil, nil)
	if err != nil {
		t.Errorf(err.Error())
	}

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
		t.Errorf("Expected returned definitions to have correctly attached env vars, was %v", dA.Env)
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
	dl, _ = sm.ListDefinitions(1, 0, "alias", "asc", map[string][]string{"image": {"imageC"}}, nil)
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

	_, err := sm.GetDefinition("Z")
	if err == nil {
		t.Errorf("Expected get for non-existent definition Z to return error, was nil")
	}
}

func TestSQLStateManager_GetDefinitionByAlias(t *testing.T) {
	defer tearDown()
	sm := setUp()

	dE, _ := sm.GetDefinitionByAlias("aliasE")
	if dE.DefinitionID != "E" {
		t.Errorf("Expected definition E to be fetched, got %s", dE.DefinitionID)
	}

	if dE.Env != nil {
		t.Errorf("Expected empty environment but got %s", *dE.Env)
	}

	_, err := sm.GetDefinitionByAlias("aliasZ")
	if err == nil {
		t.Errorf("Expected get for non-existent definition Z to return error, was nil")
	}
}

func TestSQLStateManager_CreateDefinition(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	memory := int64(512)
	d := Definition{
		DefinitionID: "id:cupcake",
		GroupName:    "group:cupcake",
		Alias:        "cupcake",
		Command:      "echo 'hi'",
		ExecutableResources: ExecutableResources{
			Memory: &memory,
			Image:  "image:cupcake",
			Env: &EnvList{
				{Name: "E1", Value: "V1"},
			},
			Ports: &PortsList{12345, 6789},
			Tags:  &Tags{"apple", "orange", "tiger"},
		},
	}

	err = sm.CreateDefinition(d)
	if err != nil {
		t.Errorf(err.Error())
	}

	f, err := sm.GetDefinition("id:cupcake")
	if err != nil {
		t.Errorf("Expected create definition to create definition with id [id:cupcake]")
		t.Error(err)
	}

	if f.Alias != d.Alias ||
		len(*f.Env) != len(*d.Env) ||
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

	tags := Tags{
		"cupcake",
	}
	updates := Definition{
		ExecutableResources: ExecutableResources{
			Tags:  &tags,
			Image: "updated",
			Env:   &env,
			Ports: &PortsList{}, // <---- empty, set ports to empty list
		},
	}
	_, err := sm.UpdateDefinition("A", updates)
	if err != nil {
		t.Errorf(err.Error())
	}

	d, _ := sm.GetDefinition("A")
	if d.Image != "updated" {
		t.Errorf("Expected image to be updated to [updated] but is %s", d.Image)
	}

	if len(*d.Env) != 2 {
		t.Errorf("Expected new env to have length 2, was %v", len(*d.Env))
	}

	updatedEnv := *d.Env
	matches := 0
	for i := range updatedEnv {
		updatedVar := updatedEnv[i]
		for j := range env {
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
	err = sm.DeleteDefinition("A")
	if err != nil {
		t.Errorf(err.Error())
	}

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
	rl, err := sm.ListRuns(1, 0, "started_at", "asc", nil, nil, nil)
	if err != nil {
		t.Errorf(err.Error())
	}

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

	if r0.Env == nil {
		t.Errorf("Expected non-nil env for run")
	}

	if len(*r0.Env) != 1 {
		t.Errorf("Expected returned runs to have correctly attached env vars, was %v", r0.Env)
	}

	// Test ordering and offset
	// - there's only two, so offset 1 should return second one
	rl, err = sm.ListRuns(1, 1, "cluster_name", "desc", nil, nil, nil)
	if rl.Runs[0].ClusterName != "clusta" {
		t.Errorf("Error ordering with offset - expected clusta but got %s", rl.Runs[0].ClusterName)
	}

	// Test order validation
	rl, err = sm.ListRuns(1, 0, "nonexistent_field", "asc", nil, nil, nil)
	if err == nil {
		t.Errorf("Sorting by [nonexistent_field] did not produce an error")
	}
	rl, err = sm.ListRuns(1, 0, "started_at", "nooop", nil, nil, nil)
	if err == nil {
		t.Errorf("Sort order [nooop] is not valid but did not produce an error")
	}

	// Test filtering on fields
	rl, err = sm.ListRuns(1, 0, "started_at", "asc", map[string][]string{"cluster_name": {"clustb"}}, nil, nil)
	if rl.Runs[0].ClusterName != "clustb" {
		t.Errorf("Error filtering by field - expected clustb but got %s", rl.Runs[0].ClusterName)
	}

	// Test filtering on environment variables
	rl, err = sm.ListRuns(1, 0, "started_at", "desc", nil, map[string]string{"E2": "V2"}, nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	if rl.Runs[0].RunID != "run2" {
		t.Errorf(
			`Expected environment variable filters (E2:V2) to yield
            run run2, but was %s`, rl.Runs[0].RunID)
	}
}

func TestSQLStateManager_ListRuns2(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	expectedTotal := 1
	expectedRun := "run4"
	rl, err := sm.ListRuns(100, 0, "started_at", "asc", map[string][]string{
		"started_at_since": {
			"2017-07-04T00:02:59+00:00",
		},
		"started_at_until": {
			"2017-07-04T00:03:01+00:00",
		},
	}, nil, nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	if rl.Total != expectedTotal {
		t.Errorf("Expected total to be %v but was %v", expectedTotal, rl.Total)
	}

	r := rl.Runs[0]
	if r.RunID != expectedRun {
		t.Errorf("Got unexpected run: %s", r.RunID)
	}
}

func TestSQLStateManager_ListRuns3(t *testing.T) {
	defer tearDown()
	sm := setUp()

	var err error
	expectedTotal := 2
	expectedRuns := map[string]bool{"run3": true, "run5": true}
	rl, err := sm.ListRuns(100, 0, "started_at", "asc", map[string][]string{
		"status": {
			StatusPending,
			StatusQueued,
		},
	}, nil, nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	if rl.Total != expectedTotal {
		t.Errorf("Expected total to be %v but was %v", expectedTotal, rl.Total)
	}

	for _, r := range rl.Runs {
		if _, ok := expectedRuns[r.RunID]; !ok {
			t.Errorf("Got unexpected run: %s", r.RunID)
		}
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
		t.Errorf("Expected environment to have exactly one entry, but was %v", len(*r2.Env))
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
		Alias:        "cute",
		Image:        "someImage",
		DefinitionID: "A",
		ClusterName:  "clusta",
		Status:       StatusQueued,
		Env: &EnvList{
			{Name: "RUN_PARAM", Value: "VAL"},
		},
		Engine: &DefaultEngine,
	}

	ec := int64(137)
	reason := "instance is ded."
	cmd := "_test cmd__"

	mem := int64(10)
	t1, _ := time.Parse(time.RFC3339, "2017-07-04T00:01:00+00:00")
	t2, _ := time.Parse(time.RFC3339, "2017-07-04T00:02:00+00:00")
	t1 = t1.UTC()
	t2 = t2.UTC()
	r2 := Run{
		RunID:        "run:18",
		GroupName:    "group:cupcake",
		DefinitionID: "A",
		Alias:        "AliasA",
		Image:        "ImageA",
		ExitCode:     &ec,
		ExitReason:   &reason,
		StartedAt:    &t1,
		FinishedAt:   &t2,
		ClusterName:  "clusta",
		Status:       StatusStopped,
		Env: &EnvList{
			{Name: "RUN_PARAM", Value: "VAL"},
		},
		Command: &cmd,
		Memory:  &mem,
		Engine:  &DefaultEngine,
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

	if *f2.ExitReason != *r2.ExitReason {
		t.Errorf("Expected exit reason %s but was %s", *r2.ExitReason, *f2.ExitReason)
	}

	if (*f2.StartedAt).UTC().String() != (*r2.StartedAt).String() {
		t.Errorf("Expected started_at %s but was %s", *r2.StartedAt, *f2.StartedAt)
	}

	if (*f2.FinishedAt).UTC().String() != (*r2.FinishedAt).String() {
		t.Errorf("Expected finished_at %s but was %s", *r2.FinishedAt, *f2.FinishedAt)
	}

	if f2.Alias != r2.Alias {
		t.Errorf("Expected alias: [%s] but was [%s]", r2.Alias, f2.Alias)
	}

	if f2.Image != r2.Image {
		t.Errorf("Expected image: [%s] but was [%s]", r2.Image, f2.Image)
	}

	if f1.Command != nil {
		t.Errorf("Expected null command, but was [%s]", *f1.Command)
	}

	if f1.Memory != nil {
		t.Errorf("Expected null mem, but was [%d]", *f1.Memory)
	}

	if f2.Command == nil {
		t.Errorf("Expected non-null command, but was null")
	}

	if f2.Memory == nil {
		t.Errorf("Expected non-null memory, but was null")
	}

	if f2.Command != nil && *f2.Command != cmd {
		t.Errorf("Expected command [%s], but got [%s]", cmd, *f2.Command)
	}

	if f2.Memory != nil && *f2.Memory != mem {
		t.Errorf("Expected mem [%d], but got [%d]", mem, *f2.Memory)
	}

}

func TestSQLStateManager_UpdateRun(t *testing.T) {
	defer tearDown()
	sm := setUp()

	ec := int64(1)
	env := EnvList{
		{Name: "NEW1", Value: "NEWVAL1"},
		{Name: "NEW2", Value: "NEWVAL2"},
	}
	t1, _ := time.Parse(time.RFC3339, "2017-07-04T00:01:00+00:00")
	t2, _ := time.Parse(time.RFC3339, "2017-07-04T00:02:00+00:00")
	t1 = t1.UTC()
	t2 = t2.UTC()
	u := Run{
		Alias:      "alien",
		Image:      "imagine",
		ExitCode:   &ec,
		Status:     StatusStopped,
		StartedAt:  &t1,
		FinishedAt: &t2,
		Env:        &env,
	}
	u2 := Run{
		Status: StatusNeedsRetry,
	}
	_, e := sm.UpdateRun("run3", u)
	if e != nil {
		t.Errorf("Error while updating %v", e)
	}

	r, e := sm.GetRun("run3")

	if e != nil {
		t.Errorf("Error in GetRun %v", e)
	}
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

	if r.Alias != u.Alias {
		t.Errorf("Expected update to set alias: [%s] but was [%s]", u.Alias, r.Alias)
	}

	if r.Image != u.Image {
		t.Errorf("Expected update to set image: [%s] but was [%s]", u.Image, r.Image)
	}

	updatedEnv := *r.Env
	matches := 0
	for i := range updatedEnv {
		updatedVar := updatedEnv[i]
		for j := range env {
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

	sm.UpdateRun("run3", u2)
	r, _ = sm.GetRun("run3")
	if r.Status != u2.Status {
		t.Errorf("Expected to update status to %s but was %s", u2.Status, r.Status)
	}
}
