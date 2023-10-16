package dbservice

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"rulesengine/eurekaservice"
	"rulesengine/metamodel"
	"rulesengine/properties"
)

func GetAllCaches() ([]string, error) {
	serviceUrl, err := eurekaservice.GetApplicationUrl(properties.GetProperty("db-service.service-name"))
	if err != nil {
		return nil, err
	}
	response, err := http.Get(serviceUrl + "/cache-registry")
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var responseObject []string
	err = json.Unmarshal(responseData, &responseObject)
	if err == nil {
		return responseObject, nil
	} else {
		return nil, err
	}
}

func GetCacheData(cacheName string) (*metamodel.MetaData, error) {
	serviceUrl, err := eurekaservice.GetApplicationUrl(properties.GetProperty("db-service.service-name"))
	if err != nil {
		return nil, err
	}
	response, err := http.Get(serviceUrl + "/cache-registry/" + cacheName)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var responseObject metamodel.MetaData
	err = json.Unmarshal(responseData, &responseObject)
	if err == nil {
		return &responseObject, nil
	} else {
		return nil, err
	}
}
