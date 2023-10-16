package main

import (
	"log"
	"rulesengine/cacheservice"
	"rulesengine/clickhouseservice"
	"rulesengine/controller"
	dbservice "rulesengine/dbservice"
	"rulesengine/eurekaservice"
	"rulesengine/properties"
	"rulesengine/sourceservice"
)

func main() {
	properties.LoadProperties()
	eurekaservice.InitEureka()
	clickhouseservice.Connect()
	cacheservice.StartCache(properties.GetProperty("cache.nodes"),
		properties.GetProperty("cache.host"))

	caches, err := dbservice.GetAllCaches()
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range caches {
		model, err := dbservice.GetCacheData(s)
		if err != nil {
			log.Fatal(err)
		}
		cacheservice.LoadCache(s, model)
		log.Printf("Кеш проиницилизирован: " + s)
	}

	ruleSets, err := sourceservice.GetAllRuleSets()
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range ruleSets {
		err := sourceservice.Compile(s)
		if err != nil {
			log.Fatal("Не удалось скомпилировать набор правил:"+s, err)
		}
	}

	controller.InitController()
}
