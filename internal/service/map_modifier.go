package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"reflect"
	"strconv"
	"strings"
)

// AddRecursive is a function that do the add key-value based on the listTraverse
func AddRecursive(listTraverse []string, value string, in interface{}, index int) interface{} {
	if index == len(listTraverse)-1 {
		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			//*only add when the value of the key is null
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				in.(map[string]interface{})[listTraverse[index]] = value
			}

		}
		return in
	}

	if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
		//* allocate new map if map[key] null
		if in.(map[string]interface{})[listTraverse[index]] == nil {
			in.(map[string]interface{})[listTraverse[index]] = make(map[string]interface{})
		}
		//* recursively traverse the map
		in.(map[string]interface{})[listTraverse[index]] = AddRecursive(listTraverse, value, in.(map[string]interface{})[listTraverse[index]], index+1)
		return in.(map[string]interface{})
	}

	return in

}

// ModifyRecursive is a function that do modify key-value based on listTraverse
func ModifyRecursive(listTraverse []string, value string, in interface{}, index int) interface{} {
	if index == len(listTraverse)-1 {

		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				return nil
			}
			in.(map[string]interface{})[listTraverse[index]] = value
		}
		return in
	}
	if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
		if in.(map[string]interface{})[listTraverse[index]] != nil {
			ModifyRecursive(listTraverse, value, in.(map[string]interface{})[listTraverse[index]], index+1)
			return in.(map[string]interface{})
		}
	}

	return nil

}

// DeleteRecursive is a function that do the deletion of key based on list traverse
func DeleteRecursive(listTraverse []string, in interface{}, index int) interface{} {
	if index == len(listTraverse)-1 {

		if in.(map[string]interface{})[listTraverse[index]] == nil {
			return nil
		}

		//in.(map[string]interface{})[listTraverse[index]] = "deleted"
		delete(in.(map[string]interface{}), listTraverse[index])

		return in
	}

	if in.(map[string]interface{})[listTraverse[index]] != nil {
		DeleteRecursive(listTraverse, in.(map[string]interface{})[listTraverse[index]], index+1)
		return in.(map[string]interface{})
	}

	return nil
}

// checkValue is a function that check the value type value from configure and retrieve the value from header,body, or query
func checkValue(value interface{}, takeFrom model.Fields) interface{} {
	//*declare empty result
	var realValue interface{}
	//* check the type of the value
	vt := reflect.TypeOf(value).Kind()

	if reflect.String == vt {
		//* We Call Sanitizevalue to clear the value from the square bracket and the Dollar Sign
		listTraverseVal, destination := SanitizeValue(fmt.Sprintf("%v", value))
		if listTraverseVal != nil {
			if destination == "body" {
				realValue = GetValue(listTraverseVal, takeFrom.Body, 0)
			} else if destination == "header" {
				realValue = GetValue(listTraverseVal, takeFrom.Header, 0)
			} else if destination == "query" {
				realValue = GetValue(listTraverseVal, takeFrom.Query, 0)
			} else if destination == "path" {
				//realValue = c.Param(listTraverseVal[0])
				realValue = takeFrom.Param[listTraverseVal[0]]
			}
		} else {
			realValue = value
		}

	} else {

		realValue = value
	}
	return realValue
}

//* GetValue is a function that will recursively traverse the whole map
//* get the value based on the listTraverse
func GetValue(listTraverse []string, in interface{}, index int) interface{} {
	if len(listTraverse) > 0 {
		if index == len(listTraverse)-1 {
			//*check the type of the target
			rt := reflect.TypeOf(in)
			switch rt.Kind() {
			case reflect.Slice:
				var indexInt int
				//*check type slice element
				et := reflect.TypeOf(in).Elem().Kind()
				//* example :  $body[user][name][0]. Now we have the 0 as index type string. we need to
				//* convert the 0 to become integer
				indexInt, _ = strconv.Atoi(listTraverse[index])
				//*if the type of the interface is slice
				if et == reflect.Interface {
					return in.([]interface{})[indexInt]
				}
				return in.([]string)[indexInt]
			case reflect.Map:

				return in.(map[string]interface{})[listTraverse[index]]
			default:
				// return the whole interface
				return in
			}
		}

		//*if the type is map, we need to traverse recursively again
		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			//*if the map is nil, return interface
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				return in
			}
			//* recursively traverse the map again
			return GetValue(listTraverse, in.(map[string]interface{})[listTraverse[index]], index+1)
		} else {
			return nil
		}
	}

	return in
}

//*DoCommand is a function that will do the command from configure.json for Header, Query, and Body
//* Here, we call DoCommandConfigure for each Header, Query, and Body
//* fields is field that want to be modify
func DoCommand(command model.Command, fields model.Fields, takeFrom map[string]model.Wrapper) {

	//*header
	AddToWrapper(command.Adds.Header, command.Separator, fields.Header, takeFrom)
	//*modify header
	ModifyWrapper(command.Modifies.Header, command.Separator, fields.Header, takeFrom)
	//*Deletion Header
	DeletionHeaderOrQuery(command.Deletes.Header, fields.Header)

	//* Add Query
	AddToWrapper(command.Adds.Query, command.Separator, fields.Query, takeFrom)
	//*modify Query
	ModifyWrapper(command.Modifies.Query, command.Separator, fields.Query, takeFrom)
	//*Deletion Query
	DeletionHeaderOrQuery(command.Deletes.Query, fields.Query)

	//* add body
	AddToWrapper(command.Adds.Body, command.Separator, fields.Body, takeFrom)
	//*modify body
	ModifyWrapper(command.Modifies.Body, command.Separator, fields.Body, takeFrom)
	//*deletion to body
	DeletionBody(command.Deletes, fields)

}

func DeletionBody(deleteField model.DeleteFields, fields model.Fields) {
	//* Do Deletion
	for _, key := range deleteField.Body {
		listTraverse := strings.Split(key, ".")
		DeleteRecursive(listTraverse, fields.Body, 0)
	}
}

func DeletionHeaderOrQuery(deleteField []string, mapToBeDeleted map[string]interface{}) {
	//* Do Deletion
	for _, key := range deleteField {
		delete(mapToBeDeleted, key)
	}
}

//*AddToWrapper is a function that will add value to the specified key to a map
func AddToWrapper(commands map[string]interface{}, separator string, mapToBeAdded map[string]interface{}, takeFrom map[string]model.Wrapper) {
	//* Add key
	for key, value := range commands {
		//*get the value
		//*split value : $configure1.json-$request-$body[user][name]
		var realValue interface{}
		//* if value has prefix $configure
		if strings.HasPrefix(fmt.Sprintf("%v", value), "$configure") {
			splittedValue := strings.Split(fmt.Sprintf("%v", value), separator) //$configure1.json, $request, $body[user][name]
			//remove dollar sign
			splittedValue[0] = RemoveDollar(splittedValue[0])
			if splittedValue[1] == "$request" {
				//* get the request from fields

				realValue = checkValue(splittedValue[2], takeFrom[splittedValue[0]].Request)

			} else {

				//* get the response from fields
				realValue = checkValue(splittedValue[2], takeFrom[splittedValue[0]].Response)
			}
		} else {
			realValue = fmt.Sprintf("%v", value)
		}

		listTraverseKey := strings.Split(key, ".")
		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), mapToBeAdded, 0)
	}
}

//*ModifyWrapper is a function that will modify value based from specific key
func ModifyWrapper(commands map[string]interface{}, separator string, mapToBeModified map[string]interface{}, takeFrom map[string]model.Wrapper) {
	for key, value := range commands {

		var realValue interface{}
		//* if value has prefix $configurex.json
		if strings.HasPrefix(fmt.Sprintf("%v", value), "$configure") {
			//* split : $configure1.json-$request-$body[user]
			//* into $configure1.json, $request, $body[user]
			splittedValue := strings.Split(fmt.Sprintf("%v", value), separator) //$configure1.json, $request, $body[user][name]
			//remove dollar sign from $configure
			splittedValue[0] = RemoveDollar(splittedValue[0])

			if splittedValue[1] == "$request" {
				//* get the request from fields
				realValue = checkValue(value, takeFrom[splittedValue[0]].Request)
			} else {
				//* get the response from fields
				realValue = checkValue(value, takeFrom[splittedValue[0]].Response)
			}

		} else {
			realValue = fmt.Sprintf("%v", value)
		}

		listTraverseKey := strings.Split(key, ".")
		ModifyRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), mapToBeModified, 0)
	}
}
