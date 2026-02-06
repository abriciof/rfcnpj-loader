package dav

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListZips_RealIntegration(t *testing.T) {
	runIntegration := strings.TrimSpace(os.Getenv("RUN_INTEGRATION")) == "1" ||
		strings.TrimSpace(os.Getenv("RUN_DAV_INTEGRATION")) == "1"
	if !runIntegration {
		t.Skip("set RUN_INTEGRATION=1 to run integration tests")
	}

	loadDotEnvForDAVTest()

	listURL := strings.TrimSpace(os.Getenv("DAV_TEST_URL"))
	if listURL == "" {
		tpl := strings.TrimSpace(os.Getenv("DAV_LIST_URL_TEMPLATE"))
		month := strings.TrimSpace(os.Getenv("DAV_TEST_MONTH"))
		if month == "" {
			month = strings.TrimSpace(os.Getenv("FORCE_MONTH"))
		}
		if month == "" {
			month = strings.TrimSpace(os.Getenv("START_MONTH"))
		}

		if tpl == "" || month == "" {
			t.Fatal("set DAV_TEST_URL or DAV_LIST_URL_TEMPLATE with DAV_TEST_MONTH/FORCE_MONTH/START_MONTH")
		}
		listURL = fmt.Sprintf(tpl, month)
	}

	client := NewClient()
	items, err := client.ListZips(context.Background(), listURL)
	if err != nil {
		t.Fatalf("ListZips real integration failed for %s: %v", listURL, err)
	}

	if len(items) == 0 {
		t.Fatalf("no zip files found at %s", listURL)
	}

	t.Logf("DAV URL: %s", listURL)
	t.Logf("zip files found: %d", len(items))
}

func loadDotEnvForDAVTest() {
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
