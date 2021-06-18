package test

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFind(t *testing.T) {
	letters := []string{"a", "b", "c", "d"}
	expected := true
	_, exist := util.FindInSliceOfString(letters, "a")
	assert.Equal(t, expected, exist)
}

func Test_Get_List_File_Inside_Folder(t *testing.T) {
	expected := []string{"base.json", "test-1_configure-0.json",
		"test-1.1_expected.json", "serial.json"}
	fullProjectDir := dirName + "/" + "test-1.1"
	files, err := util.GetListFolder(fullProjectDir)
	if err != nil {
		assert.Error(t, err, "Cannot get list folder")
	}
	for _, file := range files {

		_, exist := util.FindInSliceOfString(expected, file.Name())
		assert.Truef(t, exist, " file : %s ", file.Name())
	}
}

func Test_Get_List_Folder(t *testing.T) {
	expected := []string{
		"emailotp", "test-1.1", "test-1.2", "test-2.2", "test-2.4", "test-3.2",
		"test-3.4", "test-4.2",
		"test-4.4", "test-5.2", "test-6.1", "test-6.3", "test-7.1", "test-7.3", "test-8.1_parallel_to_serial",
		"imdb", "smsotp", "test-2.1", "test-2.3", "test-3.1", "test-3.3", "test-4.1", "test-4.3",
		"test-5.1", "test-5.3", "test-6.2", "test-6.4", "test-7.2",
		"test-7.4", "test-8.2_serial_to_parallel", "router.json",
	}
	files, err := util.GetListFolder(dirName)
	if err != nil {
		assert.Error(t, err, "Cannot get list folder")
	}
	for _, file := range files {
		_, exist := util.FindInSliceOfString(expected, file.Name())
		assert.Truef(t, exist, " file : %s ", file.Name())
	}
}
