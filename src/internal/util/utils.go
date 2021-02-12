package util

import (
	"encoding/json"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/clbanning/mxj/x2j"
	"github.com/labstack/echo"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
)

//* Find is a function that will chek if  item exist in slice of string
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

//* FindRouteIndex is a function that return index of a certain route given a path example : /serial/smsotp/generate
func FindRouteIndex(routes []model.Route, path string) int {
	for index, route := range routes {
		if strings.Contains(path, route.Path) {
			return index
		}
	}
	return -1
}

func GetListFolder(dirname string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	logrus.Info("reading directory" + dirname + " : ")
	for _, file := range files {
		logrus.Info(file.Name())
	}
	return files, nil

}

//* ReadJsonFile is a function that will read configure.json File.
func ReadJsonFile(path string) []byte {
	// Open our jsonFile
	jsonFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		logrus.Error(err.Error())
		logrus.Error("Failed to read : " + path)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadFile(path)
	return byteValue

}

//*ResponseWriter is a function that will return response
func ResponseWriter(wrapper model.Wrapper, c echo.Context) error {

	switch strings.ToLower(wrapper.Configure.Response.Transform) {
	case strings.ToLower("ToJson"):
		return c.JSON(200, wrapper.Response.Body)
	case strings.ToLower("ToXml"):
		resByte, _ := x2j.MapToXml(wrapper.Response.Body)
		return c.XMLBlob(200, resByte)
	default:
		logrus.Info("type not supported. only support ToJson and ToXml")
		return c.JSON(404, "Type Not Supported. only support ToJson and ToXml")
	}
}

//*ErrorWriter is a function that will return Error Response
func ErrorWriter(c echo.Context, configure model.Configure, err error, status int) error {
	responseMap := make(map[string]interface{})
	responseMap["message"] = err.Error()
	switch configure.Response.Transform {
	case "ToXml":
		logrus.Warn(err.Error())
		xmlByte, _ := x2j.MapToXml(responseMap)
		return c.XMLBlob(status, xmlByte)

	default:
		logrus.Warn(err.Error())
		return c.JSON(status, responseMap)
	}
}

//* RemoveSquareBracketAndConvertToSlice is a function that will remove the square bracket
//* and convert values into slice
func RemoveSquareBracketAndConvertToSlice(value string, separator string) []string {
	listTraverse := make([]string, 0)
	temp := ""
	arraySplit := strings.Split(value, separator)
	for _, val := range arraySplit {
		if val != "[" {
			if val == "]" {
				//*push
				listTraverse = append(listTraverse, temp)
				temp = ""
			} else {
				//* add character to temp
				temp += val
			}
		}

	}
	return listTraverse
}

// SanitizeValue is a function that will remove the dollar sign from the value in configure.json.
// we remove the dollar sign example :  $body[user], and only pick the rest of the word ex : body[user] */
func SanitizeValue(value string) ([]string, string) {
	var destination string
	var sanitized string
	if strings.HasPrefix(value, "$body") {
		destination = "body"
		sanitized = value[5:]
	} else if strings.HasPrefix(value, "$header") {
		destination = "header"
		sanitized = value[7:]
	} else if strings.HasPrefix(value, "$query") {
		destination = "query"
		sanitized = value[6:]
	} else if strings.HasPrefix(value, "$response") {
		destination = "response"
		sanitized = value[9:]
	} else if strings.HasPrefix(value, "$path") {
		destination = "path"
		sanitized = value[5:]
	} else {
		return nil, value
	}

	//* We call this function to remove square bracket. ex : $body[user] will become :  body user (as a slice)
	return RemoveSquareBracketAndConvertToSlice(sanitized, ""), destination
}

func RemoveCharacters(input string, characters string) string {
	filter := func(r rune) rune {
		if strings.IndexRune(characters, r) < 0 {
			return r
		}
		return -1
	}

	return strings.Map(filter, input)

}

// JSONEqual compares the JSON from two Readers.
func JSONEqual(a, b io.Reader) (bool, error) {
	var j, j2 interface{}
	d := json.NewDecoder(a)
	if err := d.Decode(&j); err != nil {
		return false, err
	}
	d = json.NewDecoder(b)
	if err := d.Decode(&j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

// JSONBytesEqual compares the JSON in two byte slices.
func JSONBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		logrus.Error("Error unmarshaling a")
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		logrus.Error("Error unmarshaling b ")
	}

	return reflect.DeepEqual(j2, j), nil
}

func DoLogging(logValue string, field model.Fields, event string, fileName string, isRequest bool) {
	if len(logValue) > 0 {
		sentence := "logging "
		if isRequest {
			sentence += "response "
		} else {
			sentence += "response "
		}

		if event == "before" {
			sentence += "before modify for " + fileName + " : "
		} else {
			sentence += "after modify for " + fileName + " : "
		}

		value := service.CheckValue(logValue, field)
		logrus.Info(sentence)
		logrus.Info(value)
	}
}
