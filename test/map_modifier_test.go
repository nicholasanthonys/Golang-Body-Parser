package test

import (
	"encoding/json"
	"errors"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

var mapWrapper map[string]model.Wrapper
var requestFromUser model.Wrapper

func init() {
	parallelConfigByte := service.ReadConfigure("./mock/response.json")
	err := json.Unmarshal(parallelConfigByte, &requestFromUser.Configure)
	if err != nil {
		logrus.Error("error reading ./mock/response.json")
		os.Exit(1)
	}
	mapWrapper = make(map[string]model.Wrapper)
	files, err := service.GetListFolder("./mock")
	if err != nil {
		logrus.Error("Error reading ./mock folder")
		os.Exit(1)
	}

	for _, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {
			requestFromUser = model.Wrapper{
				Configure: configure,
				Request: model.Fields{
					Param:  make(map[string]interface{}),
					Header: make(map[string]interface{}),
					Body:   make(map[string]interface{}),
					Query:  make(map[string]interface{}),
				},
				Response: model.Fields{
					Param:  make(map[string]interface{}),
					Header: make(map[string]interface{}),
					Body:   make(map[string]interface{}),
					Query:  make(map[string]interface{}),
				},
			}
			configByte := service.ReadConfigure("./mock/" + file.Name())
			//* assign configure byte to configure
			err = json.Unmarshal(configByte, &configure)
			requestFromUser.Configure = configure
			//*save to map
			mapWrapper[file.Name()] = requestFromUser
		}
	}
}

//* add to body
func TestMapModifierBody(t *testing.T) {
	//Add Body
	wrapperConfigure0 := mapWrapper["configure0.json"]

	//take configure index
	service.AddToWrapper(wrapperConfigure0.Configure.Request.Adds.Body, "--", requestFromUser.Request.Body, mapWrapper)

	transformFunction := service.LoadFunctionFromModule("../plugin/transform.so", "ToJson")
	if transformFunction == nil {
		assert.Error(t, errors.New("Transform function is nil"), "error opening transform function")
	}
	resultByte, err := transformFunction(requestFromUser.Request.Body)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}
	expected := `{"user":{"id":"0","last_name":"peter"}}`

	equal, err := JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}

	//*Modify Body
	service.ModifyWrapper(wrapperConfigure0.Configure.Request.Modifies.Body, "--", requestFromUser.Request.Body, mapWrapper)
	resultByte, err = transformFunction(requestFromUser.Request.Body)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}
	expected = `{"user":{"id":"99","last_name":"parker"}}`
	equal, err = JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}

	//* Deletion Body
	service.DeletionBody(wrapperConfigure0.Configure.Request.Deletes, requestFromUser.Request)
	resultByte, err = transformFunction(requestFromUser.Request.Body)
	expected = `{"user":{"last_name":"parker"}}`
	equal, err = JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}
}

//* Test Add to Header
func TestMapModifierHeader(t *testing.T) {
	//* Add Header
	wrapperConfigure0 := mapWrapper["configure0.json"]
	service.AddToWrapper(wrapperConfigure0.Configure.Request.Adds.Header, "--", requestFromUser.Request.Header, mapWrapper)

	transformFunction := service.LoadFunctionFromModule("../plugin/transform.so", "ToJson")
	if transformFunction == nil {
		assert.Error(t, errors.New("Transform function is nil"), "error opening transform function")
	}

	resultByte, err := transformFunction(requestFromUser.Request.Header)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}

	expected := `{"odd_number":"1357", "fav_character":"naruto"}`
	equal, err := JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}

	//*Modify Header
	service.ModifyWrapper(wrapperConfigure0.Configure.Request.Modifies.Header, "--", requestFromUser.Request.Header, mapWrapper)
	resultByte, err = transformFunction(requestFromUser.Request.Header)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}
	expected = `{"odd_number":"2468", "fav_character":"kakashi"}`
	equal, err = JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}

	//*Deletion Header
	service.DeletionHeaderOrQuery(wrapperConfigure0.Configure.Request.Deletes.Header, requestFromUser.Request.Header)
	resultByte, err = transformFunction(requestFromUser.Request.Header)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}
	expected = `{"fav_character":"kakashi"}`
	equal, err = JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}
}

func TestMapModifierQuery(t *testing.T) {
	//* Add Query
	wrapperConfigure0 := mapWrapper["configure0.json"]
	service.AddToWrapper(wrapperConfigure0.Configure.Request.Adds.Query, "--", requestFromUser.Request.Query, mapWrapper)

	transformFunction := service.LoadFunctionFromModule("../plugin/transform.so", "ToJson")
	if transformFunction == nil {
		assert.Error(t, errors.New("Transform function is nil"), "error opening transform function")
	}

	resultByte, err := transformFunction(requestFromUser.Request.Query)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}

	expected := `{"address":"kopo", "key":"123-456"}`
	equal, err := JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}

	//*Modify query
	service.ModifyWrapper(wrapperConfigure0.Configure.Request.Modifies.Query, "--", requestFromUser.Request.Query, mapWrapper)
	resultByte, err = transformFunction(requestFromUser.Request.Query)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}

	expected = `{"address":"cibaduyut", "key":"456-789"}`
	equal, err = JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}

	//*Deletion Query
	service.DeletionHeaderOrQuery(wrapperConfigure0.Configure.Request.Deletes.Query, requestFromUser.Request.Query)
	resultByte, err = transformFunction(requestFromUser.Request.Query)
	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}
	expected = `{"key":"456-789"}`
	equal, err = JSONBytesEqual([]byte(expected), resultByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte), "should be equal")
	}

}

// JSONEqual compares the JSON from two Readers.
func JSONEqual(a, b io.Reader) (bool, error) {
	var j, j2 interface{}
	d := json.NewDecoder(a)
	if err := d.Decode(&j); err != nil {
		return false, err
	}
	d = json.NewDecoder(b)
	if err := d.Decode(&j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

// JSONBytesEqual compares the JSON in two byte slices.
func JSONBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		logrus.Error("Error unmarhsaling a")
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		logrus.Error("Error unmarhsaling b ")
	}

	return reflect.DeepEqual(j2, j), nil
}
