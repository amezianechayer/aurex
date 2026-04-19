package tests

import (
	"reflect"
	"testing"

	"github.com/amezianechayer/corren/api/actions"
	"github.com/amezianechayer/corren/api/services"
)

func TestNewConfigController(t *testing.T) {
	newConfigController := actions.NewConfigController(services.CreateConfigService())
	if reflect.TypeOf(newConfigController) != reflect.TypeOf(&actions.ConfigController{}) {
		t.Errorf(
			"%s return type is '%s' should be '%s'",
			t.Name(),
			reflect.TypeOf(newConfigController),
			reflect.TypeOf(&actions.ConfigController{}),
		)
	}
}
