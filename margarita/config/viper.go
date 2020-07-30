package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Load defaults configuration files
func Load() {
	setDefaults()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/app/")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("no config file found, using default values")
		} else {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
}

func setDefaults() {
	var configMap = map[string]interface{}{
		"envoy.port": uint32(443),
		"envoy.host": "127.0.0.1",

		"xds.port":    uint32(8080),
		"xds.cluster": "xds_cluster",
	}
	for key, value := range configMap {
		viper.SetDefault(key, value)
	}
}

// GetEnvoyPort port
func GetEnvoyPort() uint32 {
	return viper.GetUint32("envoy.port")
}

// GetEnvoyHost host
func GetEnvoyHost() string {
	return viper.GetString("envoy.host")
}

// GetXdsPort port
func GetXdsPort() uint32 {
	return viper.GetUint32("xds.port")
}

// GetXdsCluster cluster
func GetXdsCluster() string {
	return viper.GetString("xds.cluster")
}
