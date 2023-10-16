package properties

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var propertyFileName = ""

var data = make(map[string]interface{})
var debug bool

func LoadProperties() {
	if len(os.Args) > 1 {
		propertyFileName = os.Args[1]
	} else {
		propertyFileName = "application.yml"
	}

	propertyFile, errorFileRead := ioutil.ReadFile(propertyFileName)
	if errorFileRead != nil {
		log.Fatal(errorFileRead)
	}
	errorUnMarshallFile := yaml.Unmarshal(propertyFile, &data)
	if errorUnMarshallFile != nil {
		log.Fatal(errorUnMarshallFile)
	}
	log.Println("completed read properties: " + propertyFileName)
}

func GetProperty(propertyName string) string {
	propertyArray := strings.Split(propertyName, ".")
	return getProperty(propertyArray, data)
}

func getProperty(path []string, propertyMap interface{}) string {
	key := path[0]
	if reflect.TypeOf(propertyMap).Kind() == reflect.Map {
		if len(path) == 1 {
			return propertyMap.(map[string]interface{})[key].(string)
		} else {
			return getProperty(removeIndex(path, 0), propertyMap.(map[string]interface{})[key])
		}
	}
	return ""
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func ConvertToBoolean(s string) bool {
	res, err := strconv.ParseBool(s)
	if err != nil {
		log.Println(err.Error())
		return false
	} else {
		return res
	}
}

func IsDebugEnabled() bool {
	return debug
}
