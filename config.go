package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Prompt struct {
	System      string   `yaml:"system"`
	User        string   `yaml:"user"`
	Attachments []string `yaml:"attachments"`
}

type Config struct {
	Prompts map[string]Prompt `yaml:"prompts"`
}

func LoadConfig(filePath string) (Config, error) {
	var config Config
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func SaveConversation(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
