package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"strings"
)

func AddRecursive(listTraverse []string, value string, in interface{}, index int) interface{} {
	if index == len(listTraverse)-1 {
		in.(map[string]interface{})[listTraverse[index]] = value
		return in
	}
	//* allocate new map if map[key] null
	if in.(map[string]interface{})[listTraverse[index]] == nil {
		in.(map[string]interface{})[listTraverse[index]] = make(map[string]interface{})
	}
	in.(map[string]interface{})[listTraverse[index]] = AddRecursive(listTraverse, value, in.(map[string]interface{})[listTraverse[index]], index+1)
	return in.(map[string]interface{})
}

func ModifyRecursive(listTraverse []string, value string, in interface{}, index int) interface{} {
	if index == len(listTraverse)-1 {
		if in.(map[string]interface{})[listTraverse[index]] == nil {
			return nil
		}
		in.(map[string]interface{})[listTraverse[index]] = value
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

func DoCommandConfigure(command model.Command, requestFromUser model.Fields) {

	//*Do add
	for key, value := range command.Adds.Body {
		listTraverse := strings.Split(key, ".")
		AddRecursive(listTraverse, fmt.Sprintf("%v", value), requestFromUser.Body, 0)

	}
	//*Do Modify
	for key, value := range command.Modifies.Body {
		listTraverse := strings.Split(key, ".")

		ModifyRecursive(listTraverse, fmt.Sprintf("%v", value), requestFromUser.Body, 0)
	}

	//* Do Deletion
	for _, key := range command.Deletes.Body {
		listTraverse := strings.Split(key, ".")
		DeleteRecursive(listTraverse, requestFromUser.Body, 0)
	}

}
