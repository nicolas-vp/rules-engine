package scriptservice

import (
	"fmt"
	goja "github.com/dop251/goja"
	"rulesengine/cacheservice"
)

func RunScriptValue(programCode string, context interface{}) (goja.Value, error) {
	vm := goja.New()

	//programCode = "var par = JSON.parse(par); \n" + programCode
	vm.Set("par", context)
	vm.Set("logEvent", func(decision string, rule string) {
		println("Decision:" + decision + " rule:" + rule)
	})
	vm.Set("print", func(x string) { fmt.Println(x) })
	vm.Set("cache", func(cacheName string, key string) string {
		return cacheservice.GetValueString(cacheName, key)
	})
	var value, err = vm.RunString(programCode)
	if err != nil {
		return nil, err
	}
	//value := vm.
	//if err != nil {
	//		log.Print("ERROR Getting variable: " + err.Error())
	//}
	return value, nil
}
