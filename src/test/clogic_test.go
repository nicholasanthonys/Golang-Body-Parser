package test

import (
	"encoding/json"
	"github.com/diegoholiveira/jsonlogic"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestReadWithoutConfigure(t *testing.T) {
	var project model.Project
	jsonFile, err := os.Open("../../configures.example/smsotp/project.json")
	if err != nil {
		logrus.Error("error when open project.json")
		logrus.Error(err)
		assert.Error(t, err, "Error reading project.json")
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
	var project model.Project
	configureDir := os.Getenv("CONFIGURES_DIRECTORY")
	fullProjectDirectory := "../" + configureDir + "/" + "emailotp"

	jsonFile, err := os.Open(fullProjectDirectory + "/project.json")
	if err != nil {
		assert.Error(t, err, "Error reading project.json")
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

		service.DoCommand(requestFromUser.Configure.Request, requestFromUser.Request, mapWrapper)
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
