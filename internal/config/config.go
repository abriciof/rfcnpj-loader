package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	OutputFilesPath    string
	ExtractedFilesPath string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	// WebDAV listing
	DavBaseDomain      string
	DavListURLTemplate string
	StartMonth         string
	ForceMonth         string

	EnableDownload bool
	EnableExtract  bool
	CreateIndexes  bool

	// what to load
	LoadEmpresa         bool
	LoadEstabelecimento bool
	LoadSocios          bool
	LoadSimples         bool
	LoadCnae            bool
	LoadMoti            bool
	LoadMunic           bool
	LoadNatju           bool
	LoadPais            bool
	LoadQuals           bool

	// parallelism
	DownloadWorkers int
	ExtractWorkers  int
	TableWorkers    int
	FileWorkers     int

	// email
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPass          string
	MailTo            string
	MailNotifyUpToDate bool

	LogLevel string
}

func Load() (Config, error) {
	cfg := Config{
		OutputFilesPath:    getenv("OUTPUT_FILES_PATH", "/data/output"),
		ExtractedFilesPath: getenv("EXTRACTED_FILES_PATH", "/data/extracted"),

		DBHost: getenv("DB_HOST", "localhost"),
		DBPort: getenv("DB_PORT", "5432"),
		DBUser: getenv("DB_USER", "postgres"),
		DBPass: getenv("DB_PASSWORD", "postgres"),
		DBName: getenv("DB_NAME", "rfcnpj"),

		DavBaseDomain:      getenv("DAV_BASE_DOMAIN", "https://arquivos.receitafederal.gov.br"),
		DavListURLTemplate: getenv("DAV_LIST_URL_TEMPLATE", ""),
		StartMonth:         getenv("START_MONTH", ""),
		ForceMonth:         getenv("FORCE_MONTH", ""),

		EnableDownload: getenvBool("ENABLE_DOWNLOAD", true),
		EnableExtract:  getenvBool("ENABLE_EXTRACT", true),
		CreateIndexes:  getenvBool("CREATE_INDEXES", false),

		LoadEmpresa:         getenvBool("LOAD_EMPRESA", false),
		LoadEstabelecimento: getenvBool("LOAD_ESTABELECIMENTO", false),
		LoadSocios:          getenvBool("LOAD_SOCIOS", false),
		LoadSimples:         getenvBool("LOAD_SIMPLES", true),
		LoadCnae:            getenvBool("LOAD_CNAE", false),
		LoadMoti:            getenvBool("LOAD_MOTI", true),
		LoadMunic:           getenvBool("LOAD_MUNIC", false),
		LoadNatju:           getenvBool("LOAD_NATJU", false),
		LoadPais:            getenvBool("LOAD_PAIS", false),
		LoadQuals:           getenvBool("LOAD_QUALS", true),

		DownloadWorkers: getenvInt("DOWNLOAD_WORKERS", 4),
		ExtractWorkers:  getenvInt("EXTRACT_WORKERS", 2),
		TableWorkers:    getenvInt("TABLE_WORKERS", 2),
		FileWorkers:     getenvInt("FILE_WORKERS", 2),

		SMTPHost: getenv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort: getenvInt("SMTP_PORT", 587),
		SMTPUser: getenv("SMTP_USER", ""),
		SMTPPass: getenv("SMTP_PASS", ""),
		MailTo:   getenv("MAIL_TO", ""),
		MailNotifyUpToDate: getenvBool("MAIL_NOTIFY_UPTODATE", false),
		LogLevel:           getenv("LOG_LEVEL", "info"),
	}

	if strings.TrimSpace(cfg.DavListURLTemplate) == "" {
		return Config{}, fmt.Errorf("DAV_LIST_URL_TEMPLATE não configurada")
	}

	return cfg, nil
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvBool(k string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(k)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
