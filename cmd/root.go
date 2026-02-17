package cmd

import (
	"fmt"
	"os"

	"github.com/amezianechayer/aurex/api"
	"github.com/amezianechayer/aurex/config"
	"github.com/amezianechayer/aurex/ledger"
	"github.com/spf13/cobra"
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
					func() *config.Overrides {
						v := config.Overrides{}

						if cmd.Flag("http-bind-addr").Value.String() != "" {
							v["http-bind-addr"] = cmd.Flag("http-bind-addr").Value.String()
						}

						return &v
					},
					config.GetConfig,
					ledger.NewLedger,
					api.NewHttpAPI,
				),
				fx.Invoke(func(lc fx.Lifecycle, h *api.HttpAPI) {
				}),
			)

			app.Run()
		},
	}

	start.Flags().StringVarP(
		&FlagBindAddr,
		"http-bind-addr",
		// no shorthand
		"",
		// no default
		"",
		"override http api bind address",
	)

	server.AddCommand(start)

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
