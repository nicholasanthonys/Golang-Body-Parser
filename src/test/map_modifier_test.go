package test

import (
	"bytes"
	"encoding/json"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mapWrapper cmap.ConcurrentMap
var requestFromUser model.Wrapper
var routes model.Route

func init() {
	configureDir := "../../configures.example"
	// * Read router.json
	routesByte := util.ReadJsonFile(configureDir + "/router.json")
	err := json.Unmarshal(routesByte, &routes)

	//* let's use email otp project for testing
	fullProjectDir := configureDir + "/" + "emailotp"

	mapWrapper = cmap.New()

	var project model.Serial
	projectByte := util.ReadJsonFile(fullProjectDir + "/" + "serial.json")
	err = json.Unmarshal(projectByte, &project)

	if err != nil {
		logrus.Error(err.Error())
		logrus.Fatal("Error read serial.json")
	}

	for _, configureItem := range project.Configures {
		var configure model.Configure
		requestFromUser = model.Wrapper{
			Configure: configure,
			Request:   cmap.New(),
			Response:  cmap.New(),
		}
		requestFromUser.Request.Set("param", make(map[string]interface{}))
		requestFromUser.Request.Set("header", make(map[string]interface{}))
		requestFromUser.Request.Set("body", make(map[string]interface{}))
		requestFromUser.Request.Set("query", make(map[string]interface{}))

		requestFromUser.Response.Set("statusCode", make(map[string]interface{}))
		requestFromUser.Response.Set("header", make(map[string]interface{}))
		requestFromUser.Response.Set("body", make(map[string]interface{}))

		configByte := util.ReadJsonFile(fullProjectDir + "/" + configureItem.FileName)
		//* assign configure byte to configure
		err = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure

		//*save to map
		//mapWrapper[configureItem.Alias] = requestFromUser
		mapWrapper.Set(configureItem.Alias, requestFromUser)
	}
}

//* add to body
func TestMapModifierBody(t *testing.T) {
	//Add Body
	var wrapperConfigure1 model.Wrapper
	if tmp, ok := mapWrapper.Get("$configure_second_configure"); ok {
		wrapperConfigure1 = tmp.(model.Wrapper)
	}

	//take configure index
	tmpBody := make(map[string]interface{})
	if tmp, ok := requestFromUser.Request.Get("body"); ok {
		tmpBody = tmp.(map[string]interface{})
	}
	tmpBody = service.AddToWrapper(wrapperConfigure1.Configure.Request.Adds.Body, "--", tmpBody, &mapWrapper, 0)
	stream, err := service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpBody)

	if err != nil {
		assert.Error(t, err, "error transform body")
	}

	resultByte := new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected := `{"user":{"id":1,"last_name":"peter", "name":"bokir"},"from": "configure-1.json"}`

	equal, err := util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	// Modify Body
	tmpBody = service.ModifyWrapper(wrapperConfigure1.Configure.Request.Modifies.Body, "--", tmpBody, &mapWrapper, 0)
	stream, err = service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpBody)
	if err != nil {
		assert.Error(t, err, "error transform body")
	}
	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}

	expected = `{"user":{"id":99,"last_name":"parker", "name" : "bokir"},"from": "configure-1.json"}`
	equal, err = util.JSONBytesEqual([]byte(expected), resultByte.Bytes())
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(resultByte.Bytes()), "should be equal")
	}

	// Deletion Body
	tmpBody = service.DeletionBody(wrapperConfigure1.Configure.Request.Deletes, tmpBody)
	stream, err = service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpBody)
	if err != nil {
		assert.Error(t, err, "error transform body")
	}

	resultByte = new(bytes.Buffer)
	_, err = resultByte.ReadFrom(stream)
	if err != nil {
		assert.Error(t, err, "error read from stream ")
	}
	expected = `{"user":{"last_name":"parker", "name" : "bokir"}}`
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
	var wrapperConfigure1 model.Wrapper
	if tmp, ok := mapWrapper.Get("$configure_second_configure"); ok {
		wrapperConfigure1 = tmp.(model.Wrapper)
	}
	//wrapperConfigure1 := mapWrapper["$configure_second_configure"]
	tmpHeader := make(map[string]interface{})
	if tmp, ok := requestFromUser.Request.Get("header"); ok {
		tmpHeader = tmp.(map[string]interface{})
	}
	tmpHeader = service.AddToWrapper(wrapperConfigure1.Configure.Request.Adds.Header, "--", tmpHeader, &mapWrapper, 0)

	stream, err := service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpHeader)
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
	tmpHeader = service.ModifyWrapper(wrapperConfigure1.Configure.Request.Modifies.Header, "--", tmpHeader, &mapWrapper, 0)

	stream, err = service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpHeader)
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
	tmpHeader = service.DeletionHeaderOrQuery(wrapperConfigure1.Configure.Request.Deletes.Header, tmpHeader)
	stream, err = service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpHeader)
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
	//wrapperConfigure1 := mapWrapper["$configure_second_configure"]
	var wrapperConfigure1 model.Wrapper
	if tmp, ok := mapWrapper.Get("$configure_second_configure"); ok {
		wrapperConfigure1 = tmp.(model.Wrapper)
	}
	tmpQuery := make(map[string]interface{})
	if tmp, ok := requestFromUser.Request.Get("query"); ok {
		tmpQuery = tmp.(map[string]interface{})
	}
	tmpQuery = service.AddToWrapper(wrapperConfigure1.Configure.Request.Adds.Query, "--", tmpQuery, &mapWrapper, 0)

	stream, err := service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpQuery)
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
	tmpQuery = service.ModifyWrapper(wrapperConfigure1.Configure.Request.Modifies.Query, "--", tmpQuery, &mapWrapper, 0)

	stream, err = service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpQuery)
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
	service.DeletionHeaderOrQuery(wrapperConfigure1.Configure.Request.Deletes.Query, tmpQuery)
	stream, err = service.Transform(wrapperConfigure1.Configure.Request.Transform, tmpQuery)
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
