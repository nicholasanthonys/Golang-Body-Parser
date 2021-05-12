package service

import (
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var log = logrus.New()

var mutex = &sync.Mutex{}

func init() {
	//* init logger with timestamp
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.Level = util.GetLogLevelFromEnv()
}

// AddRecursive is a function that do the add key-value based on the listTraverse
func AddRecursive(listTraverse []string, value interface{}, in interface{}, index int) interface{} {

	if index == len(listTraverse)-1 {
		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {

			//*only add when the value of the key is null
			//mutex.Lock()
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				in.(map[string]interface{})[listTraverse[index]] = value
			}
			//mutex.Unlock()
		}
		return in
	}

	if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
		//* allocate new map if map[key] null
		//mutex.Lock()
		if in.(map[string]interface{})[listTraverse[index]] == nil {
			in.(map[string]interface{})[listTraverse[index]] = make(map[string]interface{})
		}
		//mutex.Unlock()

		//* recursively traverse the map
		in.(map[string]interface{})[listTraverse[index]] = AddRecursive(listTraverse, value, in.(map[string]interface{})[listTraverse[index]], index+1)

		return in.(map[string]interface{})
	}

	return in

}

// ModifyRecursive is a function that do modify key-value based on listTraverse
func ModifyRecursive(listTraverse []string, value interface{}, in interface{}, index int) interface{} {

	if index == len(listTraverse)-1 {

		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				return nil
			}
			in.(map[string]interface{})[listTraverse[index]] = value
		}
		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "[]interface {}" {
			realIndex, err := strconv.Atoi(listTraverse[index])
			if err == nil {
				in.([]interface{})[realIndex] = value
			} else {
				log.Error(err.Error())
				log.Error("error converting index")
			}

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

// RetrieveValue is a function that check the value type value from configure and retrieve the value from header,body, or query
func RetrieveValue(value interface{}, takeFrom cmap.ConcurrentMap, loopIndex int) interface{} {
	//*declare empty result
	var realValue interface{}
	//* check the type of the value
	vt := reflect.TypeOf(value).Kind()

	if reflect.String == vt {
		//* We Call Sanitizevalue to clear the value from the square bracket and the Dollar Sign
		listTraverseVal, destination := util.SanitizeValue(fmt.Sprintf("%v", value))

		if len(destination) == 0 {
			log.Info("destination not found, returning : ", realValue)

			return realValue
		}
		if listTraverseVal != nil {
			var key string
			if destination == "body" {
				key = "body"
			}
			if destination == "header" {
				key = "header"
			}
			if destination == "query" {
				key = "query"
			}
			if destination == "path" {
				key = "param"
			}
			if destination == "status_code" {
				key = "statusCode"
			}
			if tmp, ok := takeFrom.Get(key); ok {
				if key != "statusCode" {
					mutex.Lock()
					tmpMap := tmp.(map[string]interface{})
					realValue = recursiveGetValue(listTraverseVal, tmpMap, 0, loopIndex)
					mutex.Unlock()
				} else {
					mutex.Lock()
					vt := reflect.TypeOf(tmp)
					if vt.Kind() == reflect.String {
						tmp, _ = strconv.Atoi(tmp.(string))
					}
					tmpMap := tmp.(int)
					realValue = recursiveGetValue(listTraverseVal, tmpMap, 0, loopIndex)
					mutex.Unlock()
				}

			}
		} else {
			realValue = value
		}

	} else {

		realValue = value
	}
	if realValue == nil {
		//* if value is not found, return empty string
		return ""
	}
	return realValue
}

//* recursiveGetValue is a function that will recursively traverse the whole map
//* get the value based on the listTraverse
func recursiveGetValue(listTraverse []string, in interface{}, index int, loopIndex int) interface{} {

	if len(listTraverse) > 0 {
		if index == len(listTraverse)-1 {
			//*check the type of the target
			rt := reflect.TypeOf(in)
			switch rt.Kind() {
			case reflect.Slice:
				var indexInt int
				var err error

				//*check type slice element
				//* example :  $body[user][name][0]. Now we have the 0 as index type string. we need to
				//* convert the 0 to become integer

				// handle case if index is loop, ex $body[user][loop]
				if listTraverse[index] == "loop" {
					indexInt = loopIndex
				} else {
					indexInt, err = strconv.Atoi(listTraverse[index])
				}

				if err != nil {
					log.Error("error converting string to integer")
					return nil
				}

				//*if the type of the interface is slice
				if len(in.([]interface{})) > indexInt {

					return in.([]interface{})[indexInt]
				}
				return nil

			case reflect.Map:
				if val, ok := in.(map[string]interface{})[listTraverse[index]]; ok {
					return val
				}
			default:
				// return the whole interface
				return nil
			}
		}

		//*if the type is map, we need to traverse recursively again
		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			//*if the map is nil, return interface
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				return nil
			}
			//* recursively traverse the map again
			return recursiveGetValue(listTraverse, in.(map[string]interface{})[listTraverse[index]], index+1, loopIndex)
		} else {
			return nil
		}
	}

	return in
}

//*DoAddModifyDelete is a function that will do the command from configure.json for Header, Query, and Body
//* Here, we call DoCommandConfigure for each Header, Query, and Body
//* fields is field that want to be modify
func DoAddModifyDelete(command model.Command, fields *cmap.ConcurrentMap, takeFrom *cmap.ConcurrentMap, loopIndex int) map[string]interface{} {
	//tmpHeader := make(map[string]interface{})
	//tmpBody := make(map[string]interface{})
	//tmpQuery := make(map[string]interface{})
	var fieldHeader map[string]interface{}
	var fieldQuery map[string]interface{}
	var fieldBody map[string]interface{}

	//*header
	if tmp, ok := fields.Get("header"); ok {
		fieldHeader = tmp.(map[string]interface{})

		fieldHeader = AddToWrapper(command.Adds.Header, "--", fieldHeader, takeFrom, loopIndex)
		//*modify header
		fieldHeader = ModifyWrapper(command.Modifies.Header, "--", fieldHeader, takeFrom, loopIndex)
		//*Deletion Header
		fieldHeader = DeletionHeaderOrQuery(command.Deletes.Header, fieldHeader)
	}

	if tmp, ok := fields.Get("query"); ok {
		fieldQuery = tmp.(map[string]interface{})
		//* Add Query
		fieldQuery = AddToWrapper(command.Adds.Query, "--", fieldQuery, takeFrom, loopIndex)
		//*modify Query
		fieldQuery = ModifyWrapper(command.Modifies.Query, "--", fieldQuery, takeFrom, loopIndex)
		//*Deletion Query
		fieldQuery = DeletionHeaderOrQuery(command.Deletes.Query, fieldQuery)
	}

	if tmp, ok := fields.Get("body"); ok {
		fieldBody = tmp.(map[string]interface{})

		//* add body
		fieldBody = AddToWrapper(command.Adds.Body, "--", fieldBody, takeFrom, loopIndex)
		//*modify body
		fieldBody = ModifyWrapper(command.Modifies.Body, "--", fieldBody, takeFrom, loopIndex)
		//*deletion to body
		fieldBody = DeletionBody(command.Deletes, fieldBody)
	}
	return map[string]interface{}{
		"header": fieldHeader,
		"body":   fieldBody,
		"query":  fieldQuery,
	}
}

func DeletionBody(deleteField model.DeleteFields, mapKeyToBeRemoved map[string]interface{}) map[string]interface{} {
	//* Do Deletion
	for _, key := range deleteField.Body {
		listTraverse := strings.Split(key, ".")
		mutex.Lock()
		DeleteRecursive(listTraverse, mapKeyToBeRemoved, 0)
		mutex.Unlock()
	}
	return mapKeyToBeRemoved
}

func DeletionHeaderOrQuery(deleteField []string, mapToBeDeleted map[string]interface{}) map[string]interface{} {
	//* Do Deletion
	for _, key := range deleteField {
		delete(mapToBeDeleted, key)
	}
	return mapToBeDeleted
}

func ModifyPath(path string, separator string, takeFrom *cmap.ConcurrentMap, loopIndex int) string {
	//*example, what we got here is like this
	//* /person/{{$configure1.json--$request--$body[user][name]/transaction/{{$configure1.json--$request--$body[user][name]}}
	//* we need to split based from separator /, and looping and find if there is {{ }}
	splittedPath := strings.Split(path, "/")
	for _, val := range splittedPath {
		if strings.Contains(val, "{{") && strings.Contains(val, "}}") {
			removedBracket := util.RemoveCharacters(val, "{{}}")

			//*split value : $configure1.json-$request-$body[user][name]
			var realValue interface{}
			//* if value has prefix $configure
			if strings.HasPrefix(fmt.Sprintf("%v", removedBracket), "$configure") {
				splittedValue := strings.Split(fmt.Sprintf("%v", removedBracket), separator) //$configure1.json, $request, $body[user][name]
				if len(splittedValue) != 3 {
					log.Error("referenced syntax wrong for : ", removedBracket)
					log.Error(splittedValue)
					return ""
				}

				//remove dollar sign
				var wrapper *model.Wrapper
				if tmp, ok := takeFrom.Get(splittedValue[0]); ok {
					wrapper = tmp.(*model.Wrapper)
				} else {
					return ""
				}
				if splittedValue[1] == "$request" {
					//* get the request from fields
					realValue = RetrieveValue(splittedValue[2], wrapper.Request, loopIndex)
				} else {
					//* get the response from fields
					realValue = RetrieveValue(splittedValue[2], wrapper.Response, loopIndex)
				}

				if realValue != nil {

					vt := reflect.TypeOf(realValue).Kind()
					if reflect.String == vt {

						path = strings.Replace(path, val, realValue.(string), -1)
					}

				} else {
					log.Info("real value for path is nil, returning empty string")
					return ""
				}

			}

		}
	}

	return path

}

//*AddToWrapper is a function that will add value to the specified key to a map
func AddToWrapper(commands map[string]interface{}, separator string, mapToBeAdded map[string]interface{}, takeFrom *cmap.ConcurrentMap, loopIndex int) map[string]interface{} {
	//* Add key
	for key, value := range commands {
		//*get the value
		//*split value : $configure1.json-$request-$body[user][name]
		var realValue interface{}
		//* if value has prefix $configure
		if strings.HasPrefix(fmt.Sprintf("%v", value), "$configure") {
			splittedValue := strings.Split(fmt.Sprintf("%v", value), separator) //$configure1.json, $request, $body[user][name]
			if len(splittedValue) != 3 {
				log.Error("referenced syntax wrong for : ", value)
				log.Error(splittedValue)
				return mapToBeAdded
			}

			//remove dollar sign
			//splittedValue[0] = util.RemoveCharacters(splittedValue[0], "$")
			var wrapper *model.Wrapper
			if tmp, ok := takeFrom.Get(splittedValue[0]); ok {
				wrapper = tmp.(*model.Wrapper)
				if splittedValue[1] == "$request" {
					//* get the request from fields
					realValue = RetrieveValue(splittedValue[2], wrapper.Request, loopIndex)
				} else {
					//* get the response from fields
					realValue = RetrieveValue(splittedValue[2], wrapper.Response, loopIndex)
				}
			}

		} else {
			realValue = value
		}
		listTraverseKey := strings.Split(key, ".")

		mutex.Lock()
		AddRecursive(listTraverseKey, realValue, mapToBeAdded, 0)
		mutex.Unlock()
	}
	return mapToBeAdded
}

//*ModifyWrapper is a function that will modify value based from specific key
func ModifyWrapper(commands map[string]interface{}, separator string, mapToBeModified map[string]interface{}, takeFrom *cmap.ConcurrentMap, loopIndex int) map[string]interface{} {
	for key, value := range commands {

		var realValue interface{}
		//* if value has prefix $configurex.json
		if strings.HasPrefix(fmt.Sprintf("%v", value), "$configure") {
			//* split : $configure1.json-$request-$body[user]
			//* into $configure1.json, $request, $body[user]
			splittedValue := strings.Split(fmt.Sprintf("%v", value), separator) //$configure1.json, $request, $body[user][name]
			if len(splittedValue) != 3 {
				log.Error("referenced syntax wrong for : ", value)
				log.Error(splittedValue)
				return mapToBeModified
			}
			////remove dollar sign from $configure
			//splittedValue[0] = util.RemoveCharacters(splittedValue[0], "$")
			var wrapper *model.Wrapper
			if tmp, ok := takeFrom.Get(splittedValue[0]); ok {
				wrapper = tmp.(*model.Wrapper)
			}
			if splittedValue[1] == "$request" {
				//* get the request from fields
				realValue = RetrieveValue(splittedValue[2], wrapper.Request, loopIndex)
			} else {
				//* get the response from fields
				realValue = RetrieveValue(splittedValue[2], wrapper.Response, loopIndex)
			}

		} else {
			realValue = value
		}

		listTraverseKey := strings.Split(key, ".")
		mutex.Lock()
		ModifyRecursive(listTraverseKey, realValue, mapToBeModified, 0)
		mutex.Unlock()
	}
	return mapToBeModified
}
