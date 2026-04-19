package tests

import (
	"reflect"
	"testing"

	"github.com/amezianechayer/corren/api/actions"
)

func TestNewScriptController(t *testing.T) {
	if reflect.TypeOf(actions.NewScriptController()) != reflect.TypeOf(&actions.ScriptController{}) {
		t.Errorf(
			"%s return type is '%s' should be '%s'",
			t.Name(),
			reflect.TypeOf(actions.NewScriptController()),
			reflect.TypeOf(&actions.ScriptController{}),
		)
	}
}
