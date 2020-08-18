package model

type Configure struct {
	DestinationUrl string         `json:"destinationurl"`
	Method         string         `json:"method"`
	Request        CommandRequest `json:"request"`
}

type CommandRequest struct {
	Adds     StructAdd    `json:"adds"`
	Deletes  []string     `json:"deletes"`
	Modifies StructModify `json:"modifies"`
}

type StructAdd map[string]interface {
}

type StructModify map[string]interface{}
