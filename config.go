package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

type HostConfig struct {
	Host string
	Port string
}

type DbConfig struct {
	Http   *HostConfig
	DBName string
}

type ImageConfig struct {
	StoreWidth   int
	StoreHeight  int
	StoreQuality int
}

type Configuration struct {
	Http  *HostConfig
	Db    *DbConfig
	Image *ImageConfig
}

var (
	Config *Configuration
)

func init() {
	viper.SetConfigName("caaas")        // name of config file (without extension)
	viper.AddConfigPath("/etc/caaas/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.caaas") // call multiple times to add many search paths
	viper.AddConfigPath(".")            // optionally look for config in the working directory
	err := viper.ReadInConfig()         // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		log.Panicln("Can not read config file", err)
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	httpConfig := &HostConfig{viper.GetString("http.host"), viper.GetString("http.port")}
	imageConfig := &ImageConfig{viper.GetInt("image.storeWidth"), viper.GetInt("image.storeHeight"), viper.GetInt("image.storeQuality")}
	dbConfig := &DbConfig{&HostConfig{viper.GetString("db.host"), viper.GetString("db.port")}, viper.GetString("db.dbName")}

	Config = &Configuration{
		Http:  httpConfig,
		Db:    dbConfig,
		Image: imageConfig,
	}
}
