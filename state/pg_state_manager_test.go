package state

import (
	"os"
	"testing"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "log"
)

func getDB() *sqlx.DB {
    db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    return db
}

func setUp() SQLStateManager {
    db := getDB()

    //
    // Implicit testing - this will create tables
    //
    sm := SQLStateManager{}
    sm.Initialize(os.Getenv("DATABASE_URL"))
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

    portsql := `
    INSERT INTO task_def_ports(task_def_id, port) VALUES ($1, $2)
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

    if len(dA.Env) != 1 {
        t.Errorf("Expected returned definitions to have correctly attached env vars, was %s", dA.Env)
    }

    if len(dA.Ports) != 1 {
        t.Errorf("Expected returned definitions to have correctly attached ports, was %s", dA.Ports)
    }

    // Test ordering and offset
    dl, _ = sm.ListDefinitions(1, 1, "group_name", "asc", nil, nil)
    if (dl.Definitions[0].GroupName != "groupW") {
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
    dl, _ = sm.ListDefinitions(1, 0, "alias", "asc", map[string]string{"image":"imageC",}, nil)
    if (dl.Definitions[0].Image != "imageC") {
        t.Errorf("Error filtering by field - expected imageC but got %s", dl.Definitions[0].Image)
    }

    // Test filtering on environment variables
    dl, _ = sm.ListDefinitions(1, 0, "alias", "desc", nil, map[string]string{"E_B1":"V_B1","E_B2":"V_B2"})
    if (dl.Definitions[0].DefinitionID != "B") {
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

    if len(dE.Env) != 0 {
        t.Errorf("Expected empty environment but got %s", dE.Env)
    }

    if len(dE.Ports) != 2 {
        t.Errorf("Expected 2 ports but got %s", dE.Ports)
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
    d := Definition{
        Arn: "arn:cupcake",
        DefinitionID: "id:cupcake",
        GroupName: "group:cupcake",
        ContainerName: "container:cupcake",
        User: "noone",
        Memory: 512,
        Alias: "cupcake",
        Image: "image:cupcake",
        Command: "echo 'hi'",
        Env: []EnvVar{
            {Name: "E1", Value:"V1"},
        },
        Ports: []int{12345, 6789},
    }

    sm.CreateDefinition(d)

    f, err := sm.GetDefinition("id:cupcake")
    if err != nil {
        t.Errorf("Expected create definition to create definition with id [id:cupcake]")
        t.Error(err)
    }

    if (f.Alias != d.Alias || len(f.Env) != len(d.Env) || len(f.Ports) != len(d.Ports)) {
        t.Errorf("Expected created definition to match the one passed in for creation")
    }
}