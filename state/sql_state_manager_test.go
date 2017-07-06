package state

import (
	"os"
	"testing"
)

func TestSQLStateManager_Initialize(t *testing.T) {
	sm := SQLStateManager{}
	sm.Initialize(os.Getenv("DATABASE_URL"))
	// TBD
}
