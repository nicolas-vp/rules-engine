package controller

import (
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"io/ioutil"
	"log"
	"net/http"
	"rulesengine/cacheservice"
	"rulesengine/metamodel"
	"rulesengine/properties"
	"rulesengine/scriptservice"
	"rulesengine/sourceservice"

	"github.com/gorilla/mux"
)

func InitController() {
	port := properties.GetProperty("server.port")
	router := mux.NewRouter()
	router.HandleFunc("/source/compile/{key}", compileHandler).Methods("POST")
	router.HandleFunc("/ruleset/run/{key}", runHandler).Methods("POST")

	router.HandleFunc("/data/{cache}/{key}", getHandler).Methods("GET")
	router.HandleFunc("/data/{cache}/{key}", addHandler).Methods("POST")
	router.HandleFunc("/data/{cache}", createHandler).Methods("POST")

	log.Println("server listen on port: " + port)

	err := http.ListenAndServe(":"+port, router)

	if err != nil {
		log.Fatal(err)
	}
}

func compileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	err := sourceservice.Compile(key)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
	} else {
		w.Write([]byte("200 - recompile success"))
	}
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	source, err := sourceservice.GetProgramSource(key)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var obj interface{}
	err = json.Unmarshal(body, &obj)

	if err != nil {
		log.Fatal(err)
	}

	var result goja.Value
	result, err = scriptservice.RunScriptValue(source, obj)

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Fprint(w, result.String())
	}

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	cache := vars["cache"]

	result := cacheservice.GetValue(cache, key)
	fmt.Fprint(w, string(result))
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	cache := vars["cache"]

	body, errorReadingFile := ioutil.ReadAll(r.Body)
	if errorReadingFile != nil {
		log.Println("error getting body from rest: " + errorReadingFile.Error())
	}
	var data map[string]interface{}
	marshallError := json.Unmarshal(body, &data)
	if marshallError != nil {
		log.Println("error parsing body from rest: " + marshallError.Error())
	}

	meta := cacheservice.GetModel(cache)

	cacheservice.AddValue(cache, key, data, &meta)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cache := vars["cache"]
	log.Println("create cache: " + cache)

	body, errorReadingFile := ioutil.ReadAll(r.Body)
	if errorReadingFile != nil {
		log.Println("error getting body from rest: " + errorReadingFile.Error())
	}
	var data metamodel.MetaData
	marshallError := json.Unmarshal(body, &data)
	if marshallError != nil {
		log.Println("error parsing body from rest: " + marshallError.Error())
	}
	cacheservice.CreateNewCache(cache, data)
}
