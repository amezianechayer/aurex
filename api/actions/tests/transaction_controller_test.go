package tests

import (
	"reflect"
	"testing"

	"github.com/amezianechayer/corren/api/actions"
)

func TestNewTransactionController(t *testing.T) {
	if reflect.TypeOf(actions.NewTransactionController()) != reflect.TypeOf(&actions.TransactionController{}) {
		t.Errorf(
			"%s return type is '%s' should be '%s'",
			t.Name(),
			reflect.TypeOf(actions.NewTransactionController()),
			reflect.TypeOf(&actions.TransactionController{}),
		)
	}
}
