package test

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/request"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetRequestToWrapper(t *testing.T) {
	requestJson := `{
 	  "phone_numbers" : [
       	"123",
       	"456",
       	"789"
	  ],
       "lucky_number" : 7,
		"user" :  {
			"name" : "nicholas"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	alias := "$configure_first_configure"

	tmpMapWrapper := cmap.New()
	c := model.CustomContext{
		Context:              e.NewContext(req, rec),
		DefinedRoute:         nil,
		FullProjectDirectory: "",
		BaseProject:          model.Base{},
		MapWrapper:           &tmpMapWrapper,
	}
	wrapper := model.Wrapper{
		Configure: model.Configure{},
		Request:   cmap.New(),
		Response:  cmap.New(),
	}

	err := request.SetRequestToWrapper(alias, &c, &wrapper)
	if err != nil {
		assert.Error(t, err, "error when calling function SetRequestToWrapper")
	}

	actualBody := make(map[string]interface{})
	if tmp, ok := wrapper.Request.Get("body"); ok {
		log.Info("tmp is ", tmp)
		actualBody = tmp.(map[string]interface{})
	}
	log.Info("actual body is ", actualBody)
	resByte, err := json.Marshal(actualBody)
	if err != nil {
		assert.Error(t, err, " error when marshal actualBody")
	}

	equal, err := util.JSONBytesEqual([]byte(requestJson), resByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}

	if !equal {
		assert.Equal(t, requestJson, string(resByte), "should be equal")
	}

}
