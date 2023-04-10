package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	config *Config
)

type ServiceConfig struct {
	BaseURL  string `yaml:"base-url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type GroupMapping struct {
	Lldap     string `yaml:"lldap"`
	Sftpgo    string `yaml:"sftpgo"`
	GroupType int    `yaml:"group-type"`
}

type LldapConfig struct {
	RequiredGroup string `yaml:"required-group"`
	ServiceConfig
}

type SftpgoConfig struct {
	DefaultPrimaryGroup string `yaml:"default-primary-group"`
	ServiceConfig
}

type Config struct {
	Lldap         LldapConfig    `yaml:"lldap"`
	Sftpgo        SftpgoConfig   `yaml:"sftpgo"`
	GroupMappings []GroupMapping `yaml:"group-mappings"`
}

func init() {
	configDir := getEnvWithDefault("CONFIG_DIR", ".")
	configContents, err := os.ReadFile(fmt.Sprintf("%s/config.yml", configDir))
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(configContents, &config)
	if err != nil {
		log.Fatal(err)
	}

	config.Lldap.RequiredGroup = getEnvWithDefault("LLDAP_REQUIRED_URL", config.Lldap.RequiredGroup)
	config.Lldap.BaseURL = getEnvWithDefault("LLDAP_URL", config.Lldap.BaseURL)
	config.Sftpgo.User = getEnvWithDefault("SFTPGO_ADMIN_USER", config.Sftpgo.User)
	config.Sftpgo.Password = getEnvWithDefault("SFTPGO_ADMIN_PASSWORD", config.Sftpgo.Password)
	config.Sftpgo.BaseURL = getEnvWithDefault("SFTPGO_URL", config.Sftpgo.BaseURL)
}

func getEnvWithDefault(envName string, defaultValue string) string {
	val := os.Getenv(envName)
	if val != "" {
		return val
	}
	return defaultValue
}
