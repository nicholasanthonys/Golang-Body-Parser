package service

import (
	"fmt"
	"github.com/labstack/echo"
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
func checkValue(c echo.Context, value interface{}, requestFromUser model.Fields, arrRes []model.Wrapper) interface{} {
	//*declare empty result
	var realValue interface{}
	//* check the type of the value
	vt := reflect.TypeOf(value).Kind()

	if reflect.String == vt {
		//* We Call Sanitizevalue to clear the value from the square bracket and the Dollar Sign
		listTraverseVal, destination := SanitizeValue(fmt.Sprintf("%v", value))
		if listTraverseVal != nil {
			if destination == "body" {
				realValue = GetValue(listTraverseVal, requestFromUser.Body, 0)
			} else if destination == "header" {
				realValue = GetValue(listTraverseVal, requestFromUser.Header, 0)
			} else if destination == "query" {
				realValue = GetValue(listTraverseVal, requestFromUser.Query, 0)
			} else if destination == "response" {
				tempSplit := strings.Split(listTraverseVal[0], "")
				index, _ := strconv.Atoi(tempSplit[0])
				listTraverseVal = listTraverseVal[1:]
				//* masih hardcode
				realValue = GetValue(listTraverseVal, arrRes[index].Response.Body, 0)
			} else if destination == "path" {
				realValue = c.Param(listTraverseVal[0])
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
func DoCommandConfigureBody(c echo.Context, command model.Command, requestFromUser model.Fields, arrRes []model.Wrapper) {
	//* Add key
	for key, value := range command.Adds.Body {
		//*get the value
		realValue := checkValue(c, value, requestFromUser, arrRes)
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
		realValue := checkValue(c, value, requestFromUser, arrRes)
		listTraverseKey := strings.Split(key, ".")
		ModifyRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Body, 0)
	}
}

// DoCommandConfigureHeader is a wrapper function that do add, modify, delete for header
func DoCommandConfigureHeader(c echo.Context, command model.Command, requestFromUser model.Fields, arrRes []model.Wrapper) {
	//*Add to map header
	for key, value := range command.Adds.Header {
		realValue := checkValue(c, value, requestFromUser, arrRes)
		listTraverseKey := strings.Split(key, ".")
		AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), requestFromUser.Header, 0)

	}

	//*Delete
	for _, key := range command.Deletes.Header {
		delete(requestFromUser.Header, key)
	}

	//* Modify
	for key, value := range command.Modifies.Header {
		existValue := fmt.Sprintf("%s", requestFromUser.Header[strings.Title(key)])
		if len(existValue) > 0 {
			realValue := checkValue(c, value, requestFromUser, arrRes)
			requestFromUser.Header[key] = realValue
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
func DoCommandConfigureQuery(c echo.Context, command model.Command, requestFromUser model.Fields, arrRes []model.Wrapper) {
	//* Add
	for key, value := range requestFromUser.Query {
		realValue := checkValue(c, value, requestFromUser, arrRes)
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

//*DoCommand is a function that will do the command from configure.json for Header, Query, and Body
//* Here, we call DoCommandConfigure for each Header, Query, and Body
func DoCommand(c echo.Context, command model.Command, requestFromUser model.Fields, arrRes []model.Wrapper) {

	DoCommandConfigureHeader(c, command, requestFromUser, arrRes)
	DoCommandConfigureQuery(c, command, requestFromUser, arrRes)
	DoCommandConfigureBody(c, command, requestFromUser, arrRes)
}
