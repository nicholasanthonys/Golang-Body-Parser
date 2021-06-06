package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestParallel_To_Serial(t *testing.T) {
	projectDir := dirName + "/test-8.1_parallel_to_serial"
	mapBody := map[string]interface{}{
		"content": "selamat anda menang",
		"phone_numbers": []string{
			"123456789",
			"2456789",
			"3456789",
		},
	}
	body, err := service.Transform("ToJson", mapBody)

	if err != nil {
		assert.NoError(t, err, "Should not error. Error is %s", err)
	}

	req, _ := http.NewRequest("POST", URL+"/test-8-1", body)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected nil, received %s", err.Error())
	}

	assert.Equalf(t, http.StatusCreated, res.StatusCode, " Expected %s but got %s", http.StatusCreated,
		res.StatusCode)

	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, "should not error")
	}

	expectedByte := util.ReadJsonFile(projectDir + "/test-8.1_expected.json")
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
