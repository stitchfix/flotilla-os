package log

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type TestDomainSpecificEvent struct {
	Timestamp time.Time
	Message   struct {
		A int `json: "a`
		B int `json: "b"`
	}
}

func TestHTTPEventSink_Receive(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		content := r.Header.Get("Content-Type")
		if content != "application/json" {
			t.Errorf("Expected Content-Type to eq %s got %s", "application/json", content)
		}

		e := TestDomainSpecificEvent{}
		err := json.NewDecoder(r.Body).Decode(&e)

		if err != nil {
			t.Errorf("Expected body to deserialize properly but got error %s", err.Error())
		}
	}))

	httpSink := NewHTTPSink(testServer.URL, "/", "POST")
	httpSink.Receive("a", 1, "b", 2)

	err := httpSink.Receive(1, "noway")
	if err == nil {
		t.Errorf("Expected message construction to fail with non-string keys")
	}
}
