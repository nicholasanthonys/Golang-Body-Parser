package test

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFind(t *testing.T) {
	letters := []string{"a", "b", "c", "d"}
	expected := true
	_, exist := util.Find(letters, "a")
	assert.Equal(t, expected, exist)
}

func TestGetListFolder(t *testing.T) {
	expected := []string{"configure0.json", "configure1.json", "configure2.json", "response.json"}
	files, err := util.GetListFolder("mock")
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