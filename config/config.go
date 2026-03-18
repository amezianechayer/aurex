package config

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

func Init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/root"
	}

	os.MkdirAll(path.Join(home, ".aurex", "data"), 0700)

	viper.SetDefault("debug", false)
	viper.SetDefault("storage.driver", "sqlite")
	viper.SetDefault("storage.dir", path.Join(home, ".aurex/data"))
	viper.SetDefault("storage.sqlite.db_name", "aurex")
	viper.SetDefault("storage.postgres.conn_string", "postgresql://localhost/postgres")
	viper.SetDefault("server.http.bind_address", "localhost:3068")

	viper.SetConfigName("aurex")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.aurex")
	viper.AddConfigPath("/etc/aurex")

	viper.SetEnvPrefix("aurex")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
