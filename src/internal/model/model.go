package model

type (
	CLogicItem struct {
		Rule        interface{}            `json:"rule"`
		Data        map[string]interface{} `json:"data"`
		NextSuccess string                 `json:"next_success"`
		Response    Command                `json:"response"`
	}

	Route struct {
		Path             string `json:"path"`
		ProjectDirectory string `json:"project_directory"`
		Type             string `json:"type"`
		Method           string `json:"method"`
	}

	Routes []Route

	Configure struct {
		ConfigureBased        string  `json:"configure_based"`
		ListStatusCodeSuccess []int   `json:"list_status_code_success"`
		Request               Command `json:"request"`
		Response              Command `json:"response"`
	}

	ConfigureItem struct {
		FileName    string       `json:"file_name"`
		Alias       string       `json:"alias"`
		CLogics     []CLogicItem `json:"c_logics"`
		NextFailure Command      `json:"next_failure"`
	}

	Project struct {
		Configures []ConfigureItem `json:"configures"`
	}

	Command struct {
		StatusCode      int          `json:"status_code"`
		DestinationPath string       `json:"destination_path"`
		DestinationUrl  string       `json:"destination_url"`
		Method          string       `json:"method"`
		Transform       string       `json:"transform"`
		LogBeforeModify string       `json:"log_before_modify"`
		LogAfterModify  string       `json:"log_after_modify"`
		Adds            Fields       `json:"adds"`
		Deletes         DeleteFields `json:"deletes"`
		Modifies        Fields       `json:"modifies"`
	}

	Wrapper struct {
		Configure Configure
		Request   Fields
		Response  Fields
	}

	Fields struct {
		StatusCode int
		Param      map[string]interface{}
		Header     map[string]interface{} `json:"header"`
		Body       map[string]interface{} `json:"body"`
		Query      map[string]interface{} `json:"query"`
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
