package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func init() {
	err := godotenv.Load("../.env.testing")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	URL = "http://" + os.Getenv("APP_URL") + ":" + os.Getenv("APP_PORT")
	logrus.Info("init triggered")

	dirName = os.Getenv("CONFIGURES_DIRECTORY_TESTING_NAME")
}

func TestSingleSerial_OnlyTrueLogicSerialResponse(t *testing.T) {
	projectDir := dirName + "/test-3.1"

	req, _ := http.NewRequest("GET", URL+"/test-3-1?movie_id=550", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}

	assert.Equalf(t, http.StatusCreated, res.StatusCode, " Expected %s but got %s", http.StatusCreated, res.StatusCode)

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}

	expectedByte := util.ReadJsonFile(projectDir + "/test-3.1_expected.json")
	equal, err := util.JSONBytesEqual(expectedByte, responseByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		var jsonExpectedPretty bytes.Buffer
		var jsonResponseBodyPretty bytes.Buffer
		err := json.Indent(&jsonExpectedPretty, expectedByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		err = json.Indent(&jsonResponseBodyPretty, responseByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		log.Infof("expected : ")
		fmt.Printf("%s \n", jsonExpectedPretty.Bytes())
		log.Infof("actual : ")
		fmt.Printf("%s \n", jsonResponseBodyPretty.Bytes())
		assert.Equal(t, string(jsonExpectedPretty.Bytes()), string(jsonResponseBodyPretty.Bytes()), "should be equal")
	}

}

func TestSingleSerial_OnlyTrueLogicWithoutSerialResponse(t *testing.T) {
	projectDir := dirName + "/test-3.2"

	req, _ := http.NewRequest("GET", URL+"/test-3-2?movie_id=550", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}

	assert.Equalf(t, http.StatusCreated, res.StatusCode, " Expected %s but got %s", http.StatusCreated, res.StatusCode)

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}

	expectedByte := util.ReadJsonFile(projectDir + "/test-3.2_expected.json")
	equal, err := util.JSONBytesEqual(expectedByte, responseByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		var jsonExpectedPretty bytes.Buffer
		var jsonResponseBodyPretty bytes.Buffer
		err := json.Indent(&jsonExpectedPretty, expectedByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		err = json.Indent(&jsonResponseBodyPretty, responseByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		log.Infof("expected : ")
		fmt.Printf("%s \n", jsonExpectedPretty.Bytes())
		log.Infof("actual : ")
		fmt.Printf("%s \n", jsonResponseBodyPretty.Bytes())
		assert.Equal(t, string(jsonExpectedPretty.Bytes()), string(jsonResponseBodyPretty.Bytes()), "should be equal")
	}

}

func TestSingleSerial_FalseLogicWithSerialFailureResponse(t *testing.T) {
	projectDir := dirName + "/test-3.3"

	req, _ := http.NewRequest("GET", URL+"/test-3-3?movie_id=550", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}

	assert.Equalf(t, http.StatusBadRequest, res.StatusCode,
		" Expected %s but got %s",
		http.StatusInternalServerError,
		res.StatusCode)

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}

	expectedByte := util.ReadJsonFile(projectDir + "/test-3.4_expected.json")
	equal, err := util.JSONBytesEqual(expectedByte, responseByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		var jsonExpectedPretty bytes.Buffer
		var jsonResponseBodyPretty bytes.Buffer
		err := json.Indent(&jsonExpectedPretty, expectedByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		err = json.Indent(&jsonResponseBodyPretty, responseByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		log.Infof("expected : ")
		fmt.Printf("%s \n", jsonExpectedPretty.Bytes())
		log.Infof("actual : ")
		fmt.Printf("%s \n", jsonResponseBodyPretty.Bytes())
		assert.Equal(t, string(jsonExpectedPretty.Bytes()), string(jsonResponseBodyPretty.Bytes()), "should be equal")
	}

}

func TestSingleSerial_FalseLogicWithoutSerialFailureResponse(t *testing.T) {
	projectDir := dirName + "/test-3.4"

	req, _ := http.NewRequest("GET", URL+"/test-3-4?movie_id=550", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}

	assert.Equalf(t, http.StatusInternalServerError, res.StatusCode,
		" Expected %s but got %s",
		http.StatusInternalServerError,
		res.StatusCode)

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}

	expectedByte := util.ReadJsonFile(projectDir + "/test-3.4_expected.json")
	equal, err := util.JSONBytesEqual(expectedByte, responseByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		var jsonExpectedPretty bytes.Buffer
		var jsonResponseBodyPretty bytes.Buffer
		err := json.Indent(&jsonExpectedPretty, expectedByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		err = json.Indent(&jsonResponseBodyPretty, responseByte, "", "\t")
		if err != nil {
			assert.Error(t, err, "should not error")
		}

		log.Infof("expected : ")
		fmt.Printf("%s \n", jsonExpectedPretty.Bytes())
		log.Infof("actual : ")
		fmt.Printf("%s \n", jsonResponseBodyPretty.Bytes())
		assert.Equal(t, string(jsonExpectedPretty.Bytes()), string(jsonResponseBodyPretty.Bytes()), "should be equal")
	}

}
