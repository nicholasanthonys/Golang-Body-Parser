package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestSingleSerial_True_Logic_With_Response(t *testing.T) {
	projectDir := dirName + "/test-2.1"

	req, _ := http.NewRequest("GET", URL+"/test-2-1?movie_id=550", nil)
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

	expectedByte := util.ReadJsonFile(projectDir + "/test-2.1_expected.json")
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

func TestSingleSerial_True_Logic_Without_Response(t *testing.T) {
	projectDir := dirName + "/test-2.2"

	req, _ := http.NewRequest("GET", URL+"/test-2-2?movie_id=550", nil)
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

	expectedByte := util.ReadJsonFile(projectDir + "/test-2.2_expected.json")
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

func TestSingleSerial_False_Logic_Without_Failure_Response(t *testing.T) {
	projectDir := dirName + "/test-2.3"

	req, _ := http.NewRequest("GET", URL+"/test-2-3?movie_id=550", nil)
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

	expectedByte := util.ReadJsonFile(projectDir + "/test-2.3_expected.json")
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

func TestSingleSerial_False_Logic_With_Failure_Response(t *testing.T) {
	projectDir := dirName + "/test-2.4"

	req, _ := http.NewRequest("GET", URL+"/test-2-4?movie_id=550", nil)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}

	assert.Equalf(t, http.StatusUnauthorized, res.StatusCode,
		" Expected %s but got %s",
		http.StatusUnauthorized,
		res.StatusCode)

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}

	expectedByte := util.ReadJsonFile(projectDir + "/test-2.4_expected.json")
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
