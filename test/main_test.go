package test

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var URL = ""

func init() {
	URL = "http://localhost:8888"
	logrus.Info("init triggered")
}

func TestWithoutBody(t *testing.T) {

	req, _ := http.NewRequest("POST", URL+"/serial/persons/3/transactions/20?tesquery=abcd", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}
}

func TestWithBody(t *testing.T) { //*SERIAL
	json := `{"user" : { "name" : "nicholas", "cars" : ["honda", "fiat", "daihatsu", "toyota"]}}`
	req, err := http.NewRequest("POST", URL+"/serial/persons/3/transactions/20", strings.NewReader(json))
	if err != nil {
		t.Errorf("Error constructing Request %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)

	if assert.NoError(t, err) {
		expected :=
			`{
				"user": {
					"favorite_cars": "honda",
					"id": "0",
					"name": "Peter Parker",
 					"transaction_id": "20",
					"configure0_query": "kopo"
				}
			}`

		expectedJson, err := service.FromJson([]byte(expected))
		responseByte, err := ioutil.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, "error read response", err.Error())
		}
		actualResponse, _ := service.FromJson(responseByte)
		assert.Exactlyf(t, expectedJson, actualResponse, "actual-expected")
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}

	//*PARALLEL
	req, err = http.NewRequest("POST", URL+"/serial/persons/3/transactions/20", strings.NewReader(json))
	if err != nil {
		t.Errorf("Error constructing Request %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client = http.Client{}
	res, err = client.Do(req)

	if assert.NoError(t, err) {
		expected :=
			`{
				"user": {
					"favorite_cars": "honda",
					"id": "0",
					"name": "Peter Parker",
					"transaction_id": "20",
					"configure0_query": "kopo"
				}
			}`

		expectedJson, err := service.FromJson([]byte(expected))
		responseByte, err := ioutil.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, "error read response", err.Error())
		}
		actualResponse, _ := service.FromJson(responseByte)
		assert.Exactlyf(t, expectedJson, actualResponse, "actual-expected")
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}
}

func TestWithoutHeader(t *testing.T) {
	//* SERIAL
	req, _ := http.NewRequest("PUT", URL+"/serial/persons/3/transactions/20?tesquery=abcd", nil)
	client := http.Client{}
	res, err := client.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("error reading response byte")
	}
}
