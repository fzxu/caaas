package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type HostConfig struct {
	Host string
	Port string
}

type DbConfig struct {
	Http     *HostConfig
	DBName   string
	NumConns int
	Timeout  int
}

type ImageConfig struct {
	StoreWidth    int
	StoreHeight   int
	DefaultWidth  int
	DefaultHeight int
	StoreQuality  int
	ReadQuality   int
	CacheDir      string
	ProcessPar    int
}

type Configuration struct {
	Http  *HostConfig
	Db    *DbConfig
	Image *ImageConfig
}

var (
	Config       *Configuration
	ImageChannel chan int
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
	imageConfig := &ImageConfig{
		StoreWidth:    viper.GetInt("image.storeWidth"),
		StoreHeight:   viper.GetInt("image.storeHeight"),
		DefaultWidth:  viper.GetInt("image.defaultWidth"),
		DefaultHeight: viper.GetInt("image.defaultHeight"),
		StoreQuality:  viper.GetInt("image.storeQuality"),
		ReadQuality:   viper.GetInt("image.readQuality"),
		CacheDir:      viper.GetString("image.cacheDir"),
		ProcessPar:    viper.GetInt("image.processPar"),
	}
	dbConfig := &DbConfig{&HostConfig{viper.GetString("db.host"), viper.GetString("db.port")},
		viper.GetString("db.dbName"), viper.GetInt("db.numConns"), viper.GetInt("db.timeout")}

	Config = &Configuration{
		Http:  httpConfig,
		Db:    dbConfig,
		Image: imageConfig,
	}
	err = os.MkdirAll(Config.Image.CacheDir, 0755)
	if err != nil {
		panic(err)
	}

	// semaphore for max number of image processor run in parallel
	ImageChannel = make(chan int, Config.Image.ProcessPar)
}
