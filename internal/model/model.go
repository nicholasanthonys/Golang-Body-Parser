package model

type Configure struct {
	ConfigureBased string   `json:"configureBased"`
	Methods        []string `json:"methods"`
	Path           string   `json:"path"`
	Request        Command  `json:"request"`
	Response       Command  `json:"response"`
}

type Command struct {
	DestinationUrl string       `json:"destinationUrl"`
	MethodUsed     string       `json:"methodUsed"`
	Transform      string       `json:"transform"`
	Adds           Fields       `json:"adds"`
	Deletes        DeleteFields `json:"deletes"`
	Modifies       Fields       `json:"modifies"`
}

type Wrapper struct {
	Request  Fields
	Response Fields
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
