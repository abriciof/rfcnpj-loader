package email

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCheckConnection_RealSMTP(t *testing.T) {
	runIntegration := strings.TrimSpace(os.Getenv("RUN_INTEGRATION")) == "1" ||
		strings.TrimSpace(os.Getenv("RUN_SMTP_INTEGRATION")) == "1"
	if !runIntegration {
		t.Skip("set RUN_INTEGRATION=1 to run integration tests")
	}

	loadDotEnvForTest()

	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	portStr := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := strings.TrimSpace(os.Getenv("SMTP_PASS"))

	if host == "" || portStr == "" {
		t.Fatal("SMTP_HOST and SMTP_PORT are required")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("invalid SMTP_PORT=%q: %v", portStr, err)
	}

	cfg := SMTPConfig{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		To:   strings.TrimSpace(os.Getenv("MAIL_TO")),
	}

	requireAuth := strings.TrimSpace(os.Getenv("SMTP_REQUIRE_AUTH")) == "1" || (user != "" && pass != "")
	var errConn error
	if requireAuth {
		errConn = CheckConnectionRequireAuth(cfg)
	} else {
		errConn = CheckConnection(cfg)
	}

	if errConn != nil {
		t.Fatalf("real SMTP connection failed: %v", errConn)
	}
}

func loadDotEnvForTest() {
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
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		v = strings.Trim(v, `"'`)
		if k == "" {
			continue
		}
		if _, exists := os.LookupEnv(k); !exists {
			_ = os.Setenv(k, v)
		}
	}
}
