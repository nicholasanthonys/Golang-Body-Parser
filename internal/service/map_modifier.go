package service

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
)

// AddRecursive is a function that do the add key-value based on the listtraverse
func AddRecursive(listTraverse []string, value string, in interface{}, index int) interface{} {

	if index == len(listTraverse)-1 {

		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			in.(map[string]interface{})[listTraverse[index]] = value
		}

		return in
	}

	if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
		//* allocate new map if map[key] null
		if in.(map[string]interface{})[listTraverse[index]] == nil {
			logrus.Warn("map string interface ", listTraverse[index], " is nil")
			in.(map[string]interface{})[listTraverse[index]] = make(map[string]interface{})
		}
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
func checkValue(value interface{}, requestFromUser model.Fields, arrRes []map[string]interface{}) interface{} {
	var realValue interface{}
	vt := reflect.TypeOf(value).Kind()

	if reflect.String == vt {
		//*validate if value has $ or not
		listTraverseVal, destination := validateValue(fmt.Sprintf("%v", value))
		logrus.Info("list traveres value is ", listTraverseVal)
		if listTraverseVal != nil {
			if destination == "body" {
				realValue = getValue(listTraverseVal, requestFromUser.Body, 0)
			} else if destination == "header" {
				realValue = getValue(listTraverseVal, requestFromUser.Header, 0)
			} else if destination == "query" {
				realValue = getValue(listTraverseVal, requestFromUser.Query, 0)
			} else if destination == "response" {
				logrus.Info("list traversal response is ", listTraverseVal)
				tempSplit := strings.Split(listTraverseVal[0], "")
				logrus.Info("index is ")
				index, _ := strconv.Atoi(tempSplit[0])
				logrus.Info("arr ressis ")
				logrus.Info(arrRes[index])
				logrus.Info("eliminate index list traversal")
				listTraverseVal = listTraverseVal[1:]
				logrus.Info("list traversal become")
				logrus.Info(listTraverseVal)
				logrus.Info("traverse response")
				realValue = getValue(listTraverseVal, arrRes[index], 0)
			}
		} else {
			realValue = value
		}

	} else {
		logrus.Info("return real value")
		realValue = value
	}
	return realValue
}

//DoCommandConfigureBody is a wrapper function to do Add, Deletion and Modify for body
func DoCommandConfigureBody(command model.Command, requestFromUser model.Fields, arrRes []map[string]interface{}) {

	//*Do add
	for key, value := range command.Adds.Body {
		realValue := checkValue(value, requestFromUser, arrRes)
		listTraverseKey := strings.Split(key, ".")

		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Body, 0)
	}

	//* Do Deletion
	for _, key := range command.Deletes.Body {
		listTraverse := strings.Split(key, ".")
		DeleteRecursive(listTraverse, requestFromUser.Body, 0)
	}

	//*Do Modify
	for key, value := range command.Modifies.Body {
		realValue := checkValue(value, requestFromUser, arrRes)
		listTraverseKey := strings.Split(key, ".")
		ModifyRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Body, 0)
	}
}

// DoCommandConfigureHeader is a wrapper function that do add, modify, delete for header
func DoCommandConfigureHeader(command model.Command, requestFromUser model.Fields, arrRes []map[string]interface{}) {
	//*Add to map header
	for key, value := range command.Adds.Header {
		realValue := checkValue(value, requestFromUser, arrRes)
		listTraverseKey := strings.Split(key, ".")
		logrus.Info("list traversal is ", listTraverseKey)
		logrus.Info("key is ", key)
		logrus.Info("value is ", value)
		logrus.Info("real value add header is ", realValue)
		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Header, 0)

	}

	//*Delete
	for _, key := range command.Deletes.Header {
		//header.Del(key)
		delete(requestFromUser.Header, key)
	}

	//* Modify
	for key, value := range command.Modifies.Header {

		existValue := fmt.Sprintf("%s", requestFromUser.Header[strings.Title(key)])
		if len(existValue) > 0 {
			realValue := checkValue(value, requestFromUser, arrRes)
			requestFromUser.Header[key] = realValue

		}
	}

}

func getValue(listTraverse []string, in interface{}, index int) interface{} {

	if len(listTraverse) > 0 {

		if index == len(listTraverse)-1 {

			rt := reflect.TypeOf(in)

			switch rt.Kind() {
			case reflect.Slice:
				var indexInt int

				//*check type slice element
				et := reflect.TypeOf(in).Elem().Kind()
				indexInt, _ = strconv.Atoi(listTraverse[index])
				if et == reflect.Interface {
					return in.([]interface{})[indexInt]
				}
				return in.([]string)[indexInt]

			case reflect.Map:

				logrus.Info(in, " is map")
				logrus.Info("returned in map is ", in.(map[string]interface{})[listTraverse[index]])
				//logrus.Info(in, " is map ", rt.Elem())
				return in.(map[string]interface{})[listTraverse[index]]
			default:
				//logrus.Info(in, "is something else entirely")
				return in
			}

		}

		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			logrus.Info(in, " is map string interface")
			logrus.Info("list traverse index is ", listTraverse[index])
			logrus.Info("map nya ")
			logrus.Info(in.(map[string]interface{})[listTraverse[index]])
			//logrus.Warn("type is map string interface")
			//* allocate new map if map[key] null
			if in.(map[string]interface{})[listTraverse[index]] == nil {
				logrus.Info("returned in is ", in)
				return in
			}
			return getValue(listTraverse, in.(map[string]interface{})[listTraverse[index]], index+1)

		} else {
			//logrus.Warn(in, " not ", " map string interface")
			return nil
		}
	}

	return in
}

func validateValue(value string) ([]string, string) {

	listTraverse := make([]string, 0)
	var destination string

	if strings.HasPrefix(value, "$body") {
		destination = "body"
		value = string(value[5:])
	} else if strings.HasPrefix(value, "$header") {
		destination = "header"
		value = string(value[7:])
	} else if strings.HasPrefix(value, "$query") {
		destination = "query"
		value = string(value[6:])

	} else if strings.HasPrefix(value, "$response") {
		destination = "response"
		value = string(value[9:])
	} else {
		return nil, value
	}

	//*split become [tes],[tos]
	arraySplit := strings.Split(value, "")

	temp := ""
	for _, val := range arraySplit {
		if val != "[" {
			if val == "]" {
				//*push
				listTraverse = append(listTraverse, temp)
				temp = ""
			} else {
				//*
				temp += val
			}
		}

	}

	return listTraverse, destination

}

// DoCommandConfigureQuery is a wrapper function that do add, modify, delete for query
func DoCommandConfigureQuery(command model.Command, requestFromUser model.Fields, arrRes []map[string]interface{}) {
	//* Add
	for key, value := range requestFromUser.Query {
		realValue := checkValue(value, requestFromUser, arrRes)
		listTraverseKey := strings.Split(key, ".")

		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Query, 0)

	}

	//* Delete
	for _, key := range command.Deletes.Query {
		delete(requestFromUser.Query, key)
	}

	//* Modify
	for key, value := range command.Modifies.Query {
		existingValue := fmt.Sprintf("%s", requestFromUser.Query[key])
		if len(existingValue) > 0 {
			requestFromUser.Query[key] = value
		}
	}
}

//* if c request method
func DoCommand(command model.Command, requestFromUser model.Fields, arrRes []map[string]interface{}) {

	DoCommandConfigureHeader(command, requestFromUser, arrRes)
	DoCommandConfigureQuery(command, requestFromUser, arrRes)
	DoCommandConfigureBody(command, requestFromUser, arrRes)
}
