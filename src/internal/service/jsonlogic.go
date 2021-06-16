package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	CustomPrometheus "github.com/nicholasanthonys/Golang-Body-Parser/internal/prometheus"
	cmap "github.com/orcaman/concurrent-map"
	"reflect"
	"strings"
)

func InterfaceDirectModifier(in interface{}, mapWrapper *cmap.ConcurrentMap, separator string) interface{} {

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
			// Retrieve item from map.
			var wrapper *model.Wrapper
			if tmp, ok := mapWrapper.Get(splittedValue[0]); ok {
				wrapper = tmp.(*model.Wrapper)
			}

			if wrapper == nil {
				return in
			}

			if len(splittedValue) != 3 {
				log.Error("referenced syntax wrong for : ", in)
				log.Error(splittedValue)
				return in
			}

			if splittedValue[1] == "$request" {
				val = GetFromHalfReferenceValue(splittedValue[2], wrapper.Request, 0)
				in = val
			} else {
				val := GetFromHalfReferenceValue(splittedValue[2], wrapper.Response, 0)

				in = val
			}
		}
	}

	return in

}

func CLogicsChecker(cLogicItem model.CLogicItem, mapWrapper *cmap.ConcurrentMap, prefixMetricName string) (bool,
	error) {
	cLogicItem.Data = InterfaceDirectModifier(cLogicItem.Data, mapWrapper, "--")
	cLogicItem.Rule = InterfaceDirectModifier(cLogicItem.Rule, mapWrapper, "--")
	if cLogicItem.Rule == nil {
		return false, nil
	}

	ruleByte, err := json.Marshal(cLogicItem.Rule)
	if err != nil {
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(prefixMetricName)+"ERR_MARSHALLING_CONFIGURE_LOGIC_RULE"].Inc()
		return false, err
	}

	dataByte, err := json.Marshal(cLogicItem.Data)
	if err != nil {
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(prefixMetricName)+"ERR_MARSHALLING_CONFIGURE_LOGIC_DATA"].Inc()
		return false, err
	}
	ruleReader := bytes.NewReader(ruleByte)
	dataReader := bytes.NewReader(dataByte)
	var resultBuf bytes.Buffer
	err = jsonlogic.Apply(ruleReader, dataReader, &resultBuf)

	if err != nil {
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(prefixMetricName)+"ERR_APPLY_CONFIGURE_LOGIC"].Inc()
		log.Error("error is ")
		log.Error(err.Error())
		return false, err
	}

	var result interface{}
	decoder := json.NewDecoder(&resultBuf)

	err = decoder.Decode(&result)
	if err != nil {
		log.Errorf("Error decode logic result : %v", err)
		return false, err
	}

	// get type of json logic result
	vt := reflect.TypeOf(result)
	if vt.Kind() == reflect.Bool {
		if result.(bool) {
			return true, nil
		} else {
			// result is false
			return false, nil
		}
	}
	return true, nil

}
