package service

import (
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
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
		//* if value has prefix $configure
		if strings.HasPrefix(fmt.Sprintf("%v", in), "$configure") {
			splittedValue := strings.Split(fmt.Sprintf("%v", in), separator) //$configure1.json, $request, $body[user][name]
			if splittedValue[1] == "$request" {
				val := RetrieveValue(splittedValue[2], mapWrapper[splittedValue[0]].Request, 0)
				in = val
			} else {
				val := RetrieveValue(splittedValue[2], mapWrapper[splittedValue[0]].Response, 0)
				in = val
			}
		}
	}

	return in

}
