package model

type Configure struct {
	Methods  []string `json:"methods"`
	Request  Command  `json:"request"`
	Response Command  `json:"response"`
}

type Command struct {
	DestinationUrl string       `json:"destinationUrl"`
	MethodUsed     string       `json:"methodUsed"`
	Transform      string       `json:"transform"`
	Adds           Fields       `json:"adds"`
	Deletes        DeleteFields `json:"deletes"`
	Modifies       Fields       `json:"modifies"`
}

type Fields struct {
	Header map[string]interface{} `json:"header"`
	Body   map[string]interface{} `json:"body"`
	Query  map[string]interface{} `json:"query"`
}

type DeleteFields struct {
	Header []string `json:"header"`
	Body   []string `json:"body"`
	Query  []string `json:"query"`
}

type ParallelConfigure struct {
	ConfigureIndex int     `json:"configureIndex"`
	Response       Command `json:"response"`
}
