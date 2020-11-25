package test

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

//* Test to load function from plugin
func TestLoadFunction(t *testing.T) {
	_, err := service.LoadFunctionFromModule("../plugin/transform.so", "ToJson")
	if err != nil {
		assert.Error(t, err, "unable to open function module")
	}
}
