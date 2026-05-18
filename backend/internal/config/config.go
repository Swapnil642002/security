package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv            string
	HTTPAddr          string
	DatabaseURL       string
	JWTSecret         string
	JWTIssuer         string
	JWTTTL            time.Duration
	BootstrapToken    string
	FirewallProvider  string
	FirewallDryRun    bool
	NFTablesBin       string
	OPNsenseBaseURL   string
	OPNsenseAPIKey    string
	OPNsenseAPISecret string
	AppBaseURL        string
}

func Load() (Config, error) {
	loadDotEnvIfPresent()

	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:         strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTIssuer:         getEnv("JWT_ISSUER", "firewall-manager"),
		BootstrapToken:    strings.TrimSpace(os.Getenv("BOOTSTRAP_TOKEN")),
		FirewallProvider:  getEnv("FIREWALL_PROVIDER", "nftables"),
		NFTablesBin:       getEnv("NFTABLES_BIN", "nft"),
		OPNsenseBaseURL:   strings.TrimSpace(os.Getenv("OPNSENSE_BASE_URL")),
		OPNsenseAPIKey:    strings.TrimSpace(os.Getenv("OPNSENSE_API_KEY")),
		OPNsenseAPISecret: strings.TrimSpace(os.Getenv("OPNSENSE_API_SECRET")),
		AppBaseURL:        getEnv("APP_BASE_URL", "http://localhost:8080"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	ttlMinutes := getEnv("JWT_TTL_MINUTES", "480")
	mins, err := strconv.Atoi(ttlMinutes)
	if err != nil || mins <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL_MINUTES must be a positive integer")
	}
	cfg.JWTTTL = time.Duration(mins) * time.Minute

	dryRunRaw := getEnv("FIREWALL_DRY_RUN", "true")
	dryRun, err := strconv.ParseBool(dryRunRaw)
	if err != nil {
		return Config{}, fmt.Errorf("FIREWALL_DRY_RUN must be true or false")
	}
	cfg.FirewallDryRun = dryRun

	return cfg, nil
}

func loadDotEnvIfPresent() {
	candidates := []string{".env", filepath.Join("backend", ".env")}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			_ = loadDotEnvFile(p)
			return
		}
	}
}

func loadDotEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
