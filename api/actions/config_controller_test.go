package actions

import (
	"reflect"
	"testing"

	"github.com/amezianechayer/corren/api/services"
)

func TestNewConfigController(t *testing.T) {
	newConfigController := NewConfigController(services.CreateConfigService())
	if reflect.TypeOf(newConfigController) != reflect.TypeOf(&ConfigController{}) {
		t.Errorf(
			"NewConfigController return type is '%s' should be '%s'",
			reflect.TypeOf(newConfigController),
			reflect.TypeOf(&ConfigController{}),
		)
	}
}
