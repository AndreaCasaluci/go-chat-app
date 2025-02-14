package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	DbHost         string `mapstructure:"DB_HOST"`
	DbPort         string `mapstructure:"DB_PORT"`
	ServerPort     string `mapstructure:"SERVER_PORT"`
	DbUsername     string `mapstructure:"DB_USER"`
	DbName         string `mapstructure:"DB_NAME"`
	DbPassword     string `mapstructure:"DB_PASSWORD"`
	JwtSecretKey   string `mapstructure:"JWT_SECRET_KEY"`
	MinioAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
}

var AppConfig Config

func loadConfig() (config Config, err error) {
	envs := Config{}
	viper.SetConfigFile(".env")

	er := viper.ReadInConfig()
	if er != nil {
		log.Fatal("Can't find the file .env : ", er)
		return envs, er
	}

	er = viper.Unmarshal(&envs)
	if er != nil {
		log.Fatal("Environment can't be loaded: ", er)
		return envs, er
	}

	AppConfig = envs
	return envs, nil
}

func GetConfig() (config Config, err error) {
	if AppConfig.DbHost != "" {
		return AppConfig, nil
	}
	config, err = loadConfig()
	if err != nil {
		fmt.Println("cannot load config:", err)
		return AppConfig, err
	}
	return config, nil
}
