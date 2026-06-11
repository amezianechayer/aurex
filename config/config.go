package config

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

// ConfigInfo struct
type ConfigInfo struct {
	Server  string      `json:"server"`
	Version interface{} `json:"version"`
	Config  *Config     `json:"config"`
}

// Config struct
type Config struct {
	LedgerStorage *LedgerStorage `json:"storage"`
}

// LedgerStorage struct
type LedgerStorage struct {
	Driver  interface{} `json:"driver"`
	Ledgers interface{} `json:"ledgers"`
}

func Init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/root"
	}

	os.MkdirAll(path.Join(home, ".corren", "data"), 0700)

	viper.SetDefault("debug", false)
	viper.SetDefault("storage.driver", "sqlite")
	viper.SetDefault("storage.dir", path.Join(home, ".corren/data"))
	viper.SetDefault("storage.sqlite.db_name", "corren")
	viper.SetDefault("storage.postgres.conn_string", "postgresql://localhost/postgres")
	viper.SetDefault("server.http.bind_address", "localhost:3068")
	viper.SetDefault("ui.http.bind_address", "localhost:3078")
	viper.SetDefault("ledgers", []interface{}{"quickstart"})

	viper.SetConfigName("corren")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.corren")
	viper.AddConfigPath("/etc/corren")
	viper.ReadInConfig()

	viper.SetEnvPrefix("corren")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func Remember(ledger string) {
	ledgers := viper.Get("ledgers").([]interface{})

	for _, v := range ledgers {
		if ledger == v.(string) {
			return
		}
	}

	viper.Set("ledgers", append(ledgers, ledger))

	// Persist ONLY the ledgers list. Writing the whole global viper state
	// would leak env vars and runtime overrides (storage.dir, auth.enabled,
	// test settings…) into the user's config file — that bug has bitten:
	// a test run once rewrote storage.dir and the user "lost" their data.
	file := viper.New()
	file.SetConfigName("corren")
	file.SetConfigType("yaml")
	file.AddConfigPath("$HOME/.corren")
	file.AddConfigPath("/etc/corren")

	if err := file.ReadInConfig(); err != nil {
		log.Printf(
			"failed to read config: ledger %s will not be remembered\n",
			ledger,
		)
		return
	}

	existing := []string{}
	for _, v := range file.GetStringSlice("ledgers") {
		if v == ledger {
			return
		}
		existing = append(existing, v)
	}
	file.Set("ledgers", append(existing, ledger))

	if err := file.WriteConfig(); err != nil {
		log.Printf(
			"failed to write config: ledger %s will not be remembered\n",
			ledger,
		)
	}
}
