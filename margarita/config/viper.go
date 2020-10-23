package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// Load defaults configuration files
func Load() {
	setDefaults()

	viper.SetConfigName("overrides")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.Getenv("CONFIG_PATH"))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("no overrides file found, using default values")
		} else {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
}

func setDefaults() {
	var configMap = map[string]interface{}{
		"envoy.port": uint32(8443),
		"envoy.host": "0.0.0.0",

		"margarita.domain": "margarita.poolside.dev",
		"margarita.host":   "0.0.0.0",
		"margarita.port":   uint32(10010),

		"xds.port":    uint32(10020),
		"xds.cluster": "xds_cluster",
	}
	log.Printf("Defaults: %v", configMap)
	for key, value := range configMap {
		viper.SetDefault(key, value)
	}
}

// GetEnvoyHost the host envoy should listen on to receive requests from the outside
// Default to 0.0.0.0 to listen on all interfaces in the docker container
func GetEnvoyHost() string {
	return viper.GetString("envoy.host")
}

// GetEnvoyPort the port number envoy should listen on to receive requests from the outside
// Default to 8443 for https connections
func GetEnvoyPort() uint32 {
	return viper.GetUint32("envoy.port")
}

// GetMargaritaDomain the domain name margarita should be configured with inside envoy to have
// a nice url to access it from
// Default is margarita.poolside.dev
func GetMargaritaDomain() string {
	return viper.GetString("margarita.domain")
}

// GetMargaritaHost the host the margarita service ui should listen on
// Default to 0.0.0.0 to listen on all interfaces in the docker container
func GetMargaritaHost() string {
	return viper.GetString("margarita.host")
}

// GetMargaritaPort the port number the margarita service ui should listen on
// Default to 10010
func GetMargaritaPort() uint32 {
	return viper.GetUint32("margarita.port")
}

// GetXdsCluster the cluster name given to envoy
// Default is xds_cluster
func GetXdsCluster() string {
	return viper.GetString("xds.cluster")
}

// GetXdsPort the grpc port that envoy will connect to for configurations
// Default is 10020
func GetXdsPort() uint32 {
	return viper.GetUint32("xds.port")
}
