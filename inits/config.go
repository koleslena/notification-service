package inits

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type FlagConfig struct {
	LocalPort         string
	LogConfigFilePath string
	ConfigFilePath    string
}

func NewFlagConfig() (*FlagConfig, error) {
	flag.String("configpath", "./app.env", "file path to a config file")
	flag.String("logconfigpath", "./configs/config.yaml", "file path to a log config file")
	flag.String("localport", "", "port number for JSON RPC connection")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, errors.Errorf("error trying to parse flag configuration: %v", err)
	}

	return &FlagConfig{
		LocalPort:         viper.GetString("localport"),
		ConfigFilePath:    viper.GetString("configpath"),
		LogConfigFilePath: viper.GetString("logconfigpath"),
	}, nil
}

type Config struct {
	DBHost         string `mapstructure:"POSTGRES_HOST"`
	DBUserName     string `mapstructure:"POSTGRES_USER"`
	DBUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DBName         string `mapstructure:"POSTGRES_DB"`
	DBPort         string `mapstructure:"POSTGRES_PORT"`
	ServerPort     string `mapstructure:"PORT"`
	ClientOrigin   string `mapstructure:"CLIENT_ORIGIN"`
	SendingService string `mapstructure:"SENDING_SERVICE"`
	Token          string `mapstructure:"TOKEN"`
}

func InitConfig() (*Config, *LogConfig) {
	flags, err := NewFlagConfig()
	if err != nil {
		panic(err)
	}

	config := &Config{}
	err = loadConfig(flags.ConfigFilePath, config)
	if err != nil {
		log.Fatal("could not load logger config", err)
	}

	logConfig, err := loadLogConfig(flags.LogConfigFilePath)
	if err != nil {
		log.Fatal("could not load environment variables", err)
	}
	return config, logConfig
}

func loadConfig(path string, cfg interface{}) (err error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return errors.Errorf("error trying to parse configuration: %v", err)
	}

	err = viper.Unmarshal(cfg)
	return err
}

func loadLogConfig(path string) (*LogConfig, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return &LogConfig{}, errors.Errorf("error trying to parse configuration: %v", err)
	}

	return &LogConfig{
		IsConsoleEncoderEnabled: viper.GetBool("logging.log_console"),
		IsFileEncoderEnabled:    viper.GetBool("logging.log_file.enabled"),
		LogFilePath:             viper.GetString("logging.log_file.file_path"),
		Level:                   viper.GetString("logging.level"),
		Logger:                  viper.GetString("logging.logger"),
	}, nil
}

type LogConfig struct {
	IsConsoleEncoderEnabled bool   `yaml:"logging.log_console"`
	IsFileEncoderEnabled    bool   `yaml:"logging.log_file.enabled"`
	LogFilePath             string `yaml:"logging.log_file.file_path"`
	TimeEncoder             []byte `yaml:"logging.time_encoder"`
	Level                   string `yaml:"logging.level"`
	Logger                  string `yaml:"logging.logger"`
}

func (c Config) CreateRequestWithAuth(requestType string, id interface{}, data interface{}) (*http.Request, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%d", c.SendingService, id)
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.Token)
	req.Header.Add("Accept", "application/json")
	return req, nil
}
