package eurekaservice

import (
	"errors"
	"github.com/HikoQiu/go-eureka-client/eureka"
	"log"
	"rulesengine/properties"
	"strconv"
)

var api *eureka.EurekaServerApi

func InitEureka() error {
	eurekaUrl := properties.GetProperty("eureka.url")
	eureka.SetLogger(func(level int, format string, a ...interface{}) {
		if level == eureka.LevelError {
			log.Printf(format, a...)
		} else {
			if properties.GetProperty("debug") == "true" {
				log.Printf(format, a...)
			}
		}
	})
	config := eureka.GetDefaultEurekaClientConfig()
	config.UseDnsForFetchingServiceUrls = false
	config.Region = "region"
	config.AvailabilityZones = map[string]string{
		"region": "zone",
	}
	config.ServiceUrl = map[string]string{
		"zone": eurekaUrl,
	}
	c := eureka.DefaultClient.Config(config)

	port, _ := strconv.Atoi(properties.GetProperty("server.port"))
	c.Register("rules-engine", port).Run()

	var err error
	api, err = c.Api()
	if err != nil {
		return err
	}

	return nil
}

func GetApplicationUrl(serviceName string) (string, error) {
	instances, err := api.QueryAllInstanceByAppId(serviceName)
	if err != nil {
		return "", err
	}
	if len(instances) > 0 {
		return instances[0].HomePageUrl, nil
	} else {
		return "", errors.New("no instances")
	}
}
