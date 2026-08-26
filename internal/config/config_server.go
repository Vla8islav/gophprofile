// Package config parsing passed config values
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// OptionsServer configuration parameters for the gophprofile server.
//
// Values' order of precedence: environment vars, command-line flags, config file, defaults.
type OptionsServer struct {
	ServerAddress    OptionalString `env:"RUN_ADDRESS" json:"server_address" command_arg:"a"`
	DatabaseURI      OptionalString `env:"DATABASE_URI" json:"database_uri" command_arg:"d"`
	MigrationsFolder OptionalString `env:"MIGRATIONS_FOLDER" json:"migrations_folder" command_arg:"m"`
	AuthTokenSecret  OptionalString `env:"AUTH_TOKEN_SECRET" json:"auth_token_secret" command_arg:"s"`
	PublicCertKey    OptionalString `env:"PUBLIC_CERT_KEY" json:"public_cert_key" command_arg:"public-key"`
	PrivateKey       OptionalString `env:"PRIVATE_KEY" json:"private_key" command_arg:"private-key"`
	AuditLogPath     OptionalString `env:"AUDIT_LOG_PATH" json:"audit_log_path" command_arg:"audit-log"`
	S3Endpoint       OptionalString `env:"S3_ENDPOINT" json:"s3_endpoint" command_arg:"s3-endpoint"`
	S3AccessKey      OptionalString `env:"S3_ACCESS_KEY" json:"s3_access_key" command_arg:"s3-access-key"`
	S3SecretKey      OptionalString `env:"S3_SECRET_KEY" json:"s3_secret_key" command_arg:"s3-secret-key"`
	S3Bucket         OptionalString `env:"S3_BUCKET" json:"s3_bucket" command_arg:"s3-bucket"`
	S3UseSSL         OptionalBool   `env:"S3_USE_SSL" json:"s3_use_ssl" command_arg:"s3-use-ssl"`
	Config           OptionalString `env:"CONFIG" json:"-" command_arg:"config"`
}

// ReadFlagsServer  Precedence: environment variables, command-line flags, config file, defaults.
func ReadFlagsServer(args []string, logger *zap.Logger) (*OptionsServer, error) {
	if logger == nil {
		panic("config server logger is nil")
	}
	cmdOptions, err := getOptionsServer(args)
	if err != nil {
		return nil, fmt.Errorf("read command-line flags: %w", err)
	}
	logSetFlags(cmdOptions, logger)
	envOptions, err := getEnvOptionsServer()
	if err != nil {
		return nil, fmt.Errorf("read environment variables: %w", err)
	}
	logSetEnv(envOptions, logger)
	var diskConfigOptions OptionsServer
	if cmdOptions.Config.BeenSet || envOptions.Config.BeenSet {
		// We need to read the config file before assembling the full consensus.
		var configFilename string
		if envOptions.Config.BeenSet && envOptions.Config.Value != "" {
			configFilename = envOptions.Config.Value
		} else if cmdOptions.Config.BeenSet && cmdOptions.Config.Value != "" {
			configFilename = cmdOptions.Config.Value
		}
		diskConfigOptions, err = getDiskConfigOptionsServer(configFilename)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		logConfigOptions(&diskConfigOptions, logger)
	}
	finalOptions := OptionsServer{
		ServerAddress: OptionalString{
			Value:   "localhost:8080",
			BeenSet: false,
		},
		DatabaseURI: OptionalString{
			Value:   "postgres://default_user:default_password@localhost:5432/gophprofile_db?sslmode=disable",
			BeenSet: false,
		},
		MigrationsFolder: OptionalString{
			Value:   "./migrations",
			BeenSet: false,
		},
		AuthTokenSecret: OptionalString{
			Value:   "super-duper-secret-dev-change-in-prod",
			BeenSet: false,
		},
		PublicCertKey: OptionalString{
			Value:   "",
			BeenSet: false,
		},
		PrivateKey: OptionalString{
			Value:   "",
			BeenSet: false,
		},
		AuditLogPath: OptionalString{
			Value:   "", // empty = audit disabled
			BeenSet: false,
		},
		S3Endpoint: OptionalString{
			Value:   "localhost:9000",
			BeenSet: false,
		},
		S3AccessKey: OptionalString{
			Value:   "minioadmin",
			BeenSet: false,
		},
		S3SecretKey: OptionalString{
			Value:   "minioadmin",
			BeenSet: false,
		},
		S3Bucket: OptionalString{
			Value:   "avatars",
			BeenSet: false,
		},
		S3UseSSL: OptionalBool{
			Value:   false,
			BeenSet: false,
		},
		Config: OptionalString{
			Value:   "",
			BeenSet: false,
		},
	}

	mergeOptions(&finalOptions, diskConfigOptions)
	mergeOptions(&finalOptions, *cmdOptions)
	mergeOptions(&finalOptions, *envOptions)
	return &finalOptions, nil
}

func getDiskConfigOptionsServer(filename string) (OptionsServer, error) {
	if filename == "" {
		return OptionsServer{}, nil
	}
	configBytes, err := os.ReadFile(filename)
	if err != nil {
		return OptionsServer{}, err
	}
	options := OptionsServer{}
	if err = json.Unmarshal(configBytes, &options); err != nil {
		return OptionsServer{}, err
	}
	return options, nil
}

func getEnvOptionsServer() (*OptionsServer, error) {
	opt := OptionsServer{}
	if err := env.Parse(&opt); err != nil {
		return nil, err
	}
	return &opt, nil
}
func getOptionsServer(args []string) (*OptionsServer, error) {
	opt := &OptionsServer{}
	fs := flag.NewFlagSet("gophprofile-server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&opt.ServerAddress, "a", "адрес и порт запуска этого сервера")
	fs.Var(&opt.DatabaseURI, "d", "connection string/dsn для postgres базы данных")
	fs.Var(&opt.MigrationsFolder, "m", "относительный путь до миграций, например ./migrations")
	fs.Var(&opt.AuthTokenSecret, "s", "секретный ключ для генерации токенов авторизации")
	fs.Var(&opt.PublicCertKey, "public-key", "путь до публичного ключа")
	fs.Var(&opt.PrivateKey, "private-key", "путь до приватного ключа")
	fs.Var(&opt.AuditLogPath, "audit-log", "путь до файла аудита (JSONL); пусто = выключено")
	fs.Var(&opt.S3Endpoint, "s3-endpoint", "адрес S3-совместимого хранилища (minio), host:port")
	fs.Var(&opt.S3AccessKey, "s3-access-key", "access key для S3")
	fs.Var(&opt.S3SecretKey, "s3-secret-key", "secret key для S3")
	fs.Var(&opt.S3Bucket, "s3-bucket", "имя S3-бакета для аватарок")
	fs.Var(&opt.S3UseSSL, "s3-use-ssl", "использовать https при обращении к S3")
	fs.Var(&opt.Config, "config", "путь до файла с конфигурацией приложения")
	fs.Var(&opt.Config, "c", "путь до файла с конфигурацией приложения")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opt, nil
}
