package main

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/service"
	"github.com/steinfletcher/apitest"
	"net/http"
	"testing"
)

func TestSerial(t *testing.T) {

	//var req *http.Request
	//
	//
	//req,_ = http.NewRequest("GET","http://localhost:8888/serial",nil)
	//req.Header.Set("Content-Type","application/json")
	//
	//client := http.Client{}
	//res, err := client.Do(req)
	//if err != nil {
	//	t.Errorf("Expected nil, received %s", err.Error())
	//}
	//if res.StatusCode != http.StatusOK {
	//	t.Errorf("Expected %d, received %d", http.StatusOK, res.StatusCode)
	//}

	apitest.New().
		Handler(service.SetRouteHandler()).
		Post("http://localhost:8888/serial/tes").
		Header("Content-Type", "application/json").
		Expect(t).
		Status(http.StatusOK).
		End()
}
