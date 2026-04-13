package tests

import (
	"reflect"
	"testing"

	"github.com/amezianechayer/corren/api/actions"
)

func TestNewBaseController(t *testing.T) {
	if reflect.TypeOf(actions.NewBaseController()) != reflect.TypeOf(&actions.BaseController{}) {
		t.Errorf(
			"%s return type is '%s' should be '%s'",
			t.Name(),
			reflect.TypeOf(actions.NewBaseController()),
			reflect.TypeOf(&actions.BaseController{}),
		)
	}
}
