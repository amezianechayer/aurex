package actions

import (
	"net/http"

	"github.com/amezianechayer/corren/config"
	_ "github.com/amezianechayer/corren/docs"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/swaggo/swag"
)

// ConfigController -
type ConfigController struct {
	BaseController
}

// NewConfigController -
func NewConfigController() ConfigController {
	return ConfigController{}
}

// GetInfo godoc
// @Summary Server Info
// @Description Show server informations
// @Tags server
// @Schemes
// @Accept json
// @Produce json
// @Success 200 {object} config.ConfigInfo{}
// @Router /_info [get]
func (ctl *ConfigController) GetInfo(c *gin.Context) {
	ctl.response(
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
	)
}
func (ctl *ConfigController) GetDocs(c *gin.Context) {
	doc, err := swag.ReadDoc("swagger")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Writer.Write([]byte(doc))
}
