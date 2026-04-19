package tests

import (
	"reflect"
	"testing"

	"github.com/amezianechayer/corren/api/actions"
)

func TestNewAccountController(t *testing.T) {
	if reflect.TypeOf(actions.NewAccountController()) != reflect.TypeOf(&actions.AccountController{}) {
		t.Errorf(
			"%s return type is '%s' should be '%s'",
			t.Name(),
			reflect.TypeOf(actions.NewAccountController()),
			reflect.TypeOf(&actions.AccountController{}),
		)
	}
}
