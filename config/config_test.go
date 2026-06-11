package config

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// Remember must persist ONLY the ledgers list into the config file —
// never the full viper state (env vars, test overrides, runtime Sets).
// Regression test for the bug where running tests with CORREN_* env vars
// rewrote the user's storage.dir, db_name and auth.enabled.
func TestRememberOnlyTouchesLedgers(t *testing.T) {
	home, err := ioutil.TempDir("", "corren-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(home)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", oldHome)

	confDir := path.Join(home, ".corren")
	if err := os.MkdirAll(confDir, 0700); err != nil {
		t.Fatal(err)
	}
	original := "ledgers:\n- quickstart\nstorage:\n  dir: /home/user/.corren/data\n  driver: sqlite\n"
	confPath := path.Join(confDir, "corren.yaml")
	if err := ioutil.WriteFile(confPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	viper.Reset()
	defer viper.Reset()
	Init()

	// simulate runtime pollution: env-style overrides living in the
	// global viper state that must NOT leak into the file
	viper.Set("storage.dir", "/tmp/some-test-dir")
	viper.Set("storage.sqlite.db_name", "testdb")
	viper.Set("auth.enabled", true)

	Remember("newledger")

	b, err := ioutil.ReadFile(confPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)

	if !strings.Contains(content, "newledger") {
		t.Fatalf("expected newledger in config, got:\n%s", content)
	}
	if !strings.Contains(content, "quickstart") {
		t.Fatalf("existing ledgers must be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "/home/user/.corren/data") {
		t.Fatalf("storage.dir must not be rewritten, got:\n%s", content)
	}
	if strings.Contains(content, "/tmp/some-test-dir") {
		t.Fatalf("runtime storage.dir override leaked into config:\n%s", content)
	}
	if strings.Contains(content, "testdb") {
		t.Fatalf("runtime db_name override leaked into config:\n%s", content)
	}
	if strings.Contains(content, "auth") {
		t.Fatalf("runtime auth.enabled leaked into config:\n%s", content)
	}
}

func TestRememberIsIdempotent(t *testing.T) {
	home, err := ioutil.TempDir("", "corren-config-test2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(home)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", oldHome)

	confDir := path.Join(home, ".corren")
	os.MkdirAll(confDir, 0700)
	confPath := path.Join(confDir, "corren.yaml")
	ioutil.WriteFile(confPath, []byte("ledgers:\n- quickstart\n"), 0644)

	viper.Reset()
	defer viper.Reset()
	Init()

	Remember("dupledger")
	Remember("dupledger")

	b, _ := ioutil.ReadFile(confPath)
	if n := strings.Count(string(b), "dupledger"); n != 1 {
		t.Fatalf("expected dupledger once, found %d times:\n%s", n, string(b))
	}
}
