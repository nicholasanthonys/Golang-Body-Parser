package service

import (
	"github.com/sirupsen/logrus"
	"plugin"
)

func LoadFunctionFromModule(pluginPath string, functionReference string) (func(map[string]interface{}) ([]byte, error), error) {
	logrus.Info("plugin path is ", pluginPath+"/transform.so")

	plug, err := plugin.Open(pluginPath + "/transform.so")
	if err != nil {
		logrus.Warn("Unable to load plugin module")
		logrus.Warn(err.Error())
		return nil, err
	}

	logrus.Info("success read plugin")

	// lookup for an exported function
	functionSymbol, err := plug.Lookup(functionReference)
	if err != nil {
		logrus.Warn(err.Error())
		return nil, err
	}

	//cast function to the correct type
	myFunction := functionSymbol.(func(map[string]interface{}) ([]byte, error))
	return myFunction, nil

}
