package test

import (
	"github.com/joho/godotenv"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

var URL = ""
var log = logrus.New()

func init() {
	err := godotenv.Load("../.env.testing")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	URL = "http://" + os.Getenv("APP_URL") + ":" + os.Getenv("APP_PORT")
	logrus.Info("init triggered")
}

func TestWithoutBody(t *testing.T) {

	req, _ := http.NewRequest("POST", URL+"/smsotp/generate/3?tesquery=abcd", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}
	if res.StatusCode != http.StatusLoopDetected {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}

	expected := `{
		"message":"Circular response",
		"user": {
			"configure0_query": "kopo",
			"id": 0,
			"name": "Peter Parker",
			"transaction_id": "3",
			"tes" : "from configure.example directory",
			"favorite_cars" : ""
		}
	}`

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}
	equal, err := util.JSONBytesEqual([]byte(expected), responseByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expected, string(responseByte), "should be equal")
	}

}

func TestWithBody(t *testing.T) { //*SERIAL
	json := `{"user" : { "name" : "nicholas", "cars" : ["honda", "fiat", "daihatsu", "toyota"]}}`
	req, err := http.NewRequest("POST", URL+"/smsotp/generate/3", strings.NewReader(json))
	if err != nil {
		t.Errorf("Error constructing Request %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)

	if assert.NoError(t, err) {
		expected := `{
				"user": {
					"id": 0,
					"name": "Peter Parker",
					"favorite_cars":"honda",
 					"transaction_id": "3",
					"configure0_query": "kopo",
					"tes" : "from configure.example directory"
				}
			}`

		responseByte, err := ioutil.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, "error read response", err.Error())
		}

		equal, err := util.JSONBytesEqual([]byte(expected), responseByte)
		if err != nil {
			assert.Error(t, err, "error compare json byte")
		}
		if !equal {
			assert.Equal(t, expected, string(responseByte), "should be equal")
		}

	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}

}

func TestWrongMethod(t *testing.T) {
	//* SERIAL
	req, _ := http.NewRequest("PUT", URL+"/smsotp/generate/3?tesquery=abcd", nil)
	client := http.Client{}
	res, err := client.Do(req)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected %d, received %d", http.StatusMethodNotAllowed, res.StatusCode)
	}
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("error reading response byte")
	}
}

func TestSerialToParallel(t *testing.T) {
	// turn on mirror server to get loop from response
	json := `{
 	  "phone_numbers" : [
       	"123",
       	"456",
       	"789"
	  ],
       "content" : "selamat anda menang",
       "loop" : 3
	}`

	req, err := http.NewRequest("POST", URL+"/another-project", strings.NewReader(json))
	client := http.Client{}
	res, err := client.Do(req)

	if assert.NoError(t, err) {
		expected := `{
				"user": {
        			"anime": [
            			"naruto",
            			"one piece",
            			"dragon ball"
        			],
        		"cars": [
            		"Mazda",
            		"Aston  Martin",
            		"Ferrari"
        		],
        		"favorite_cars": "Mazda"
    			}
		}`

		responseByte, err := ioutil.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, "error read response", err.Error())
		}

		equal, err := util.JSONBytesEqual([]byte(expected), responseByte)
		if err != nil {
			assert.Error(t, err, "error compare json byte")
		}
		if !equal {
			assert.Equal(t, expected, string(responseByte), "should be equal")
		}
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}
}

func TestParallelToSerial(t *testing.T) {

	json := `{}`

	req, err := http.NewRequest("POST", URL+"/parallel-serial", strings.NewReader(json))
	client := http.Client{}
	res, err := client.Do(req)

	if assert.NoError(t, err) {
		expected := `{
    		"content_0": "Hello World from configure-0.json",
    		"content_1": "Hello World from configure-0.json",
    		"numbers": [
        		"123-456",
        		"234-567",
        		"345-678"
    	],
    		"test": "test key From add response configure-2.json"
		}`

		responseByte, err := ioutil.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, "error read response", err.Error())
		}

		equal, err := util.JSONBytesEqual([]byte(expected), responseByte)
		if err != nil {
			assert.Error(t, err, "error compare json byte")
		}
		if !equal {
			assert.Equal(t, expected, string(responseByte), "should be equal")
		}
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	}

}
