package cacheservice

import (
	"context"
	"encoding/json"
	groupcache "github.com/mailgun/groupcache/v2"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"rulesengine/clickhouseservice"
	"rulesengine/dbservice"
	"rulesengine/metamodel"
	"strings"
	"time"
)

var caches = make(map[string]*groupcache.Group)
var ctx = context.Background()

func StartCache(nodes string, host string) {
	p := strings.Split(nodes, ",")
	pool := groupcache.NewHTTPPoolOpts(p[0], &groupcache.HTTPPoolOptions{Replicas: 5000})
	pool.Set(p...)

	server := http.Server{
		Addr:    host,
		Handler: pool,
	}

	go func() {
		log.Println("Кеш инициализирован на хосте: " + host)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	//defer server.Shutdown(context.Background())
}

func CreateNewCache(cacheName string, meta metamodel.MetaData) {
	clickhouseservice.AddNewCache(&meta)
	CreateCache(cacheName, meta)
}

func CreateCache(cacheName string, meta metamodel.MetaData) *groupcache.Group {
	metamodel.RegisterModel(cacheName, meta)

	var fillFunction = groupcache.GetterFunc(func(ctx context.Context, key string, dest groupcache.Sink) error {
		meta := GetModel(cacheName)
		if meta.DoPersist {
			result, err := clickhouseservice.GetData(meta, key)
			if err == nil {
				return dest.SetBytes(result, time.Now().Add(time.Second*time.Duration(meta.SecondsToLive)))
			} else {
				return err
			}
		}
		return nil
	})

	var group = groupcache.NewGroup(cacheName, meta.Bytes, fillFunction)
	caches[cacheName] = group
	return group
}

func AddValue(cacheName string, key string, data map[string]interface{}, meta *metamodel.MetaData) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Проблема сохранения хеша ", err)
	}
	getCache(cacheName).Set(ctx, key, dataBytes,
		time.Now().Add(time.Second*time.Duration(meta.SecondsToLive)), false)
	if meta.DoPersist {
		clickhouseservice.AddData(key, data, meta)
	}
}

func GetValue(cache string, key string) []byte {
	var b []byte
	err := getCache(cache).Get(ctx, key, groupcache.AllocatingByteSliceSink(&b))
	if err != nil {
		logrus.Error(err.Error())
		return nil
	} else {
		return b
	}
}

func GetValueString(cache string, key string) string {
	var value = GetValue(cache, key)
	if value == nil {
		return "{}"
	} else {
		return string(value)
	}
}

func LoadCache(cacheName string, meta *metamodel.MetaData) {
	CreateCache(cacheName, *meta)
}

func GetModel(cacheName string) metamodel.MetaData {
	result, found := metamodel.GetModel(cacheName)
	if found {
		return *result
	} else {
		meta, err := dbservice.GetCacheData(cacheName)
		if err != nil {
			log.Fatal(err)
		}
		metamodel.RegisterModel(cacheName, *meta)
		return *meta
	}
}

func getCache(cache string) *groupcache.Group {
	result, ok := caches[cache]
	if ok {
		return result
	} else {
		return CreateCache(cache, GetModel(cache))
	}
}
