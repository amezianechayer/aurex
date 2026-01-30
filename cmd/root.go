package cmd

import (
	"fmt"
	"os"

	"aurex/api"
	"aurex/config"
	"aurex/ledger"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var root = &cobra.Command{
	Use: "aurex",
}

func Execute() {
	server := &cobra.Command{
		Use: "server",
	}

	server.AddCommand(&cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			app := fx.New(
				fx.Provide(
					config.GetConfig,
					ledger.NewLedger,
					api.NewHttpAPI,
				),
				fx.Invoke(func(lc fx.Lifecycle, h *api.HttpAPI) {
				}),
			)

			app.Run()
		},
	})

	conf := &cobra.Command{
		Use: "config",
	}

	conf.AddCommand(&cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			c := config.DefaultConfig()
			b := c.Serialize()
			os.WriteFile("aurex.config.json", []byte(b), 0644)
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
