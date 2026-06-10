package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/amezianechayer/corren-vm/script/compiler"
	"github.com/amezianechayer/corren/api"
	"github.com/amezianechayer/corren/auth"
	"github.com/amezianechayer/corren/config"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/scheduler"
	"github.com/amezianechayer/corren/storage"
	"github.com/gin-gonic/gin"
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
			app := fx.New(
				fx.Provide(
					ledger.NewResolver,
					api.NewAPI,
				),
				fx.Invoke(func() {
					config.Init()
				}),
				fx.Invoke(func(lc fx.Lifecycle, h *api.API) {
				}),
				api.Module,
				scheduler.Module,
			)
			app.Run()
		},
	}
	server.AddCommand(start)

	authCmd := &cobra.Command{Use: "auth"}
	authCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create the first admin user and api key (printed once)",
		Run: func(cmd *cobra.Command, args []string) {
			config.Init()
			svc, err := auth.NewServiceFromConfig()
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
			password := auth.GenerateToken("")[:16]
			if _, err := svc.CreateUser("admin", password, auth.RoleAdmin); err != nil {
				fmt.Println("error (already initialized?):", err)
				os.Exit(1)
			}
			key, _, err := svc.CreateKey("bootstrap-admin", auth.RoleAdmin, nil)
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
			fmt.Println("admin credentials created — shown ONCE, store them now:")
			fmt.Printf("  username: admin\n  password: %s\n  api key:  %s\n", password, key)
			fmt.Println("enable enforcement with auth.enabled: true in corren.yaml (or CORREN_AUTH_ENABLED=true)")
		},
	})
	root.AddCommand(authCmd)

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
			s, err := storage.GetStore("default")
			if err != nil {
				log.Fatal(err)
			}
			err = s.Initialize()
			if err != nil {
				log.Fatal(err)
			}
		},
	})

	script_run := &cobra.Command{
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

	script_check := &cobra.Command{
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
	root.AddCommand(script_run)
	root.AddCommand(script_check)
	root.AddCommand(version)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
