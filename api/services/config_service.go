package services

import (
	"github.com/amezianechayer/corren/api/entities"
	"github.com/spf13/viper"
)

// ConfigService -
type ConfigService struct {
}

// NewConfigService -
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// CreateConfigService -
func CreateConfigService() *ConfigService {
	return NewConfigService()
}

func (s *ConfigService) GetConfigs() *entities.Infos {
	return &entities.Infos{
		Server:  "corren-ledger",
		Version: viper.Get("version"),
		Config: &entities.Config{
			LedgerStorage: &entities.LedgerStorage{
				Driver:  viper.Get("storage.driver"),
				Ledgers: viper.Get("ledgers"),
			},
		},
	}
}
