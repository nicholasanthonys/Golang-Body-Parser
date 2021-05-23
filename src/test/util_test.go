package test

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
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
	expected := []string{"base.json", "configure-0.json", "configure-1.json", "serial.json"}
	configureDir := os.Getenv("CONFIGURES_DIRECTORY_TESTING_NAME")
	fullProjectDir := configureDir + "/" + "emailotp"
	files, err := util.GetListFolder(fullProjectDir)
	results := make([]string, 0)
	if err != nil {
		assert.Error(t, err, "Cannot get list folder")
	}
	for _, file := range files {
		logrus.Info("file name is")
		logrus.Info(file.Name())
		results = append(results, file.Name())
	}
	assert.Equal(t, expected, results)
}
