package db

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenSQL_RealIntegration(t *testing.T) {
	runIntegration := strings.TrimSpace(os.Getenv("RUN_INTEGRATION")) == "1"
	if !runIntegration {
		t.Skip("set RUN_INTEGRATION=1 to run integration tests")
	}

	loadDotEnvForDBTest()

	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	user := strings.TrimSpace(os.Getenv("DB_USER"))
	pass := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	name := strings.TrimSpace(os.Getenv("DB_NAME"))

	if host == "" || port == "" || user == "" || pass == "" || name == "" {
		t.Fatal("DB_HOST, DB_PORT, DB_USER, DB_PASSWORD and DB_NAME are required")
	}

	conn, err := OpenSQL(context.Background(), host, port, user, pass, name)
	if err != nil {
		t.Fatalf("OpenSQL real integration failed: %v", err)
	}
	_ = conn.Close()
}

func loadDotEnvForDBTest() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	envPath := filepath.Clean(filepath.Join(wd, "..", "..", ".env"))
	f, err := os.Open(envPath)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}

