package actions

import (
	"net/http"

	"github.com/amezianechayer/corren/api/resources"
	"github.com/amezianechayer/corren/config"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ConfigController -
type ConfigController struct {
	BaseController
}

// NewConfigController -
func NewConfigController() *ConfigController {
	return &ConfigController{}
}

// GetInfo -
func (ctl *ConfigController) GetInfo(c *gin.Context) {
	ctl.responseResource(
		c,
		http.StatusOK,
		config.ConfigInfo{
			Server:  "corren-ledger",
			Version: viper.Get("version"),
			Config: &config.Config{
				LedgerStorage: &config.LedgerStorage{
					Driver:  viper.Get("storage.driver"),
					Ledgers: viper.Get("ledgers"),
				},
			},
		},
		&resources.Info{},
	)
}
