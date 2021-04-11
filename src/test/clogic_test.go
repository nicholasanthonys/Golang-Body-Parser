package test

import (
	"bytes"
	"encoding/json"
	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
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
	jsonFile, err := os.Open("../../configures.example/smsotp/serial.json")
	if err != nil {
		logrus.Error("error when open serial.json")
		logrus.Error(err)
		assert.Error(t, err, "Error reading serial.json")
	}
	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &project)

	if err != nil {
		logrus.Error(err.Error())
	}
	cLogicModified := service.InterfaceDirectModifier(project.Configures[0].CLogics[0], map[string]model.Wrapper{}, "--").(model.CLogicItem)

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
	var project model.Serial
	configureDir := os.Getenv("CONFIGURES_DIRECTORY")
	fullProjectDirectory := "../" + configureDir + "/" + "emailotp"

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
	mapWrapper := make(map[string]model.Wrapper)
	for _, configureItem := range project.Configures {
		var configure model.Configure
		requestFromUser := model.Wrapper{
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
		configByte := util.ReadJsonFile(fullProjectDirectory + "/" + configureItem.FileName)
		//* assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure
		mapWrapper[configureItem.Alias] = requestFromUser

		requestFromUser.Request = service.DoAddModifyDelete(requestFromUser.Configure.Request, requestFromUser.Request, mapWrapper, 0)
	}

	var tempMap map[string]interface{}

	cLogicBeforeByte, _ := json.Marshal(project.Configures[0].CLogics[0])
	err = json.Unmarshal(cLogicBeforeByte, &tempMap)

	if err != nil {
		assert.Error(t, err, "Error unmarshalling CLogicBefore to tempMap ")
	}

	logrus.Info("temp map is ")
	logrus.Info(tempMap)

	clogicModified := model.CLogicItem{
		Rule:        service.InterfaceDirectModifier(tempMap["rule"], mapWrapper, "--"),
		Data:        service.InterfaceDirectModifier(tempMap["data"], mapWrapper, "--"),
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
	expectedLogic := true
	result, err := jsonlogic.ApplyInterface(clogicModified.Rule, clogicModified.Data)
	boolResult := result.(bool)
	assert.Equal(t, expectedLogic, boolResult)
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

	jsonlogic.Apply(rule, data, &resultBuf)

	var result interface{}
	decoder := json.NewDecoder(&resultBuf)
	decoder.Decode(&result)

	vt := reflect.TypeOf(result)

	assert.Equal(t, reflect.Bool, vt.Kind())
	if vt.Kind() == reflect.Bool {
		assert.Equal(t, true, result.(bool))
	}

}
