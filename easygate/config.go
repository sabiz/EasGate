package easygate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	CONFIG_FILE_NAME = "easy_gate_config"
)

type Config struct {
	Proxy    Proxy  `mapstructure:"proxy"`
	Serve    Serve  `mapstructure:"serve"`
	LogLevel string `mapstructure:"log_level"`
}

type Serve struct {
	ListenPort  string `mapstructure:"listen_port"`
	PacFilePath string `mapstructure:"pac_file_path"`
}

type Proxy struct {
	Url      string `mapstructure:"url"`
	UserName string `mapstructure:"user_name"`
	Password string `mapstructure:"password"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName(CONFIG_FILE_NAME)
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(configPath, os.ModePerm)
	}
	viper.AddConfigPath(configPath)
	viper.SetDefault("proxy.url", "")
	viper.SetDefault("proxy.user_name", "")
	viper.SetDefault("proxy.password", "")
	viper.SetDefault("serve.listen_port", "44380")
	viper.SetDefault("serve.pac_file_path", "")
	viper.SetDefault("log_level", "DEBUG")

	viper.SafeWriteConfig()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("Can't read config: %s\n", err)
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("Can't load config: %s\n", err)
	}
	return &config, nil
}

func (config *Config) Save() {
	viper.Set("proxy.url", config.Proxy.Url)
	viper.Set("proxy.user_name", config.Proxy.UserName)
	viper.Set("proxy.password", config.Proxy.Password)
	viper.Set("serve.listen_port", config.Serve.ListenPort)
	viper.Set("serve.pac_file_path", config.Serve.PacFilePath)
	viper.Set("log_level", config.LogLevel)
	viper.WriteConfig()
}

func getConfigPath() string {
	home, err := os.UserConfigDir()
	if err != nil {
		exePath, err := os.Executable()
		if err != nil {
			panic("Can't get config path.")
		}
		return filepath.Dir(exePath)
	}
	return filepath.Join(home, "EasyGate")
}
