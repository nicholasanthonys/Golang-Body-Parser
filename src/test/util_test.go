package test

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFind(t *testing.T) {
	letters := []string{"a", "b", "c", "d"}
	expected := true
	_, exist := util.FindInSliceOfString(letters, "a")
	assert.Equal(t, expected, exist)
}

func TestGetListFolder(t *testing.T) {
	expected := []string{"base.json", "test-1_configure-0.json", "test-1_expected.json", "serial.json"}
	configureDir := os.Getenv("CONFIGURES_DIRECTORY_TESTING_NAME")
	fullProjectDir := configureDir + "/" + "test-1"
	files, err := util.GetListFolder(fullProjectDir)
	if err != nil {
		assert.Error(t, err, "Cannot get list folder")
	}
	for _, file := range files {
		_, exist := util.FindInSliceOfString(expected, file.Name())
		assert.Truef(t, exist, " file : %s ", file.Name())
	}
}
