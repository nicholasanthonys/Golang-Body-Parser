package model

type Configure struct {
	DestinationUrl string  `json:"destinationurl"`
	Method         string  `json:"method"`
	Request        Command `json:"request"`
	Response       Command `json:"response"`
}

type Command struct {
	Transform string       `json:"transform"`
	Adds      StructAdd    `json:"adds"`
	Deletes   []string     `json:"deletes"`
	Modifies  StructModify `json:"modifies"`
}

type StructAdd map[string]interface {
}

type StructModify map[string]interface{}
