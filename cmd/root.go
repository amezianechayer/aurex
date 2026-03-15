package cmd

import (
	"fmt"
	"os"

	"github.com/amezianechayer/aurex/api"
	"github.com/amezianechayer/aurex/config"
	"github.com/amezianechayer/aurex/ledger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var (
	FlagBindAddr string
)

var root = &cobra.Command{
	Use: "aurex",
}

func Execute() {
	server := &cobra.Command{
		Use: "server",
	}

	start := &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			app := fx.New(
				fx.Provide(
					config.GetConfig,
					ledger.NewResolver,
					api.NewHttpAPI,
				),
				fx.Invoke(func() {
					config.Init()
				}),
				fx.Invoke(func(lc fx.Lifecycle, h *api.HttpAPI) {
				}),
			)

			app.Run()
		},
	}

	server.AddCommand(start)

	conf := &cobra.Command{
		Use: "config",
	}

	conf.AddCommand(&cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			config.Init()
			err := viper.SafeWriteConfig()
			if err != nil {
				fmt.Println(err)
			}
		},
	})

	root.AddCommand(server)
	root.AddCommand(conf)
	root.AddCommand(UICmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
