package model

import (
	"github.com/labstack/echo/v4"
	cmap "github.com/orcaman/concurrent-map"
)

type (
	CustomContext struct {
		echo.Context
		DefinedRoute         *Route
		FullProjectDirectory string
		BaseProject          Base
		MapWrapper           *cmap.ConcurrentMap
	}

	CustomResponse struct {
		StatusCode int
		Header     map[string]interface{}
		Body       map[string]interface{}
		Error      error
	}

	CLogicItem struct {
		Rule            interface{} `json:"rule"`
		Data            interface{} `json:"data"`
		NextSuccess     string      `json:"next_success"`
		Response        Command     `json:"response"`
		NextFailure     string      `json:"next_failure"`
		FailureResponse Command     `json:"failure_response"`
	}

	Route struct {
		Path             string `json:"path"`
		ProjectDirectory string `json:"project_directory"`
		Type             string `json:"type"`
		Method           string `json:"method"`
	}

	Base struct {
		MaxCircular      int     `json:"project_max_circular"`
		CircularResponse Command `json:"circular_response"`
	}

	Configure struct {
		ListStatusCodeSuccess []int   `json:"list_status_code_success"`
		Request               Command `json:"request"`
		Response              Command `json:"response"`
	}

	ConfigureItem struct {
		Loop            string       `json:"loop"`
		FileName        string       `json:"file_name"`
		Alias           string       `json:"alias"`
		CLogics         []CLogicItem `json:"c_logics"`
		FailureResponse Command      `json:"failure_response"`
	}

	Serial struct {
		Configures []ConfigureItem `json:"configures"`
	}

	Parallel struct {
		Configures      []ConfigureItem `json:"configures"`
		FailureResponse Command         `json:"failure_response"`
		CLogics         []CLogicItem    `json:"c_logics"`
	}

	Command struct {
		StatusCode      int                    `json:"status_code"`
		DestinationPath string                 `json:"destination_path"`
		DestinationUrl  string                 `json:"destination_url"`
		Method          string                 `json:"method"`
		Transform       string                 `json:"transform"`
		LogBeforeModify map[string]interface{} `json:"log_before_modify"`
		LogAfterModify  map[string]interface{} `json:"log_after_modify"`
		Adds            Fields                 `json:"adds"`
		Deletes         DeleteFields           `json:"deletes"`
		Modifies        Fields                 `json:"modifies"`
		CLogics         []CLogicItem           `json:"c_logics"`
	}

	Wrapper struct {
		Configure Configure
		Request   cmap.ConcurrentMap
		Response  cmap.ConcurrentMap
	}

	Fields struct {
		Param  map[string]interface{}
		Header map[string]interface{} `json:"header"`
		Body   map[string]interface{} `json:"body"`
		Query  map[string]interface{} `json:"query"`
	}

	DeleteFields struct {
		Header []string `json:"header"`
		Body   []string `json:"body"`
		Query  []string `json:"query"`
	}
	// errorString is a trivial implementation of error.
	ErrorString struct {
		s string
	}
)
