package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Api struct {
	PORT string `yaml:"port"`
	HOST string `yaml:"host"`
}

func (api *Api) GetAddr() string {
	return fmt.Sprintf("%s:%s", api.HOST, api.PORT)
}

type PostgresDSN string

func (p PostgresDSN) String() string {
	return string(p)
}

type ServiceConfiguration struct {
	Api         `yaml:"api"`
	PostgresDSN `yaml:"postgres_dsn"`
}

func Load() ServiceConfiguration {
	file, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}

	defer file.Close()
	decoder := yaml.NewDecoder(file)
	var cfg ServiceConfiguration
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func NewConfig() ServiceConfiguration {
	return Load()
}
