package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
)

func Add(command model.Command, requestFromUser map[string]interface{}) {
	logrus.Info("request form user is")
	logrus.Info(requestFromUser)
	for key, value := range command.Adds {
		logrus.Info("Key: ", key, " Value: "+
			"", value)
		//* if key is not exist, add to map
		if requestFromUser[key] == nil {
			requestFromUser[key] = value
		}

	}
}

func Modify(command model.Command, requestFromUser map[string]interface{}) {
	fmt.Println("Modify")
	//*modify key from configure.json to requestFromUser Map
	for key, value := range command.Modifies {
		fmt.Println("Key:", key, "Value:", value)
		if requestFromUser[key] != nil {
			requestFromUser[key] = value
		}

	}
}

func Delete(command model.Command, requestFromUser map[string]interface{}) {
	//*delete dari map
	fmt.Println("Deletes")
	for key, value := range command.Deletes {
		fmt.Println("key is ", key, "value is ", value)
		delete(requestFromUser, value)

	}
}

func DoCommandConfigure(command model.Command, requestFromUser map[string]interface{}) {
	Add(command, requestFromUser)
	Delete(command, requestFromUser)
	Modify(command, requestFromUser)
}
