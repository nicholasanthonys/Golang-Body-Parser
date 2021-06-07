package test

import (
	"bytes"
	"encoding/json"
	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestReadWithoutConfigure(t *testing.T) {
	var project model.Serial
	jsonFile, err := os.Open("../../configures.testing/smsotp/serial.json")
	if err != nil {
		logrus.Error("error when open serial.json")
		logrus.Error(err)
		assert.Error(t, err, "Error reading serial.json")
	}
	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &project)

	if err != nil {
		assert.Error(t, err, " should not error")
	}

	emptyCMap := cmap.New()
	cLogicModified := service.InterfaceDirectModifier(project.Configures[0].CLogics[0], &emptyCMap, "--").(model.CLogicItem)

	expected := project.Configures[0].CLogics[0]
	assert.Equal(t, expected, cLogicModified)

	expectedLogic := []int{2, 4, 6, 8, 10}
	result, err := jsonlogic.ApplyInterface(cLogicModified.Rule, cLogicModified.Data)
	arrayResult := result.([]interface{})

	intResult := make([]int, len(arrayResult))
	for i, val := range arrayResult {
		// Convert to integer
		intResult[i] = int(val.(float64))
	}
	assert.Equal(t, expectedLogic, intResult)
}

func TestReadWithConfigure(t *testing.T) {
	project := model.Serial{}
	configureDir := os.Getenv("CONFIGURES_DIRECTORY_TESTING_NAME")
	fullProjectDirectory := configureDir + "/" + "emailotp"

	jsonFile, err := os.Open(fullProjectDirectory + "/serial.json")
	if err != nil {
		assert.Error(t, err, "Error reading serial.json")
	}

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &project)

	if err != nil {
		assert.Error(t, err, "Error Unmarshalling byteValue to struct project")
	}

	//prepare map model wrapper
	mapWrapper := cmap.New()
	for _, configureItem := range project.Configures {
		var configure model.Configure
		requestFromUser := model.Wrapper{
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

		configByte := util.ReadJsonFile(fullProjectDirectory + "/" + configureItem.FileName)
		//* assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure
		mapWrapper.Set(configureItem.Alias, requestFromUser)
		// mapWrapper[configureItem.Alias] = requestFromUser

		tmpRequest := service.DoAddModifyDelete(requestFromUser.Configure.Request, &requestFromUser.Request, &mapWrapper, 0)

		requestFromUser.Request.Set("header", tmpRequest["header"])
		requestFromUser.Request.Set("body", tmpRequest["body"])
		requestFromUser.Request.Set("query", tmpRequest["query"])
	}

	var tempMap map[string]interface{}
	log.Info(project)
	cLogicBeforeByte, _ := json.Marshal(project.Configures[0].CLogics[0])
	err = json.Unmarshal(cLogicBeforeByte, &tempMap)

	if err != nil {
		assert.Error(t, err, "Error unmarshalling CLogicBefore to tempMap ")
	}

	clogicModified := model.CLogicItem{
		Rule:        service.InterfaceDirectModifier(tempMap["rule"], &mapWrapper, "--"),
		Data:        service.InterfaceDirectModifier(tempMap["data"], &mapWrapper, "--"),
		NextSuccess: "",
		Response:    model.Command{},
	}

	if err != nil {
		assert.Error(t, err, "Error marshaling cLogicModified to byte")
	}

	if err != nil {
		assert.Error(t, err, "Error unmarshaling cLogicModified byte to struct CLogicItem")
	}

	expected := model.CLogicItem{
		Rule: map[string]interface{}{
			"==": []interface{}{"bokir", "bokir"},
		},
		Data:        nil,
		NextSuccess: "",
	}

	assert.Equal(t, expected, clogicModified)

	// Apply json logic
	result, err := jsonlogic.ApplyInterface(clogicModified.Rule, clogicModified.Data)
	boolResult := result.(bool)
	assert.Equal(t, true, boolResult)
}

func TestGetVarArray(t *testing.T) {
	rule := strings.NewReader(`{
          "and": [
            {
              "==": [
                { "var" : "tempNumbers.0"},
               "123-456"
              ]
            },
            {
              "==": [
                { "var" : "tempNumbers.1"},
                "234-567"
              ]
            }
          ]
        }`)

	data := strings.NewReader(`{
          "tempNumbers": [
            "123-456",
            "234-567",
            "345-678"
          ]
        }`)

	var resultBuf bytes.Buffer

	err := jsonlogic.Apply(rule, data, &resultBuf)
	if err != nil {
		assert.Error(t, err, " should not error")
		return
	}

	var result interface{}
	decoder := json.NewDecoder(&resultBuf)
	err = decoder.Decode(&result)
	if err != nil {
		assert.Error(t, err, " should not error")
		return
	}

	vt := reflect.TypeOf(result)

	assert.Equal(t, reflect.Bool, vt.Kind())
	if vt.Kind() == reflect.Bool {
		assert.Equal(t, true, result.(bool))
	}

}
