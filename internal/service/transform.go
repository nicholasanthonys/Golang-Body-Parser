package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
)

func Add(configure model.Configure, requestFromUser map[string]interface{}) {
	for key, value := range configure.Request.Adds {
		fmt.Println("Key:", key, "Value:", value)
		//* if key is not exist, add to map
		if requestFromUser[key] == nil {
			requestFromUser[key] = value
		}

	}
}

func Modify(configure model.Configure, requestFromUser map[string]interface{}) {
	fmt.Println("Modify")
	//*modify key from configure.json to requestFromUser Map
	for key, value := range configure.Request.Modifies {
		fmt.Println("Key:", key, "Value:", value)
		if requestFromUser[key] != nil {
			requestFromUser[key] = value
		}

	}
}

func Delete(configure model.Configure, requestFromUser map[string]interface{}) {
	//*delete dari map
	fmt.Println("Deletes")
	for key, value := range configure.Request.Deletes {
		fmt.Println("key is ", key, "value is ", value)
		delete(requestFromUser, value)

	}
}
