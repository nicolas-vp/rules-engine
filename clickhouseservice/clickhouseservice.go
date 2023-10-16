package clickhouseservice

import (
	"context"
	"encoding/json"
	"reflect"
	"rulesengine/metamodel"
	"rulesengine/properties"
	"strings"
	// "database/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
	"log"
	"time"
)

var connection clickhouse.Conn
var ctx context.Context

func Connect() {
	var host = properties.GetProperty("clickhouse.host")
	var database = properties.GetProperty("clickhouse.database")
	var username = properties.GetProperty("clickhouse.username")
	var password = properties.GetProperty("clickhouse.password")

	options := &clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			clickhouse.CompressionLZ4,
		},
		Debug: properties.IsDebugEnabled(),
	}
	var err error
	connection, err = clickhouse.Open(options)
	if err != nil {
		log.Println(err.Error())
	}
	ctx = clickhouse.Context(context.Background(), clickhouse.WithStdAsync(false))
	err = connection.Ping(ctx)
	if err != nil {
		log.Println(err.Error())
	}
}

func AddNewCache(meta *metamodel.MetaData) {
	var columns []string

	for _, metaElement := range meta.Columns {
		columns = append(columns, metaElement.Name+" "+metaElement.FieldType)
	}

	println(strings.Replace(meta.CreateQuery, "{columns}", strings.Join(columns, ","), 1))
}

func AddData(key string, data map[string]interface{}, meta *metamodel.MetaData) {

	var args = make([]interface{}, 1)
	args[0] = key
	for _, metaElement := range meta.Columns {
		args = append(args, data[strings.Title(metaElement.Name)])
	}
	errExec := connection.Exec(ctx, meta.UpdateQuery, args...)
	if errExec != nil {
		log.Println(errExec.Error())
	}
}

func GetData(meta metamodel.MetaData, key string) ([]byte, error) {
	var result = connection.QueryRow(ctx, meta.GetQuery, key)
	if result.Err() != nil {
		return nil, result.Err()
	}
	//var dest = make([]interface{}, 1)
	//dest[0] = key
	var structFields = make([]reflect.StructField, 0)
	for _, metaElement := range meta.Columns {
		switch metaElement.FieldType {
		case "int32":
			structFields = append(structFields, createStruct(strings.Title(metaElement.Name), reflect.TypeOf((*int32)(nil))))
		case "string":
			structFields = append(structFields, createStruct(strings.Title(metaElement.Name), reflect.TypeOf((*string)(nil))))
		}
	}
	typ := reflect.StructOf(structFields)
	var destinationValue = reflect.New(typ).Interface()
	err := result.ScanStruct(destinationValue)
	if err != nil {
		// todo: сделать обработку ошибки
		log.Println(err)
	}
	value, err := json.Marshal(destinationValue)
	if err != nil {
		return nil, err
	}
	if properties.IsDebugEnabled() {
		log.Println("данные получены из базы данных по ключу: " + key)
	}
	return value, nil
}

func createStruct(name string, fieldType reflect.Type) reflect.StructField {
	var structField reflect.StructField

	structField.Name = name
	structField.Type = fieldType

	return structField

	/*

		reflect.StructFiet := reflect.StructOf([]reflect.StructField{
		{
		Name: "A",
		Type: reflect.TypeOf(int(0)),
		Tag:  `json:"a"`,
		},
		{
		Name: "B",
		Type: reflect.TypeOf(""),
		Tag:  `json:"B"`,
		},
		})ld {
	*/
}
