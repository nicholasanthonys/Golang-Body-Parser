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

var URL = ""
var log = logrus.New()
var dirName = ""

func init() {
	err := godotenv.Load("../.env.testing")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	URL = "http://" + os.Getenv("APP_URL") + ":" + os.Getenv("APP_PORT")
	logrus.Info("init triggered")

	dirName = os.Getenv("CONFIGURES_DIRECTORY_TESTING_NAME")
}

func TestSingleSerial_WithoutLogicRequestAndWithoutLogicFinal(t *testing.T) {
	projectDir := dirName + "/test-1"

	req, _ := http.NewRequest("GET", URL+"/test-1?movie_id=550", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}

	assert.Equalf(t, http.StatusBadRequest, res.StatusCode, " Expected %s but got %s", http.StatusBadRequest, res.StatusCode)

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}

	expectedByte := util.ReadJsonFile(projectDir + "/test-1_expected.json")
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