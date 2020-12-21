package model

type Route struct {
	Path             string `json:"path"`
	ProjectDirectory string `json:"project_directory"`
	Type             string `json:"type"`
	Method           string `json:"method"`
}

type Routes []Route

type Configure struct {
	ConfigureBased string  `json:"configure_based"`
	Request        Command `json:"request"`
	Response       Command `json:"response"`
}

type Command struct {
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

type Wrapper struct {
	Configure Configure
	Request   Fields
	Response  Fields
}

type Fields struct {
	Param  map[string]interface{}
	Header map[string]interface{} `json:"header"`
	Body   map[string]interface{} `json:"body"`
	Query  map[string]interface{} `json:"query"`
}

type DeleteFields struct {
	Header []string `json:"header"`
	Body   []string `json:"body"`
	Query  []string `json:"query"`
}

// errorString is a trivial implementation of error.
type ErrorString struct {
	s string
}
