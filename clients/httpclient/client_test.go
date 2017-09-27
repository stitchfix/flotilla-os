package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type Cupcake struct {
	Flavour   string
	Sprinkles bool
}

const cupcakeResponse = `{"flavour": "vomit", "sprinkles":  true}`

type MockExecutor struct {
	TryCount int // keep track of how many times 'Do' got called
}

func (me *MockExecutor) Do(req *http.Request, timeout time.Duration, entity interface{}) error {
	me.TryCount += 1
	if req.URL.Path == "/" {
		return HttpRetryableError{errors.New("bork")}
	} else {
		return errors.New("not found yo")
	}
}

func TestClientRetry(t *testing.T) {
	me := &MockExecutor{}
	retryCount := 2
	client := &Client{
		Host:       "nope",
		Timeout:    1 * time.Second,
		RetryCount: retryCount,
		Executor:   me,
	}

	client.Get("/", nil, &Cupcake{})
	if me.TryCount != retryCount+1 {
		t.Errorf("Expected to try request [%v] times but got [%v]", retryCount+1, me.TryCount)
	}

	me.TryCount = 0
	client.Get("/404", nil, &Cupcake{})
	if me.TryCount != 1 {
		t.Errorf("Expected to try request [%v] times but got [%v]", 1, me.TryCount)
	}
}

func TestClientDo(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET", "DELETE":
			if len(r.URL.RawQuery) > 0 {
				fmt.Fprintf(w, `{"flavour":"vague","sprinkles":false}`)
			} else {
				fmt.Fprintf(w, cupcakeResponse)
			}
		case "PUT", "POST":
			content := r.Header.Get("Content-Type")
			if content != "application/json" {
				t.Errorf("Expected Content-Type to eq %s got %s", "application/json", content)
			}

			c := Cupcake{}
			err := json.NewDecoder(r.Body).Decode(&c)
			if err != nil {
				t.Errorf("Expected body to deserialize but got error %s", err.Error())
			}
			fmt.Fprintf(w, cupcakeResponse)
		}

	}))

	cupcake := Cupcake{}

	client := &Client{
		Host:       testServer.URL,
		Timeout:    1 * time.Second,
		RetryCount: 1,
	}

	var err error
	var headers = map[string]string{
		"Content-Type": "application/json",
	}
	err = client.Get("/", nil, &cupcake)
	if err != nil {
		t.Errorf("Expected err to be nil got %s", err.Error())
	}

	if cupcake.Flavour != "vomit" {
		t.Errorf("Expected flavour to be 'vomit', got: %v", cupcake.Flavour)
	}
	if !cupcake.Sprinkles {
		t.Errorf("Expected sprinkles to be true, got: %v", cupcake.Sprinkles)
	}

	cupcake = Cupcake{}
	err = client.Get("/?some_rando_param=thing", nil, &cupcake)
	if err != nil {
		t.Errorf("Expected err to be nil got %s", err.Error())
	}

	if cupcake.Flavour != "vague" {
		t.Errorf("Expected flavour to be 'vague', got: %v", cupcake.Flavour)
	}
	if cupcake.Sprinkles {
		t.Errorf("Expected sprinkles to be false, got: %v", cupcake.Sprinkles)
	}

	cupcake = Cupcake{}
	err = client.Put("/", headers, &Cupcake{"vomit", true}, &cupcake)
	if err != nil {
		t.Errorf("Expected err to be nil got %s", err.Error())
	}

	if cupcake.Flavour != "vomit" {
		t.Errorf("Expected flavour to be 'vomit', got: %v", cupcake.Flavour)
	}
	if !cupcake.Sprinkles {
		t.Errorf("Expected sprinkles to be true, got: %v", cupcake.Sprinkles)
	}

	cupcake = Cupcake{}
	err = client.Post("/", headers, &Cupcake{"vomit", true}, &cupcake)
	if err != nil {
		t.Errorf("Expected err to be nil got %s", err.Error())
	}
	if cupcake.Flavour != "vomit" {
		t.Errorf("Expected flavour to be 'vomit', got: %v", cupcake.Flavour)
	}
	if !cupcake.Sprinkles {
		t.Errorf("Expected sprinkles to be true, got: %v", cupcake.Sprinkles)
	}

	cupcake = Cupcake{}
	err = client.Delete("/", nil, &cupcake)
	if err != nil {
		t.Errorf("Expected err to be nil got %s", err.Error())
	}
}
