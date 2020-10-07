package service

import (
	"github.com/clbanning/mxj/x2j"
	"github.com/labstack/echo"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"strconv"
)

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func ResponseWriter(configure model.Configure, resultMap []map[string]interface{}, arrByte [][]byte, c echo.Context) error {
	switch configure.Response.Transform {
	case "ToJson":
		return c.JSON(200, resultMap)
	case "ToXml":
		newResMap := make(map[string]interface{})
		for i, res := range resultMap {
			byteRes, _ := x2j.MapToXml(res)
			index := strconv.Itoa(i)
			newResMap["response"+index] = byteRes
		}
		resByte, _ := x2j.MapToXml(newResMap)
		return c.XMLBlob(200, resByte)
	default:
		return c.JSON(404, "Type Not Supported")
	}
}
