package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
)

// AddRecursive is a function that do the add key-value based on the listtraverse
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

//DoCommandConfigureBody is a wrapper function to do Add, Deletion and Modify for body
//* Wrapper : wrapper that want to be add
func DoCommandConfigureBody(command model.Command, fields model.Fields, takeFrom map[string]model.Wrapper) {

	//* Add key
	for key, value := range command.Adds.Body {
		//*get the value
		//*split value : $configure1.json-$request-$body[user][name]
		var realValue interface{}
		//* if value has prefix $configure
		if strings.HasPrefix(fmt.Sprintf("%v", value), "$configure") {
			splittedValue := strings.Split(fmt.Sprintf("%v", value), "-") //$configure1.json, $request, $body[user][name]
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
		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), fields.Body, 0)
	}

	//* Do Deletion
	for _, key := range command.Deletes.Body {
		listTraverse := strings.Split(key, ".")
		DeleteRecursive(listTraverse, fields.Body, 0)
	}

	//*Do Modify
	for key, value := range command.Modifies.Body {
		//*get the value
		//*split value : $configure1.json-$request-$body[user][name]
		splittedValue := strings.Split(fmt.Sprintf("%v", value), "-") //$configure1.json, $request, $body[user][name]
		//remove dollar sign
		splittedValue[0] = RemoveDollar(splittedValue[0])
		logrus.Info("splitted value 0 is ", splittedValue[0])
		var realValue interface{}
		if splittedValue[1] == "$request" {
			//* get the request from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Request)
		} else {
			//* get the response from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Response)
		}
		if splittedValue[1] == "$request" {
			//* get the request from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Request)
		} else {
			//* get the response from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Response)
		}
		listTraverseKey := strings.Split(key, ".")
		ModifyRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), fields.Body, 0)
	}
}

// DoCommandConfigureHeader is a wrapper function that do add, modify, delete for header
func DoCommandConfigureHeader(command model.Command, fields model.Fields, takeFrom map[string]model.Wrapper) {
	//*Add to map header
	for key, value := range command.Adds.Header {
		//*get the value
		//*split value : $configure1.json-$request-$body[user][name]
		splittedValue := strings.Split(fmt.Sprintf("%v", value), "-") //$configure1.json, $request, $body[user][name]
		logrus.Info("splitted value 0 is ", splittedValue[0])
		var realValue interface{}
		if splittedValue[1] == "$request" {
			//* get the request from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Request)
		} else {
			//* get the response from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Response)
		}

		listTraverseKey := strings.Split(key, ".")
		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), fields.Header, 0)

	}

	//*Delete
	for _, key := range command.Deletes.Header {
		delete(fields.Header, key)
	}

	//* Modify
	for key, value := range command.Modifies.Header {
		existValue := fmt.Sprintf("%s", fields.Header[strings.Title(key)])
		if len(existValue) > 0 {
			//*get the value
			//*split value : $configure1.json-$request-$body[user][name]
			splittedValue := strings.Split(fmt.Sprintf("%v", value), "-") //$configure1.json, $request, $body[user][name]
			logrus.Info("splitted value 0 is ", splittedValue[0])
			var realValue interface{}
			if splittedValue[1] == "$request" {
				//* get the request from fields
				realValue = checkValue(value, takeFrom[splittedValue[0]].Request)
			} else {
				//* get the response from fields
				realValue = checkValue(value, takeFrom[splittedValue[0]].Response)
			}

			fields.Header[key] = realValue
		}
	}
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
				logrus.Info(in, " is map")
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

// DoCommandConfigureQuery is a wrapper function that do add, modify, delete for query
func DoCommandConfigureQuery(command model.Command, fields model.Fields, takeFrom map[string]model.Wrapper) {
	//* Add
	for key, value := range fields.Query {
		//*get the value
		//*split value : $configure1.json-$request-$body[user][name]
		splittedValue := strings.Split(fmt.Sprintf("%v", value), "-") //$configure1.json, $request, $body[user][name]

		splittedValue[0] = RemoveDollar(splittedValue[0])
		logrus.Info("key is ", key, " splitted value 0 is ", splittedValue[0])
		var realValue interface{}
		if splittedValue[1] == "$request" {
			//* get the request from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Request)
		} else {
			//* get the response from fields
			realValue = checkValue(value, takeFrom[splittedValue[0]].Response)
		}

		listTraverseKey := strings.Split(key, ".")
		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), fields.Query, 0)
	}

	//* Delete
	for _, key := range command.Deletes.Query {
		delete(fields.Query, key)
	}

	//* Modify
	for key, value := range command.Modifies.Query {
		existingValue := fmt.Sprintf("%s", fields.Query[key])
		if len(existingValue) > 0 {

			//*get the value
			//*split value : $configure1.json-$request-$body[user][name]
			splittedValue := strings.Split(fmt.Sprintf("%v", value), "-") //$configure1.json, $request, $body[user][name]
			logrus.Info("splitted value 0 is ", splittedValue[0])
			var realValue interface{}
			if splittedValue[1] == "$request" {
				//* get the request from fields
				realValue = checkValue(value, takeFrom[splittedValue[0]].Request)
			} else {
				//* get the response from fields
				realValue = checkValue(value, takeFrom[splittedValue[0]].Response)
			}

			fields.Query[key] = realValue
		}
	}
}

//*DoCommand is a function that will do the command from configure.json for Header, Query, and Body
//* Here, we call DoCommandConfigure for each Header, Query, and Body
//* fields is field that want to be modify
func DoCommand(command model.Command, fields model.Fields, arrRes map[string]model.Wrapper) {

	//DoCommandConfigureHeader(command, fields, arrRes)
	//DoCommandConfigureQuery(command, fields, arrRes)
	DoCommandConfigureBody(command, fields, arrRes)
}
