package service

import (
	"github.com/sirupsen/logrus"
	"os"
	"plugin"
)

func LoadFunctionFromModule(functionReference string) func(map[string]interface{}) ([]byte, error) {
	plug, err := plugin.Open("./plugin/transform.so")
	if err != nil {
		logrus.Warn("Unable to load plugin module")
		logrus.Warn(err.Error())
		os.Exit(1)

	}

	// lookup for an exported function
	functionSymbol, err := plug.Lookup(functionReference)
	if err != nil {
		logrus.Warn(err.Error())
		os.Exit(1)
	}

	//cast function to the corect type
	myFunction := functionSymbol.(func(map[string]interface{}) ([]byte, error))
	return myFunction

}
