package service

import (
	"bytes"
	"encoding/json"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func Send(configure model.Configure, requestFromUser map[string]interface{}) {

	method := configure.Method
	url := configure.DestinationUrl
	jsonBody, _ := json.MarshalIndent(requestFromUser, "", "\t")
	//buffer := bytes.NewBuffer(jsonBody)

	//var jsonStr = []byte(`{"title" : "buy books"}`)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		logrus.Fatal("Error response")
		logrus.Fatal(err.Error())
	}

	respByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Fatal("Error read body")
	}
	logrus.Info("response is")
	logrus.Info(string(respByte))
}
