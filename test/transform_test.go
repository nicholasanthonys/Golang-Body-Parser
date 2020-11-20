package test

import (
	"errors"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

//* Test to load function from plugin
func TestLoadFunction(t *testing.T) {
	transformFunction := service.LoadFunctionFromModule("../plugin/transform.so", "ToJson")
	if transformFunction == nil {
		assert.Error(t, errors.New("Transform function is nil"), "error opening transform function")
	}

}
