package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func AddRecursive(listTraverse []string, value string, in interface{}, index int) interface{} {

	if index == len(listTraverse)-1 {
		logrus.Info("last ", listTraverse[index])
		logrus.Warn("reflect type of")
		logrus.Warn(reflect.TypeOf(in))
		if fmt.Sprintf("%v", reflect.TypeOf(in)) == "map[string]interface {}" {
			in.(map[string]interface{})[listTraverse[index]] = value
		}

		return in
	}
	//* allocate new map if map[key] null
	if in.(map[string]interface{})[listTraverse[index]] == nil {
		logrus.Warn("map strin ginterface ", listTraverse[index], " is nil")
		in.(map[string]interface{})[listTraverse[index]] = make(map[string]interface{})
	} else {
		logrus.Warn("map strin ginterface ", listTraverse[index], " not nil")
	}
	in.(map[string]interface{})[listTraverse[index]] = AddRecursive(listTraverse, value, in.(map[string]interface{})[listTraverse[index]], index+1)
	return in.(map[string]interface{})
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

	if in.(map[string]interface{})[listTraverse[index]] != nil {
		ModifyRecursive(listTraverse, value, in.(map[string]interface{})[listTraverse[index]], index+1)
		return in.(map[string]interface{})
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

func DoCommandConfigureBody(command model.Command, requestFromUser model.Fields) {

	//*Do add
	for key, value := range command.Adds.Body {
		listTraverse := strings.Split(key, ".")
		AddRecursive(listTraverse, fmt.Sprintf("%v", value), requestFromUser.Body, 0)

	}

	//* Do Deletion
	for _, key := range command.Deletes.Body {
		listTraverse := strings.Split(key, ".")
		DeleteRecursive(listTraverse, requestFromUser.Body, 0)
	}

	//*Do Modify
	for key, value := range command.Modifies.Body {
		listTraverse := strings.Split(key, ".")

		ModifyRecursive(listTraverse, fmt.Sprintf("%v", value), requestFromUser.Body, 0)
	}
}

func DoCommandConfigureHeader(command model.Command, header *http.Header) {
	//*Add
	for key, value := range command.Adds.Header {
		header.Add(key, fmt.Sprintf("%v", value))
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

func DoCommandConfigureQuery(command model.Command, q *url.Values) {
	//* Add
	for key, value := range command.Adds.Query {
		q.Set(key, fmt.Sprintf("%v", value))
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
