package test

import (
	"bytes"
	"encoding/json"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"os"

	"strings"
	"testing"
)

var mapWrapper map[string]model.Wrapper
var requestFromUser model.Wrapper
var routes model.Routes

func init() {
	configureDir := "../../configures.example"
	// * Read router.json
	routesByte := util.ReadJsonFile(configureDir + "/router.json")
	err := json.Unmarshal(routesByte, &routes)

	//* let's use email otp project for testing
	fullProjectDir := configureDir + "/" + "emailotp"

	mapWrapper = make(map[string]model.Wrapper)
	files, err := util.GetListFolder(fullProjectDir)
	if err != nil {
		logrus.Error("Error reading testing directory : " + fullProjectDir)
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
			configByte := util.ReadJsonFile(fullProjectDir + "/" + file.Name())
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
	wrapperConfigure1 := mapWrapper["configure-1.json"]

	//take configure index
	service.AddToWrapper(wrapperConfigure1.Configure.Request.Adds.Body, "--", requestFromUser.Request.Body, mapWrapper)

	stream, err := service.Transform(wrapperConfigure1.Configure, requestFromUser.Request.Body)
	if err != nil {
		assert.Error(t, err, "error transform body")
	}

	resultByte := new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected := `{"user":{"id":"1","last_name":"peter"},"from": "configure-1.json"}`
	//
	equal, err := util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	//*Modify Body
	service.ModifyWrapper(wrapperConfigure1.Configure.Request.Modifies.Body, "--", requestFromUser.Request.Body, mapWrapper)
	stream, err = service.Transform(wrapperConfigure1.Configure, requestFromUser.Request.Body)
	if err != nil {
		assert.Error(t, err, "error transform body")
	}
	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected = `{"user":{"id":"99","last_name":"parker"},"from": "configure-1.json"}`
	equal, err = util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	//* Deletion Body
	service.DeletionBody(wrapperConfigure1.Configure.Request.Deletes, requestFromUser.Request)
	stream, err = service.Transform(wrapperConfigure1.Configure, requestFromUser.Request.Body)
	if err != nil {
		assert.Error(t, err, "error transform body")
	}

	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}
	expected = `{"user":{"last_name":"parker"}}`
	equal, err = util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}
}

//* Test Add to Header
func TestMapModifierHeader(t *testing.T) {
	//* Add Header
	wrapperConfigure0 := mapWrapper["configure-1.json"]
	service.AddToWrapper(wrapperConfigure0.Configure.Request.Adds.Header, "--", requestFromUser.Request.Header, mapWrapper)

	stream, err := service.Transform(wrapperConfigure0.Configure, requestFromUser.Request.Header)
	if err != nil {
		assert.Error(t, err, "error performing transform header")
	}

	resultByte := new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected := `{"odd_number":"1357", "fav_character":"naruto"}`
	equal, err := util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	//*Modify Header
	service.ModifyWrapper(wrapperConfigure0.Configure.Request.Modifies.Header, "--", requestFromUser.Request.Header, mapWrapper)

	stream, err = service.Transform(wrapperConfigure0.Configure, requestFromUser.Request.Header)
	if err != nil {
		assert.Error(t, err, "error performing transform header")
	}

	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected = `{"odd_number":"2468", "fav_character":"kakashi"}`
	equal, err = util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	//*Deletion Header
	service.DeletionHeaderOrQuery(wrapperConfigure0.Configure.Request.Deletes.Header, requestFromUser.Request.Header)
	stream, err = service.Transform(wrapperConfigure0.Configure, requestFromUser.Request.Header)
	if err != nil {
		assert.Error(t, err, "error performing transform header")
	}

	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	if err != nil {
		assert.Error(t, err, "error performing convertion to JSON")
	}
	expected = `{"fav_character":"kakashi"}`
	equal, err = util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}
}

func TestMapModifierQuery(t *testing.T) {
	//* Add Query
	wrapperConfigure0 := mapWrapper["configure-1.json"]
	service.AddToWrapper(wrapperConfigure0.Configure.Request.Adds.Query, "--", requestFromUser.Request.Query, mapWrapper)

	stream, err := service.Transform(wrapperConfigure0.Configure, requestFromUser.Request.Query)
	if err != nil {
		assert.Error(t, err, "error performing transform query")
	}

	resultByte := new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected := `{"address":"kopo", "key":"123-456"}`
	equal, err := util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	//*Modify query
	service.ModifyWrapper(wrapperConfigure0.Configure.Request.Modifies.Query, "--", requestFromUser.Request.Query, mapWrapper)

	stream, err = service.Transform(wrapperConfigure0.Configure, requestFromUser.Request.Query)
	if err != nil {
		assert.Error(t, err, "error performing transform query")
	}

	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected = `{"address":"cibaduyut", "key":"456-789"}`
	equal, err = util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	//*Deletion Query
	service.DeletionHeaderOrQuery(wrapperConfigure0.Configure.Request.Deletes.Query, requestFromUser.Request.Query)
	stream, err = service.Transform(wrapperConfigure0.Configure, requestFromUser.Request.Query)
	if err != nil {
		assert.Error(t, err, "error performing transform query")
	}

	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}
	expected = `{"key":"456-789"}`
	equal, err = util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

}
