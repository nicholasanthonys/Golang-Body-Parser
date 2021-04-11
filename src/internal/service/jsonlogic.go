package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

func InterfaceDirectModifier(in interface{}, mapWrapper map[string]model.Wrapper, separator string) interface{} {

	if in == nil {
		return nil

	}

	it := reflect.TypeOf(in)

	if it.Kind() == reflect.Map {

		for index, val := range in.(map[string]interface{}) {
			in.(map[string]interface{})[index] = InterfaceDirectModifier(val, mapWrapper, separator)
		}

	} else if it.Kind() == reflect.Slice {
		for index, el := range in.([]interface{}) {
			in.([]interface{})[index] = InterfaceDirectModifier(el, mapWrapper, separator)
		}
	} else {
		var val interface{}
		//* if value has prefix $configure
		if strings.HasPrefix(fmt.Sprintf("%v", in), "$configure") {
			splittedValue := strings.Split(fmt.Sprintf("%v", in), separator) //$configure1.json, $request, $body[user][name]
			if splittedValue[1] == "$request" {
				val = RetrieveValue(splittedValue[2], mapWrapper[splittedValue[0]].Request, 0)
				in = val
			} else {
				val := RetrieveValue(splittedValue[2], mapWrapper[splittedValue[0]].Response, 0)

				in = val
			}
		}
	}

	return in

}

func CLogicsChecker(cLogics []model.CLogicItem, mapWrapper map[string]model.Wrapper) (*model.CLogicItem, error) {
	for _, cLogicItem := range cLogics {
		cLogicItem.Data = InterfaceDirectModifier(cLogicItem.Data, mapWrapper, "--")
		cLogicItem.Rule = InterfaceDirectModifier(cLogicItem.Rule, mapWrapper, "--")

		ruleByte, err := json.Marshal(cLogicItem.Rule)
		if err != nil {
			return nil, err
		}

		dataByte, err := json.Marshal(cLogicItem.Data)
		if err != nil {
			return nil, err
		}
		ruleReader := bytes.NewReader(ruleByte)
		dataReader := bytes.NewReader(dataByte)
		var resultBuf bytes.Buffer
		err = jsonlogic.Apply(ruleReader, dataReader, &resultBuf)

		if err != nil {
			logrus.Error("error is ")
			logrus.Error(err.Error())
			return nil, err
		}

		var result interface{}
		decoder := json.NewDecoder(&resultBuf)
		decoder.Decode(&result)

		// get type of json logic result
		vt := reflect.TypeOf(result)
		if vt.Kind() == reflect.Bool {
			if result.(bool) {

				return &cLogicItem, nil
			}
		} else {
			return &cLogicItem, nil
		}

	}
	return nil, nil

}
