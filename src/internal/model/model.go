package model

type Configure struct {
	ConfigureBased string   `json:"configureBased"`
	Methods        []string `json:"methods"`
	Path           string   `json:"path"`
	Request        Command  `json:"request"`
	Response       Command  `json:"response"`
}

type Command struct {
	DestinationPath string       `json:"destinationPath"`
	DestinationUrl  string       `json:"destinationUrl"`
	MethodUsed      string       `json:"methodUsed"`
	Transform       string       `json:"transform"`
	LogBeforeModify string       `json:"logBeforeModify"`
	LogAfterModify  string       `json:"logAfterModify"`
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
