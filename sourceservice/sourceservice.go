package sourceservice

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"rulesengine/cacheservice"
	"rulesengine/dmn"
	"rulesengine/eurekaservice"
	"rulesengine/metamodel"
	"rulesengine/properties"
)

type RuleSet struct {
	Name             string
	Description      string
	Modified         string
	ModifiedBy       string
	RuleSetContainer RuleSetContainer
}

type RuleSetContainer struct {
	RuleList      []Rule
	ParameterList []Parameter
	PreScript     string
	PostScript    string
}

type Rule struct {
	Uid               string
	Order             int32
	Name              string
	DmnFile           string
	DmnId             string
	Enabled           string
	RuleParameterList []Parameter
}

type Parameter struct {
	Name  string
	Value interface{}
	Type  string
}

type Process struct {
	WorkflowKey  int64
	ProcessId    string
	Resource     string
	ResourceName string
	Version      int32
	ResourceType string
	Created      string
}

type Source struct {
	Source string
}

var dmnServiceUrl string
var sourceModel *metamodel.MetaData

const NEW_LINE = "\n"
const PROGRAM_SOURCE = "program_source"

func Compile(ruleSetName string) error {
	log.Printf("Производится компиляция набора правил: " + ruleSetName)
	ruleSet, err := getModel(ruleSetName)
	if err != nil {
		return err
	}
	var code = ""
	for _, par := range ruleSet.RuleSetContainer.ParameterList {
		code += mapParameter(par)
	}
	code += ruleSet.RuleSetContainer.PreScript

	for _, rule := range ruleSet.RuleSetContainer.RuleList {
		if rule.Enabled == "on" && rule.DmnFile != "" {
			for _, par := range rule.RuleParameterList {
				code += mapParameter(par)
			}
			dmnSource, err := getDmn(rule.DmnFile)
			dmn.LoadDmn(dmnSource)
			dmn.Optimize(rule.DmnId)
			if err != nil {
				log.Fatal(err)
			}
			code += dmn.GetCode(rule.DmnId)
		}
	}

	code += ruleSet.RuleSetContainer.PostScript
	err = saveProgramSource(ruleSetName, code)
	log.Println(code)
	if err != nil {
		return err
	}
	return nil
}

func getModel(modelName string) (*RuleSet, error) {
	modelUrl, err := eurekaservice.GetApplicationUrl(properties.GetProperty("dmn-service.service-name"))
	if err != nil {
		return nil, err
	}
	response, err := http.Get(modelUrl + "/rule-set/" + modelName)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var responseObject RuleSet
	err = json.Unmarshal(responseData, &responseObject)
	if err == nil {
		return &responseObject, nil
	} else {
		return nil, err
	}
}

func GetAllRuleSets() ([]string, error) {
	modelUrl, err := eurekaservice.GetApplicationUrl(properties.GetProperty("dmn-service.service-name"))
	if err != nil {
		return nil, err
	}
	response, err := http.Get(modelUrl + "/rule-set")
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var responseObject []RuleSet
	err = json.Unmarshal(responseData, &responseObject)
	if err == nil {
		var result []string
		for _, o := range responseObject {
			result = append(result, o.Name)
		}
		return result, nil
	} else {
		return nil, err
	}
}

func getDmn(dmnFile string) (string, error) {
	modelUrl, err := eurekaservice.GetApplicationUrl(properties.GetProperty("dmn-service.service-name"))
	if err != nil {
		return "", err
	}
	response, err := http.Get(modelUrl + "/process/by-id/" + dmnFile)
	if err != nil {
		return "", err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var responseObject Process
	err = json.Unmarshal(responseData, &responseObject)
	if err == nil {
		return responseObject.Resource, nil
	} else {
		return "", err
	}
}

func mapParameter(par Parameter) string {
	var value = "var "
	value += par.Name
	switch par.Type {
	case "String", "Symbol":
		value = value + "=\"" + par.Value.(string) + "\""
	case "Boolean":
		value += "=" + par.Value.(string)
	case "Number", "BigInt":
		value += "=" + par.Value.(string)
	case "Undefined":
		value += ""
	}
	value += ";" + NEW_LINE
	return value
}

func saveProgramSource(ruleSet string, source string) error {
	var success bool
	if sourceModel == nil {
		sourceModel, success = metamodel.GetModel(PROGRAM_SOURCE)
		if !success {
			return errors.New("Проблемы с получением метамодели " + PROGRAM_SOURCE)
		}
	}

	var programSourceObject map[string]interface{}
	programSourceObject = make(map[string]interface{})
	programSourceObject["Source"] = source

	cacheservice.AddValue(PROGRAM_SOURCE, ruleSet, programSourceObject, sourceModel)
	return nil
}

func GetProgramSource(ruleSet string) (string, error) {
	result := cacheservice.GetValue(PROGRAM_SOURCE, ruleSet)
	var responseObject Source

	err := json.Unmarshal(result, &responseObject)

	if err == nil {
		return responseObject.Source, nil
	} else {
		return "", err
	}
}
