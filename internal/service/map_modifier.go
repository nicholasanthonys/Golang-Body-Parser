package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

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

func checkValue(value interface{}, requestFromUser model.Fields) interface{} {
	var realValue interface{}
	vt := reflect.TypeOf(value).Kind()

	if reflect.String == vt {
		//*validate if value has $ or not
		listTraverseVal, destination := validateValue(fmt.Sprintf("%v", value))

		if listTraverseVal != nil {
			if destination == "body" {
				realValue = getValue(listTraverseVal, requestFromUser.Body, 0)
			} else if destination == "header" {
				realValue = getValue(listTraverseVal, requestFromUser.Header, 0)
			} else if destination == "query" {
				realValue = getValue(listTraverseVal, requestFromUser.Query, 0)
			}
		} else {
			realValue = value
		}

	} else {
		realValue = value
	}
	return realValue
}

func DoCommandConfigureBody(command model.Command, requestFromUser model.Fields) {

	//*Do add
	for key, value := range command.Adds.Body {
		realValue := checkValue(value, requestFromUser)
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
		realValue := checkValue(value, requestFromUser)
		listTraverseKey := strings.Split(key, ".")
		ModifyRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Body, 0)
	}
}

func DoCommandConfigureHeader(command model.Command, requestFromUser model.Fields, header *http.Header) {
	//*Add to map header
	for key, value := range command.Adds.Header {
		realValue := checkValue(value, requestFromUser)
		listTraverseKey := strings.Split(key, ".")

		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Header, 0)

	}

	//actually set the header based on map header
	for key, value := range requestFromUser.Header {
		if key != "Content-Type" {

			vt := reflect.TypeOf(value).Kind()
			if vt == reflect.Slice {
				header.Add(key, fmt.Sprintf("%v", value.([]string)[0]))
			} else {
				header.Add(key, fmt.Sprintf("%s", value))
			}

		}

	}

	//*Delete
	for _, key := range command.Deletes.Header {
		header.Del(key)
	}

	//* Modify
	for key, value := range command.Modifies.Header {
		if len(header.Get(key)) > 0 {
			header.Set(key, fmt.Sprintf("%v", value))
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

				//logrus.Info(in, " is map ", rt.Elem())
				return in.(map[string]interface{})[listTraverse[index]]
			default:
				//logrus.Info(in, "is something else entirely")
				return in
			}

		}

		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			//logrus.Warn("type is map string interface")
			//* allocate new map if map[key] null
			if in.(map[string]interface{})[listTraverse[index]] == nil {
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

func DoCommandConfigureQuery(command model.Command, requestFromUser model.Fields, q *url.Values) {
	//* Add
	for key, value := range requestFromUser.Query {
		realValue := checkValue(value, requestFromUser)
		listTraverseKey := strings.Split(key, ".")

		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Query, 0)

		q.Set(key, fmt.Sprintf("%v", realValue))
		logrus.Info("q get is ", q.Get(key))
	}

	//* Delete
	for _, key := range command.Deletes.Query {
		q.Del(key)
	}

	//* Modify
	for key, value := range command.Modifies.Query {
		if len(q.Get(key)) > 0 {
			q.Set(key, fmt.Sprintf("%v", value))
		}
	}
}
