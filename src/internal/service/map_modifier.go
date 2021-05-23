package service

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
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
			if in.(map[string]interface{})[listTraverse[index]] == nil {

				tmpInterface := make(map[string]interface{})
				tmp := in.(map[string]interface{})
				err := copier.Copy(&tmpInterface, &tmp)
				if err != nil {
					log.Error(err.Error())
					return nil
				}

				tmpInterface[listTraverse[index]] = value
				err = copier.Copy(&in, &tmpInterface)
				if err != nil {
					log.Error(err.Error())
					return nil
				}
			}
		}

		return in
	}

	if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
		// allocate new map if map[key] null

		if in.(map[string]interface{})[listTraverse[index]] == nil {
			in.(map[string]interface{})[listTraverse[index]] = make(map[string]interface{})
		}

		// recursively traverse the map
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

			tmpInterface := make(map[string]interface{})
			tmp := in.(map[string]interface{})
			err := copier.Copy(&tmpInterface, &tmp)
			if err != nil {
				log.Error(err.Error())
				return nil
			}

			tmpInterface[listTraverse[index]] = value
			err = copier.Copy(&in, &tmpInterface)
			if err != nil {
				log.Error(err.Error())
				return nil
			}
			return tmpInterface
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
			in.(map[string]interface{})[listTraverse[index]] = ModifyRecursive(listTraverse, value, in.(map[string]interface{})[listTraverse[index]], index+1)
			return in.(map[string]interface{})
		}
	}

	return in

}

// DeleteRecursive is a function that do the deletion of key based on list traverse
func DeleteRecursive(listTraverse []string, in interface{}, index int) interface{} {
	if index == len(listTraverse)-1 {

		t := reflect.TypeOf(in)
		vt := t.Kind()
		if vt == reflect.Map {
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				log.Info(" returning nil with list traverse index ", listTraverse[index])
				return nil
			}
			delete(in.(map[string]interface{}), listTraverse[index])
		}

		if vt == reflect.Slice {
			deletedIndex, err := strconv.Atoi(listTraverse[index])
			if err != nil {
				log.Error(err.Error())
				return in
			}
			mySlice := make([]interface{}, 0)
			for index, val := range in.([]interface{}) {
				if deletedIndex != index {
					mySlice = append(mySlice, val)
				}
			}
			in = mySlice
			return in
		}

		return in
	}

	if in.(map[string]interface{})[listTraverse[index]] != nil {
		in.(map[string]interface{})[listTraverse[index]] = DeleteRecursive(listTraverse, in.(map[string]interface{})[listTraverse[index]], index+1)
		return in
	}

	return in
}

// GetFromHalfReferenceValue is a function that check the value type value from configure and retrieve the value from header,body, or query
func GetFromHalfReferenceValue(value interface{}, takeFrom cmap.ConcurrentMap, loopIndex int) interface{} {
	// declare empty result
	var realValue interface{}
	// check the type of the value
	vt := reflect.TypeOf(value).Kind()

	if reflect.String != vt {
		log.Error("reference value type is not string. for value : ", value)
		return value
	}

	// We Call GetListTraverseAndDestination to clear the value from the square bracket and the Dollar Sign, and get list traverse and destination
	listTraverseVal, destination := util.GetListTraverseAndDestination(fmt.Sprintf("%v", value))

	if len(destination) == 0 || listTraverseVal == nil {
		return realValue
	}

	if tmp, ok := takeFrom.Get(destination); ok {
		if destination != "statusCode" {
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

	} else {
		log.Error("cannot get reference value destination : ", destination, " for value ", value)
	}

	if realValue == nil {
		// if value is not found, return empty string
		return ""
	}
	return realValue
}

// recursiveGetValue is a function that will recursively traverse the whole map
// get the value based on the listTraverse
func recursiveGetValue(listTraverse []string, in interface{}, index int, loopIndex int) interface{} {
	if len(listTraverse) > 0 {
		if index == len(listTraverse)-1 {
			//*check the type of the target
			rt := reflect.TypeOf(in)

			switch rt.Kind() {
			case reflect.Slice:
				var indexInt int
				var err error

				// check type slice element
				// example :  $body[user][name][0]. Now we have the 0 as index type string. we need to
				// convert the 0 to become integer

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

				//if the type of the interface is slice
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

		//if the type is map, we need to traverse recursively again
		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			//if the map is nil, return interface
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				return nil
			}
			// recursively traverse the map again
			return recursiveGetValue(listTraverse, in.(map[string]interface{})[listTraverse[index]], index+1, loopIndex)
		} else {
			return nil
		}
	}

	return in
}

// DoAddModifyDelete is a function that will do the command from configure.json for Header, Query, and Body
// Here, we call DoCommandConfigure for each Header, Query, and Body
// fields is field that want to be modify
func DoAddModifyDelete(command model.Command, fields *cmap.ConcurrentMap, takeFrom *cmap.ConcurrentMap, loopIndex int) map[string]interface{} {
	var fieldHeader map[string]interface{}
	var fieldQuery map[string]interface{}
	var fieldBody map[string]interface{}

	//*header
	if tmp, ok := fields.Get("header"); ok {
		fieldHeader = tmp.(map[string]interface{})
		fieldHeader = AddToWrapper(command.Adds.Header, "--", fieldHeader, takeFrom, loopIndex)
		fieldHeader = ModifyWrapper(command.Modifies.Header, "--", fieldHeader, takeFrom, loopIndex)
		fieldHeader = DeletionHeaderOrQuery(command.Deletes.Header, fieldHeader)
	}

	if tmp, ok := fields.Get("query"); ok {
		fieldQuery = tmp.(map[string]interface{})
		fieldQuery = AddToWrapper(command.Adds.Query, "--", fieldQuery, takeFrom, loopIndex)
		fieldQuery = ModifyWrapper(command.Modifies.Query, "--", fieldQuery, takeFrom, loopIndex)
		fieldQuery = DeletionHeaderOrQuery(command.Deletes.Query, fieldQuery)
	}

	if tmp, ok := fields.Get("body"); ok {
		fieldBody = tmp.(map[string]interface{})
		fieldBody = AddToWrapper(command.Adds.Body, "--", fieldBody, takeFrom, loopIndex)
		fieldBody = ModifyWrapper(command.Modifies.Body, "--", fieldBody, takeFrom, loopIndex)
		fieldBody = DeletionBody(command.Deletes, fieldBody)
	}
	return map[string]interface{}{
		"header": fieldHeader,
		"body":   fieldBody,
		"query":  fieldQuery,
	}
}

func DeletionBody(deleteField model.DeleteFields, mapKeyToBeRemoved map[string]interface{}) map[string]interface{} {
	// Do Deletion
	for _, key := range deleteField.Body {
		listTraverse := strings.Split(key, ".")
		mutex.Lock()
		result := DeleteRecursive(listTraverse, mapKeyToBeRemoved, 0)
		if result != nil {
			mapKeyToBeRemoved = result.(map[string]interface{})
		}
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
	// example, what we got here is like this :
	// person/{{$configure1.json--$request--$body[user][name]/transaction/{{$configure1.json--$request--$body[user][name]}}
	// we need to split based from separator /, and looping and find if there is {{ }}
	splittedPath := strings.Split(path, "/")
	for _, val := range splittedPath {
		if strings.Contains(val, "{{") && strings.Contains(val, "}}") {
			valWithoutBracket := util.RemoveCharacters(val, "{{}}")

			realValue, err := GetFromFullReferenceValue(separator, takeFrom, loopIndex, valWithoutBracket)
			if err != nil {
				log.Error("from function modifyPath. Error is ", err.Error())
				return path
			}
			if realValue != nil {
				vt := reflect.TypeOf(realValue).Kind()
				if reflect.String == vt {
					// replace path with pattern splittedPath value with real value
					path = strings.Replace(path, val, realValue.(string), -1)
					return path
				}
			}
			log.Error("real value for path is nil, returning original path")
			return path

		}
	}

	return path
}

//GetFromFullReferenceValue is a function that will get value from full reference like $configure--first--configure--$body[user][name]
func GetFromFullReferenceValue(separator string, takeFrom *cmap.ConcurrentMap, loopIndex int, fullReferenceValue interface{}) (interface{}, error) {
	var realValue interface{}

	// if value has prefix $configure
	if strings.HasPrefix(fmt.Sprintf("%v", fullReferenceValue), "$configure") {
		splittedValue := strings.Split(fmt.Sprintf("%v", fullReferenceValue), separator) //$configure1.json, $request, $body[user][name]
		if len(splittedValue) != 3 {
			log.Error("referenced syntax wrong for : ", fullReferenceValue)
			log.Error(splittedValue)
			return nil, errors.New("referenced syntax wrong for : " + fullReferenceValue.(string))
		}

		var wrapper *model.Wrapper
		if tmp, ok := takeFrom.Get(splittedValue[0]); ok {
			wrapper = tmp.(*model.Wrapper)
			if splittedValue[1] == "$request" {
				// get the request from fields
				realValue = GetFromHalfReferenceValue(splittedValue[2], wrapper.Request, loopIndex)
				return realValue, nil
			}
			// get the response from fields
			realValue = GetFromHalfReferenceValue(splittedValue[2], wrapper.Response, loopIndex)
			return realValue, nil
		}

		// return empty reference
		return realValue, nil

	}

	return fullReferenceValue, nil
}

// AddToWrapper is a function that will add value to the specified key to a map
func AddToWrapper(commands map[string]interface{}, separator string, mapToBeAdded map[string]interface{}, takeFrom *cmap.ConcurrentMap, loopIndex int) map[string]interface{} {
	for key, value := range commands {
		// get real value from full reference value like $configure1.json-$request-$body[user][name]
		realValue, err := GetFromFullReferenceValue(separator, takeFrom, loopIndex, value)
		if err != nil {
			return mapToBeAdded
		}
		listTraverseKey := strings.Split(key, ".")
		mutex.Lock()
		mapToBeAdded = AddRecursive(listTraverseKey, realValue, mapToBeAdded, 0).(map[string]interface{})
		mutex.Unlock()
	}
	return mapToBeAdded
}

// ModifyWrapper is a function that will modify value based from specific key
func ModifyWrapper(commands map[string]interface{}, separator string, mapToBeModified map[string]interface{}, takeFrom *cmap.ConcurrentMap, loopIndex int) map[string]interface{} {
	for key, value := range commands {
		realValue, err := GetFromFullReferenceValue(separator, takeFrom, loopIndex, value)
		if err != nil {
			log.Error(" error in function ModifyWrapper : ", err.Error())
			return mapToBeModified
		}
		listTraverseKey := strings.Split(key, ".")
		mutex.Lock()
		result := ModifyRecursive(listTraverseKey, realValue, mapToBeModified, 0)
		if result != nil {
			mapToBeModified = result.(map[string]interface{})
		} else {
			log.Error("result is nil for value : ", value)
		}
		mutex.Unlock()
	}
	return mapToBeModified
}
