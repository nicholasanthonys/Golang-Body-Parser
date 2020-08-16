package model

type Configure struct {
	Request Request `json:"request"`
}

type Request struct {
	Adds     StructAdd    `json:"adds"`
	Deletes  []string     `json:"deletes"`
	Modifies StructModify `json:"modifies"`
}

type StructAdd map[string]interface {
}

type StructModify map[string]interface{}
