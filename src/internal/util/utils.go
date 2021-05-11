package util

import (
	"encoding/json"
	"github.com/clbanning/mxj/x2j"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

//* FindInSliceOfString is a function that will chek if  item exist in slice of string
func FindInSliceOfString(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func FindInSliceOfInt(slice []int, val int) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

//* FindRoute is a function that return index of a certain route given a path example : /serial/smsotp/generate
func FindRoute(routes []model.Route, path string, method string) *model.Route {
	for _, route := range routes {
		if path == route.Path && method == route.Method {
			return &route
		}
	}
	return nil
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

//*ErrorWriter is a function that will return Error Response
func ErrorWriter(cc *model.CustomContext, configure model.Configure, err error, status int) error {
	responseMap := make(map[string]interface{})
	responseMap["message"] = err.Error()
	switch configure.Response.Transform {
	case "ToXml":
		logrus.Warn(err.Error())
		xmlByte, _ := x2j.MapToXml(responseMap)
		return cc.XMLBlob(status, xmlByte)

	default:
		logrus.Warn(err.Error())
		return cc.JSON(status, responseMap)
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
	} else if strings.HasPrefix(value, "$header") {
		destination = "header"
	} else if strings.HasPrefix(value, "$query") {
		destination = "query"
	} else if strings.HasPrefix(value, "$response") {
		destination = "response"
	} else if strings.HasPrefix(value, "$path") {
		destination = "path"
	} else if strings.HasPrefix(value, "$status_code") {
		destination = "status_code"
	}

	sanitized = value[len(destination)+1:]

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

// DoLoggingJson will print logValue in a formatted log.
func DoLoggingJson(logValue map[string]interface{}, event string, identifier string, isRequest bool) {
	sentence := "logging "
	if isRequest {
		sentence += "response "
	} else {
		sentence += "response "
	}

	if event == "before" {
		sentence += "before modify for " + identifier + " : "
	} else {
		sentence += "after modify for " + identifier + " : "
	}

	jsonBytes, err := json.MarshalIndent(logValue, "", "\t")
	if err != nil {
		logrus.Error("error converting logValue to jsonbytes :")
		logrus.Error(err.Error())
		return
	}
	logrus.Info(string(jsonBytes))
}

func IsFileNameJson(filename string) bool {
	r := []rune(filename)

	// .json there are 5 character
	extension := string(r[len(filename)-5:])
	return extension == ".json"
}

func GetLogLevelFromEnv() logrus.Level {
	levelString := os.Getenv("LOG_LEVEL")
	if strings.ToLower(levelString) == "panic" {
		return logrus.PanicLevel
	}
	if strings.ToLower(levelString) == "fatal" {
		return logrus.FatalLevel
	}
	if strings.ToLower(levelString) == "error" {
		return logrus.ErrorLevel
	}
	if strings.ToLower(levelString) == "warn" {
		return logrus.WarnLevel
	}
	if strings.ToLower(levelString) == "debug" {
		return logrus.DebugLevel
	}

	if strings.ToLower(levelString) == "trace" {
		return logrus.TraceLevel
	}

	// default
	return logrus.InfoLevel

}
