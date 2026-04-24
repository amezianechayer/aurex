package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/amezianechayer/corren/api"
	"github.com/amezianechayer/corren/config"
	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/storage"
	"github.com/spf13/cobra"
)

func openAdminStore() storage.Store {
	config.Init()
	s, err := storage.NewAdminStore()
	if err != nil {
		log.Fatal("cannot open admin store:", err)
	}
	return s
}

var APIKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "Manage API keys",
}

var createKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate a new API key",
	Run: func(cmd *cobra.Command, args []string) {
		name, _      := cmd.Flags().GetString("name")
		tierStr, _   := cmd.Flags().GetString("tier")
		role, _      := cmd.Flags().GetString("role")
		expiresAt, _ := cmd.Flags().GetString("expires-at")
		asJSON, _    := cmd.Flags().GetBool("json")

		if name == "" {
			fmt.Fprintln(os.Stderr, "error: --name is required")
			os.Exit(1)
		}

		tier := core.APIKeyTier(tierStr)
		switch tier {
		case core.TierSandbox, core.TierStarter, core.TierGrowth, core.TierEnterprise:
		default:
			fmt.Fprintf(os.Stderr, "error: invalid tier '%s' — must be sandbox|starter|growth|enterprise\n", tierStr)
			os.Exit(1)
		}

		raw := make([]byte, 24)
		if _, err := rand.Read(raw); err != nil {
			log.Fatal("entropy failure:", err)
		}
		plain := fmt.Sprintf("crn_%s_%s", tier, hex.EncodeToString(raw))
		hash := api.HashAPIKey(plain)

		k := core.APIKey{
			KeyHash:   hash,
			Name:      name,
			Role:      role,
			Tier:      tier,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			ExpiresAt: expiresAt,
		}

		s := openAdminStore()
		defer s.Close()

		if err := s.SaveAPIKey(k); err != nil {
			log.Fatal("failed to save key:", err)
		}

		limit := core.TierRateLimit[tier]
		limitLabel := fmt.Sprintf("%d req/h", limit)
		if limit == 0 {
			limitLabel = "unlimited"
		}

		if asJSON {
			out, _ := json.MarshalIndent(map[string]interface{}{
				"key":        plain,
				"name":       name,
				"tier":       tier,
				"role":       role,
				"rate_limit": limitLabel,
				"created_at": k.CreatedAt,
				"expires_at": expiresAt,
			}, "", "  ")
			fmt.Println(string(out))
			return
		}

		fmt.Println()
		fmt.Printf("  Key   : %s\n", plain)
		fmt.Printf("  Name  : %s\n", name)
		fmt.Printf("  Tier  : %s  (%s)\n", tier, limitLabel)
		fmt.Printf("  Role  : %s\n", role)
		if expiresAt != "" {
			fmt.Printf("  Expires: %s\n", expiresAt)
		}
		fmt.Println()
		fmt.Println("  ⚠  Save this key now — it will not be shown again.")
		fmt.Println()
	},
}

var listKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys",
	Run: func(cmd *cobra.Command, args []string) {
		s := openAdminStore()
		defer s.Close()

		keys, err := s.ListAPIKeys()
		if err != nil {
			log.Fatal(err)
		}

		if len(keys) == 0 {
			fmt.Println("No API keys found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tTIER\tROLE\tCREATED\tEXPIRES")
		fmt.Fprintln(w, "----\t----\t----\t-------\t-------")
		for _, k := range keys {
			exp := k.ExpiresAt
			if exp == "" {
				exp = "never"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", k.Name, k.Tier, k.Role, k.CreatedAt[:10], exp)
		}
		w.Flush()
	},
}

var revokeKeyCmd = &cobra.Command{
	Use:   "revoke <key>",
	Short: "Revoke an API key (pass the full crn_... key)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plain := args[0]
		hash := api.HashAPIKey(plain)

		s := openAdminStore()
		defer s.Close()

		if err := s.DeleteAPIKey(hash); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Key revoked.")
	},
}

func init() {
	createKeyCmd.Flags().String("name", "", "Label for the key (required)")
	createKeyCmd.Flags().String("tier", "sandbox", "Tier: sandbox|starter|growth|enterprise")
	createKeyCmd.Flags().String("role", "reader", "Role: reader|admin")
	createKeyCmd.Flags().String("expires-at", "", "Expiry date RFC3339, e.g. 2026-12-31T00:00:00Z")
	createKeyCmd.Flags().Bool("json", false, "Output as JSON")

	APIKeysCmd.AddCommand(createKeyCmd)
	APIKeysCmd.AddCommand(listKeysCmd)
	APIKeysCmd.AddCommand(revokeKeyCmd)
}
