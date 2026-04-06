package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultAppName    = "GoAdmin"
	defaultAppEnv     = "dev"
	defaultAppVersion = "0.1.0"
)

type Config struct {
	App        AppConfig      `mapstructure:"app"`
	Server     ServerConfig   `mapstructure:"server"`
	Logger     LoggerConfig   `mapstructure:"logger"`
	Database   DatabaseConfig `mapstructure:"database"`
	CodeGen    CodeGenConfig  `mapstructure:"codegen"`
	Tenant     TenantConfig   `mapstructure:"tenant"`
	Auth       AuthConfig     `mapstructure:"auth"`
	LoadedAt   string         `mapstructure:"-"`
	LoadedFrom string         `mapstructure:"-"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Env     string `mapstructure:"env"`
	Version string `mapstructure:"version"`
}

type ServerConfig struct {
	HTTP HTTPServerConfig `mapstructure:"http"`
}

type HTTPServerConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	ReadTimeout     string `mapstructure:"read_timeout"`
	WriteTimeout    string `mapstructure:"write_timeout"`
	IdleTimeout     string `mapstructure:"idle_timeout"`
	ShutdownTimeout string `mapstructure:"shutdown_timeout"`
}

type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Format      string `mapstructure:"format"`
	Output      string `mapstructure:"output"`
	Development bool   `mapstructure:"development"`
}

type DatabaseConfig struct {
	Driver      string `mapstructure:"driver"`
	DSN         string `mapstructure:"dsn"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

type CodeGenConfig struct {
	Artifact CodeGenArtifactConfig `mapstructure:"artifact"`
}

type CodeGenArtifactConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	BaseDir string `mapstructure:"base_dir"`
	TTL     string `mapstructure:"ttl"`
}

type TenantConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type AuthConfig struct {
	JWT       JWTConfig       `mapstructure:"jwt"`
	Casbin    CasbinConfig    `mapstructure:"casbin"`
	Bootstrap BootstrapConfig `mapstructure:"bootstrap"`
}

type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	Issuer          string `mapstructure:"issuer"`
	Audience        string `mapstructure:"audience"`
	AccessTokenTTL  string `mapstructure:"access_token_ttl"`
	RefreshTokenTTL string `mapstructure:"refresh_token_ttl"`
}

type CasbinConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Source     string `mapstructure:"source"`
	ModelPath  string `mapstructure:"model_path"`
	PolicyPath string `mapstructure:"policy_path"`
}

type BootstrapConfig struct {
	Users []BootstrapUser `mapstructure:"users"`
}

type BootstrapUser struct {
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	PasswordHash string   `mapstructure:"password_hash"`
	TenantID     string   `mapstructure:"tenant_id"`
	DisplayName  string   `mapstructure:"display_name"`
	Roles        []string `mapstructure:"roles"`
	Permissions  []string `mapstructure:"permissions"`
}

func Default() Config {
	return Config{
		App: AppConfig{
			Name:    defaultAppName,
			Env:     defaultAppEnv,
			Version: defaultAppVersion,
		},
		Server: ServerConfig{
			HTTP: HTTPServerConfig{
				Host:            "0.0.0.0",
				Port:            8080,
				ReadTimeout:     "10s",
				WriteTimeout:    "10s",
				IdleTimeout:     "60s",
				ShutdownTimeout: "5s",
			},
		},
		Logger: LoggerConfig{
			Level:       "info",
			Format:      "json",
			Output:      "stdout",
			Development: false,
		},
		Database: DatabaseConfig{
			Driver:      "mysql",
			DSN:         "goadmin:goadmin@tcp(127.0.0.1:3306)/goadmin?charset=utf8mb4&parseTime=True&loc=Local",
			AutoMigrate: true,
		},
		CodeGen: CodeGenConfig{
			Artifact: CodeGenArtifactConfig{
				Enabled: true,
				BaseDir: filepath.Join(os.TempDir(), "goadmin", "codegen"),
				TTL:     "24h",
			},
		},
		Tenant: TenantConfig{
			Enabled: true,
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret:          "change-me-in-production",
				Issuer:          defaultAppName,
				Audience:        "goadmin-api",
				AccessTokenTTL:  "2h",
				RefreshTokenTTL: "168h",
			},
			Casbin: CasbinConfig{
				Enabled:    true,
				Source:     "file",
				ModelPath:  "core/auth/casbin/model/rbac.conf",
				PolicyPath: "core/auth/casbin/adapter/policy.csv",
			},
			Bootstrap: BootstrapConfig{},
		},
	}
}

func Load() (*Config, error) {
	cfg := Default()
	cfgDir, err := locateConfigDir()
	if err != nil {
		return nil, err
	}

	env := normalizedEnv()
	cfg.App.Env = env

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("GOADMIN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	applyDefaults(v, cfg)

	for _, name := range []string{"config.yaml", "config." + env + ".yaml", "local.yaml"} {
		if err := mergeFileIfExists(v, filepath.Join(cfgDir, name)); err != nil {
			return nil, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.App.Env = env
	if cfg.App.Name == "" {
		cfg.App.Name = defaultAppName
	}
	if cfg.App.Version == "" {
		cfg.App.Version = defaultAppVersion
	}
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.Format == "" {
		cfg.Logger.Format = "json"
	}
	if cfg.Logger.Output == "" {
		cfg.Logger.Output = "stdout"
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "mysql"
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "goadmin:goadmin@tcp(127.0.0.1:3306)/goadmin?charset=utf8mb4&parseTime=True&loc=Local"
	}
	if strings.TrimSpace(cfg.CodeGen.Artifact.BaseDir) == "" {
		cfg.CodeGen.Artifact.BaseDir = filepath.Join(os.TempDir(), "goadmin", "codegen")
	}
	if strings.TrimSpace(cfg.CodeGen.Artifact.TTL) == "" {
		cfg.CodeGen.Artifact.TTL = "24h"
	}
	if !cfg.Tenant.Enabled {
		cfg.Tenant.Enabled = false
	}
	if cfg.Auth.JWT.Secret == "" {
		cfg.Auth.JWT.Secret = "change-me-in-production"
	}
	if cfg.Auth.JWT.Issuer == "" {
		cfg.Auth.JWT.Issuer = defaultAppName
	}
	if cfg.Auth.JWT.Audience == "" {
		cfg.Auth.JWT.Audience = "goadmin-api"
	}
	if cfg.Auth.JWT.AccessTokenTTL == "" {
		cfg.Auth.JWT.AccessTokenTTL = "2h"
	}
	if strings.TrimSpace(cfg.Auth.JWT.RefreshTokenTTL) == "" {
		cfg.Auth.JWT.RefreshTokenTTL = "168h"
	}
	if strings.TrimSpace(cfg.Auth.Casbin.Source) == "" {
		cfg.Auth.Casbin.Source = "file"
	}
	if cfg.Auth.Casbin.ModelPath == "" {
		cfg.Auth.Casbin.ModelPath = "core/auth/casbin/model/rbac.conf"
	}
	if cfg.Auth.Casbin.PolicyPath == "" {
		cfg.Auth.Casbin.PolicyPath = "core/auth/casbin/adapter/policy.csv"
	}
	if cfg.Server.HTTP.Host == "" {
		cfg.Server.HTTP.Host = "0.0.0.0"
	}
	if cfg.Server.HTTP.Port == 0 {
		cfg.Server.HTTP.Port = 8080
	}
	if cfg.Server.HTTP.ReadTimeout == "" {
		cfg.Server.HTTP.ReadTimeout = "10s"
	}
	if cfg.Server.HTTP.WriteTimeout == "" {
		cfg.Server.HTTP.WriteTimeout = "10s"
	}
	if cfg.Server.HTTP.IdleTimeout == "" {
		cfg.Server.HTTP.IdleTimeout = "60s"
	}
	if cfg.Server.HTTP.ShutdownTimeout == "" {
		cfg.Server.HTTP.ShutdownTimeout = "5s"
	}
	if !cfg.Database.AutoMigrate {
		cfg.Database.AutoMigrate = true
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	cfg.LoadedAt = time.Now().UTC().Format(time.RFC3339)
	cfg.LoadedFrom = cfgDir
	return &cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return fmt.Errorf("app.name is required")
	}
	if c.Server.HTTP.Port <= 0 || c.Server.HTTP.Port > 65535 {
		return fmt.Errorf("server.http.port must be between 1 and 65535")
	}
	if _, _, _, _, err := c.Server.HTTP.Timeouts(); err != nil {
		return err
	}
	if strings.TrimSpace(c.Database.Driver) == "" {
		return fmt.Errorf("database.driver is required")
	}
	if strings.TrimSpace(c.Database.DSN) == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if strings.TrimSpace(c.Auth.JWT.Secret) == "" {
		return fmt.Errorf("auth.jwt.secret is required")
	}
	if _, _, err := c.Auth.JWT.Timeouts(); err != nil {
		return err
	}
	if c.Auth.Casbin.Enabled {
		source := strings.ToLower(strings.TrimSpace(c.Auth.Casbin.Source))
		if source == "" {
			source = "file"
		}
		switch source {
		case "file":
			if strings.TrimSpace(c.Auth.Casbin.ModelPath) == "" {
				return fmt.Errorf("auth.casbin.model_path is required when auth.casbin.source=file")
			}
			if strings.TrimSpace(c.Auth.Casbin.PolicyPath) == "" {
				return fmt.Errorf("auth.casbin.policy_path is required when auth.casbin.source=file")
			}
		case "db":
			// DB mode relies on the configured database connection and the built-in Casbin tables.
		default:
			return fmt.Errorf("auth.casbin.source must be file or db")
		}
	}
	if c.CodeGen.Artifact.Enabled {
		if strings.TrimSpace(c.CodeGen.Artifact.BaseDir) == "" {
			return fmt.Errorf("codegen.artifact.base_dir is required")
		}
		if _, err := c.CodeGen.Artifact.TTLDuration(); err != nil {
			return err
		}
	}
	return nil
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.HTTP.Host, c.Server.HTTP.Port)
}

func (c Config) Public() map[string]any {
	read, write, idle, shutdown, _ := c.Server.HTTP.Timeouts()
	return map[string]any{
		"app": map[string]any{
			"name":    c.App.Name,
			"env":     c.App.Env,
			"version": c.App.Version,
		},
		"server": map[string]any{
			"http": map[string]any{
				"host":             c.Server.HTTP.Host,
				"port":             c.Server.HTTP.Port,
				"read_timeout":     read.String(),
				"write_timeout":    write.String(),
				"idle_timeout":     idle.String(),
				"shutdown_timeout": shutdown.String(),
			},
		},
		"logger": map[string]any{
			"level":       c.Logger.Level,
			"format":      c.Logger.Format,
			"output":      c.Logger.Output,
			"development": c.Logger.Development,
		},
		"codegen": map[string]any{
			"artifact": map[string]any{
				"enabled":  c.CodeGen.Artifact.Enabled,
				"base_dir": c.CodeGen.Artifact.BaseDir,
				"ttl":      c.CodeGen.Artifact.TTL,
			},
		},
		"auth": map[string]any{
			"jwt": map[string]any{
				"issuer":            c.Auth.JWT.Issuer,
				"audience":          c.Auth.JWT.Audience,
				"access_token_ttl":  c.Auth.JWT.AccessTokenTTL,
				"refresh_token_ttl": c.Auth.JWT.RefreshTokenTTL,
			},
			"casbin": map[string]any{
				"enabled":     c.Auth.Casbin.Enabled,
				"source":      c.Auth.Casbin.Source,
				"model_path":  c.Auth.Casbin.ModelPath,
				"policy_path": c.Auth.Casbin.PolicyPath,
			},
			"bootstrap": map[string]any{
				"users": len(c.Auth.Bootstrap.Users),
			},
		},
		"loaded_at":   c.LoadedAt,
		"loaded_from": c.LoadedFrom,
	}
}

func (c JWTConfig) Timeouts() (time.Duration, time.Duration, error) {
	access, err := time.ParseDuration(strings.TrimSpace(c.AccessTokenTTL))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid auth.jwt.access_token_ttl: %w", err)
	}
	refresh, err := time.ParseDuration(strings.TrimSpace(c.RefreshTokenTTL))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid auth.jwt.refresh_token_ttl: %w", err)
	}
	return access, refresh, nil
}

func (c HTTPServerConfig) Timeouts() (time.Duration, time.Duration, time.Duration, time.Duration, error) {
	read, err := time.ParseDuration(strings.TrimSpace(c.ReadTimeout))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid server.http.read_timeout: %w", err)
	}
	write, err := time.ParseDuration(strings.TrimSpace(c.WriteTimeout))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid server.http.write_timeout: %w", err)
	}
	idle, err := time.ParseDuration(strings.TrimSpace(c.IdleTimeout))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid server.http.idle_timeout: %w", err)
	}
	shutdown, err := time.ParseDuration(strings.TrimSpace(c.ShutdownTimeout))
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid server.http.shutdown_timeout: %w", err)
	}
	return read, write, idle, shutdown, nil
}

func (c CodeGenArtifactConfig) TTLDuration() (time.Duration, error) {
	ttl, err := time.ParseDuration(strings.TrimSpace(c.TTL))
	if err != nil {
		return 0, fmt.Errorf("invalid codegen.artifact.ttl: %w", err)
	}
	if ttl <= 0 {
		return 0, fmt.Errorf("codegen.artifact.ttl must be greater than 0")
	}
	return ttl, nil
}

func normalizedEnv() string {
	for _, key := range []string{"GOADMIN_ENV", "APP_ENV"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return strings.ToLower(value)
		}
	}
	return defaultAppEnv
}

func applyDefaults(v *viper.Viper, cfg Config) {
	v.SetDefault("app.name", cfg.App.Name)
	v.SetDefault("app.env", cfg.App.Env)
	v.SetDefault("app.version", cfg.App.Version)
	v.SetDefault("server.http.host", cfg.Server.HTTP.Host)
	v.SetDefault("server.http.port", cfg.Server.HTTP.Port)
	v.SetDefault("server.http.read_timeout", cfg.Server.HTTP.ReadTimeout)
	v.SetDefault("server.http.write_timeout", cfg.Server.HTTP.WriteTimeout)
	v.SetDefault("server.http.idle_timeout", cfg.Server.HTTP.IdleTimeout)
	v.SetDefault("server.http.shutdown_timeout", cfg.Server.HTTP.ShutdownTimeout)
	v.SetDefault("logger.level", cfg.Logger.Level)
	v.SetDefault("logger.format", cfg.Logger.Format)
	v.SetDefault("logger.output", cfg.Logger.Output)
	v.SetDefault("logger.development", cfg.Logger.Development)
	v.SetDefault("database.driver", cfg.Database.Driver)
	v.SetDefault("database.dsn", cfg.Database.DSN)
	v.SetDefault("database.auto_migrate", cfg.Database.AutoMigrate)
	v.SetDefault("codegen.artifact.enabled", cfg.CodeGen.Artifact.Enabled)
	v.SetDefault("codegen.artifact.base_dir", cfg.CodeGen.Artifact.BaseDir)
	v.SetDefault("codegen.artifact.ttl", cfg.CodeGen.Artifact.TTL)
	v.SetDefault("tenant.enabled", cfg.Tenant.Enabled)
	v.SetDefault("auth.jwt.secret", cfg.Auth.JWT.Secret)
	v.SetDefault("auth.jwt.issuer", cfg.Auth.JWT.Issuer)
	v.SetDefault("auth.jwt.audience", cfg.Auth.JWT.Audience)
	v.SetDefault("auth.jwt.access_token_ttl", cfg.Auth.JWT.AccessTokenTTL)
	v.SetDefault("auth.jwt.refresh_token_ttl", cfg.Auth.JWT.RefreshTokenTTL)
	v.SetDefault("auth.casbin.enabled", cfg.Auth.Casbin.Enabled)
	v.SetDefault("auth.casbin.source", cfg.Auth.Casbin.Source)
	v.SetDefault("auth.casbin.model_path", cfg.Auth.Casbin.ModelPath)
	v.SetDefault("auth.casbin.policy_path", cfg.Auth.Casbin.PolicyPath)
}

func mergeFileIfExists(v *viper.Viper, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := v.MergeConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("merge config file %s: %w", path, err)
	}
	return nil
}

func locateConfigDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("GOADMIN_CONFIG_DIR")); dir != "" {
		if exists(filepath.Join(dir, "config.yaml")) {
			return dir, nil
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	searchRoots := []string{wd}
	parent := wd
	for i := 0; i < 4; i++ {
		parent = filepath.Dir(parent)
		if parent == searchRoots[len(searchRoots)-1] {
			break
		}
		searchRoots = append(searchRoots, parent)
	}

	for _, root := range searchRoots {
		for _, dir := range []string{"config", "configs"} {
			candidate := filepath.Clean(filepath.Join(root, dir))
			if exists(filepath.Join(candidate, "config.yaml")) {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("config directory not found; expected config/config.yaml")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
