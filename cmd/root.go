package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/amezianechayer/corren-vm/script/compiler"
	"github.com/amezianechayer/corren/api"
	"github.com/amezianechayer/corren/api/actions"
	"github.com/amezianechayer/corren/config"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/storage"
	"github.com/amezianechayer/corren/storage/postgres"
	"github.com/amezianechayer/corren/storage/sqlite"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var (
	Version   = "develop"
	BuildDate = "-"
	Commit    = "-"

	root = &cobra.Command{
		Use:               "corren",
		Short:             "Corren",
		DisableAutoGenTag: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			config.Init()
			switch viper.GetString("storage.driver") {
			case "sqlite":
				storage.RegisterDriver("sqlite", sqlite.NewDriver(
					viper.GetString("storage.dir"),
					viper.GetString("storage.sqlite.db_name"),
				))
			case "postgres":
				storage.RegisterDriver("postgres", postgres.NewDriver(
					viper.GetString("storage.postgres.conn_string"),
				))
			default:
				return fmt.Errorf("unknown storage driver %s", viper.GetString("storage.driver"))
			}
			return nil
		},
	}
)

func PrintVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Version: %s \n", Version)
	fmt.Printf("Date: %s \n", BuildDate)
	fmt.Printf("Commit: %s \n", Commit)
}

func Execute() {
	viper.SetDefault("version", Version)
	server := &cobra.Command{
		Use: "server",
	}
	version := &cobra.Command{
		Use:   "version",
		Short: "Get version",
		Run:   PrintVersion,
	}

	start := &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			config.Init()

			options := make([]interface{}, 0)
			options = append(options,
				fx.Annotate(func() string { return viper.GetString("version") }, fx.ResultTags(`name:"version"`)),
				fx.Annotate(func() string { return viper.GetString("storage.driver") }, fx.ResultTags(`name:"storageDriver"`)),
				fx.Annotate(func() actions.LedgerLister {
					return actions.LedgerListerFn(func() []string {
						// Ledgers are updated by function config.Remember
						// We have to resolve the list dynamically
						return viper.GetStringSlice("ledgers")
					})
				}, fx.ResultTags(`name:"ledgerLister"`)),
				fx.Annotate(func() string { return viper.GetString("server.http.basic_auth") }, fx.ResultTags(`name:"httpBasic"`)),
			)
			if viper.GetBool("storage.cache") {
				options = append(options, fx.Annotate(
					storage.NewDefaultFactory,
					fx.ParamTags(`name:"storageDriver"`),
					fx.ResultTags(`name:"underlyingStorage"`),
				))
				options = append(options, fx.Annotate(
					storage.NewCachedStorageFactory,
					fx.ParamTags(`name:"underlyingStorage"`),
					fx.As(new(storage.Factory)),
				))
				options = append(options, fx.Annotate(
					ledger.WithStorageFactory,
					fx.ResultTags(`group:"resolverOptions"`),
					fx.As(new(ledger.ResolverOption)),
				))
			} else {
				options = append(options, fx.Annotate(
					storage.NewDefaultFactory,
					fx.ParamTags(`name:"storageDriver"`),
				))
			}
			app := fx.New(
				fx.Provide(options...),
				fx.Provide(

					fx.Annotate(ledger.NewResolver, fx.ParamTags(`group:"resolverOptions"`)),
					api.NewAPI,
				),

				fx.Invoke(func(lc fx.Lifecycle, h *api.API, storageFactory storage.Factory) {
					go func() {
						err := http.ListenAndServe(viper.GetString("server.http.bind_address"), h)
						if err != nil {
							panic(err)
						}
					}()
					lc.Append(fx.Hook{
						OnStop: func(ctx context.Context) error {
							log.Println("closing storage factory")
							err := storageFactory.Close(ctx)
							if err != nil {
								return errors.Wrap(err, "closing storage factory")
							}
							return nil
						},
					})
				}),
				api.Module,
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

	store := &cobra.Command{
		Use: "storage",
	}
	store.AddCommand(&cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			config.Init()
			s, err := storage.GetStore(viper.GetString("storage.driver"), "default")
			if err != nil {
				log.Fatal(err)
			}
			err = s.Initialize(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		},
	})

	scriptRun := &cobra.Command{
		Use:  "run [ledger] [script]",
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			config.Init()
			b, err := ioutil.ReadFile(args[1])
			if err != nil {
				log.Fatal(err)
			}
			r := regexp.MustCompile(`^\n`)
			s := string(b)
			s = r.ReplaceAllString(s, "")
			b, err = json.Marshal(gin.H{
				"plain": string(s),
			})
			if err != nil {
				log.Fatal(err)
			}
			res, err := http.Post(
				fmt.Sprintf(
					"http://%s/%s/script",
					viper.Get("server.http.bind_address"),
					args[0],
				),
				"application/json",
				bytes.NewReader([]byte(b)),
			)
			if err != nil {
				log.Fatal(err)
			}
			b, err = ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			var result struct {
				Err string `json:"err,omitempty"`
				Ok  bool   `json:"ok"`
			}
			err = json.Unmarshal(b, &result)
			if err != nil {
				log.Fatal(err)
			}
			if result.Ok {
				fmt.Println("Script run successfully ✅")
			} else {
				log.Fatal(result.Err)
			}
		},
	}

	scriptCheck := &cobra.Command{
		Use:  "check [script]",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			b, err := ioutil.ReadFile(args[0])
			if err != nil {
				log.Fatal(err)
			}
			_, err = compiler.Compile(string(b))
			if err != nil {
				log.Fatal(err)
			} else {
				fmt.Println("Script is valid ✅")
			}
		},
	}

	root.AddCommand(server)
	root.AddCommand(conf)
	root.AddCommand(UICmd)
	root.AddCommand(store)
	root.AddCommand(scriptRun)
	root.AddCommand(scriptCheck)
	root.AddCommand(version)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
