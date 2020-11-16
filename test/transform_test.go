package test

import (
	"plugin"
	"testing"
)

func TestLoadFunction(t *testing.T) {
	_, err := plugin.Open("../plugin/transform.so")
	if err != nil {
		t.Error(t, err.Error())

	}
}
