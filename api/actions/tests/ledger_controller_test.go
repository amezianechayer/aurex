package tests

import (
	"reflect"
	"testing"

	"github.com/amezianechayer/corren/api/actions"
)

func TestNewLedgerController(t *testing.T) {
	if reflect.TypeOf(actions.NewLedgerController()) != reflect.TypeOf(&actions.LedgerController{}) {
		t.Errorf(
			"%s return type is '%s' should be '%s'",
			t.Name(),
			reflect.TypeOf(actions.NewLedgerController()),
			reflect.TypeOf(&actions.LedgerController{}),
		)
	}
}
