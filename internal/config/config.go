package config

import (
	"encoding/base64"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName string
	ServicePort string

	OTLPEndpoint            string
	OTLPProtocol            string
	OTLPInsecure            bool
	OTLPHeaders             map[string]string
	OTLPMetricsTemporality  string
	OTLPClientCert          string
	OTLPClientKey           string
	OTLPCACert              string

	GrafanaOTLPEndpoint string
	GrafanaAPIToken     string

	ProductServiceURL string
	OrderServiceURL   string

	ChaosErrorRoutes   map[string]float64
	ChaosLatencyRoutes map[string]time.Duration

	ChaosCPULoadEnabled bool
	ChaosCPULoadPercent int

	ChaosMemLoadEnabled bool
	ChaosMemLoadMB      int

	ChaosLogVolumeEnabled bool
	ChaosLogRatePerSec    int
	ChaosLogPattern       string

	ChaosDiskFillEnabled bool
	ChaosDiskFillPath    string
	ChaosDiskFillRateMB  int

	InventoryServiceURL string
	InventoryDataDir    string

	DatabaseURL string
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("OTEL_SERVICE_NAME", "unknown"),
		ServicePort: getEnv("SERVICE_PORT", "8080"),

		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		OTLPProtocol: getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc"),
		OTLPInsecure:           getBoolEnv("OTEL_EXPORTER_OTLP_INSECURE", true),
		OTLPHeaders:            buildOTLPHeaders(),
		OTLPMetricsTemporality: getEnv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY", "delta"),
		OTLPClientCert:         getEnv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE", ""),
		OTLPClientKey:          getEnv("OTEL_EXPORTER_OTLP_CLIENT_KEY", ""),
		OTLPCACert:             getEnv("OTEL_EXPORTER_OTLP_CERTIFICATE", ""),

		GrafanaOTLPEndpoint: getEnv("GRAFANA_OTLP_ENDPOINT", ""),
		GrafanaAPIToken:     getEnv("GRAFANA_API_TOKEN", ""),

		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8082"),

		ChaosErrorRoutes:   parseErrorRoutes(getEnv("CHAOS_ERROR_ROUTES", "")),
		ChaosLatencyRoutes: parseLatencyRoutes(getEnv("CHAOS_LATENCY_ROUTES", "")),

		ChaosCPULoadEnabled: getBoolEnv("CHAOS_CPU_LOAD_ENABLED", false),
		ChaosCPULoadPercent: getIntEnv("CHAOS_CPU_LOAD_PERCENT", 0),

		ChaosMemLoadEnabled: getBoolEnv("CHAOS_MEM_LOAD_ENABLED", false),
		ChaosMemLoadMB:      getIntEnv("CHAOS_MEM_LOAD_MB", 0),

		ChaosLogVolumeEnabled: getBoolEnv("CHAOS_LOG_VOLUME_ENABLED", false),
		ChaosLogRatePerSec:    getIntEnv("CHAOS_LOG_RATE_PER_SEC", 0),
		ChaosLogPattern:       getEnv("CHAOS_LOG_PATTERN", "mixed"),

		ChaosDiskFillEnabled: getBoolEnv("CHAOS_DISK_FILL_ENABLED", false),
		ChaosDiskFillPath:    getEnv("CHAOS_DISK_FILL_PATH", "/data"),
		ChaosDiskFillRateMB:  getIntEnv("CHAOS_DISK_FILL_RATE_MB", 1),

		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8084"),
		InventoryDataDir:    getEnv("INVENTORY_DATA_DIR", "/data"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

// buildOTLPHeaders constructs headers from OTEL_EXPORTER_OTLP_HEADERS and/or
// OTLP_BASIC_AUTH_USER + OTLP_BASIC_AUTH_PASSWORD. The basic auth pair takes
// precedence over an Authorization header in OTEL_EXPORTER_OTLP_HEADERS.
func buildOTLPHeaders() map[string]string {
	headers := parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))

	user := os.Getenv("OTLP_BASIC_AUTH_USER")
	pass := os.Getenv("OTLP_BASIC_AUTH_PASSWORD")
	if user != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
		headers["Authorization"] = "Basic " + encoded
	}
	return headers
}

// parseHeaders parses "key=value,key2=value2" (OTEL spec format).
func parseHeaders(s string) map[string]string {
	m := make(map[string]string)
	if s == "" {
		return m
	}
	for _, entry := range strings.Split(s, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), "=", 2)
		if len(parts) != 2 {
			continue
		}
		m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return m
}

// parseErrorRoutes parses "/path:0.1,/other:0.25" into a map.
func parseErrorRoutes(s string) map[string]float64 {
	m := make(map[string]float64)
	if s == "" {
		return m
	}
	for _, entry := range strings.Split(s, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
		if len(parts) != 2 {
			continue
		}
		rate, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		m[parts[0]] = rate
	}
	return m
}

// parseLatencyRoutes parses "/path:200ms,/other:1s" into a map.
func parseLatencyRoutes(s string) map[string]time.Duration {
	m := make(map[string]time.Duration)
	if s == "" {
		return m
	}
	for _, entry := range strings.Split(s, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
		if len(parts) != 2 {
			continue
		}
		d, err := time.ParseDuration(parts[1])
		if err != nil {
			continue
		}
		m[parts[0]] = d
	}
	return m
}
