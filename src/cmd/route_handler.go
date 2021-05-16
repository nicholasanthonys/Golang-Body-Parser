package main

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/request"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var configureDir string
var routes []model.Route

//SetRouteHandler called by main.go. This function set route based on router.json
func SetRouteHandler() *echo.Echo {
	//* get configures Directory
	configureDir = os.Getenv("CONFIGURES_DIRECTORY_NAME")

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &model.CustomContext{Context: c}
			return next(cc)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// * Read router.json
	routesByte := util.ReadJsonFile(configureDir + "/router.json")
	err := json.Unmarshal(routesByte, &routes)
	if err != nil {
		logrus.Error(err.Error())
	} else {
		//*add index route
		e.GET("/", func(c echo.Context) error {
			cc := c.(*model.CustomContext)
			return cc.String(http.StatusOK, "Golang-Body-Parser Active")
		})

		//*set path based from configure
		for _, route := range routes {

			if strings.ToLower(route.Method) == "post" {
				if strings.ToLower(route.Type) == "parallel" {
					e.POST(route.Path, parallelRouteHandler, prepareParallelRoute)
				} else {
					e.POST(route.Path,
						serialRouteHandler, prepareSerialRoute)
				}
			}

			if strings.ToLower(route.Method) == "get" {
				if strings.ToLower(route.Type) == "parallel" {
					e.GET(route.Path, parallelRouteHandler, prepareParallelRoute)
				} else {
					e.GET(route.Path, serialRouteHandler, prepareSerialRoute)
				}
			}

			if strings.ToLower(route.Method) == "put" {
				if strings.ToLower(route.Type) == "parallel" {
					e.PUT(route.Path, parallelRouteHandler, prepareParallelRoute)
				} else {
					e.PUT(route.Path, serialRouteHandler, prepareSerialRoute)
				}

			}

			if strings.ToLower(route.Method) == "delete" {
				if strings.ToLower(route.Type) == "parallel" {
					e.DELETE(route.Path, parallelRouteHandler, prepareParallelRoute)
				} else {
					e.DELETE(route.Path, serialRouteHandler, prepareSerialRoute)
				}
			}

		}
	}

	return e
}

// prepareSerialRoute middleware that find defined route in route.json and read SerialProject.json
func prepareSerialRoute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		route := util.FindRoute(routes, c.Path(), c.Request().Method)
		if route == nil {
			return c.JSON(404, "Cannot FindInSliceOfString DefinedRoute "+c.Path())
		}

		cc := c.(*model.CustomContext)
		cc.DefinedRoute = route
		cc.FullProjectDirectory = configureDir + "/" + route.ProjectDirectory

		log.Info("full SerialProject directory is")
		log.Info(cc.FullProjectDirectory)

		return next(cc)
	}
}

// prepareSerialRoute middleware that find defined route in route.json and read SerialProject.json
func prepareParallelRoute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		route := util.FindRoute(routes, c.Path(), c.Request().Method)
		cc := &model.CustomContext{Context: c}
		cc.DefinedRoute = route
		cc.FullProjectDirectory = configureDir + "/" + route.ProjectDirectory

		logrus.Info("full SerialProject directory is")
		logrus.Info(cc.FullProjectDirectory)

		return next(cc)
	}
}

// parallelRouteHandler execute every configure in parallel-way.
func parallelRouteHandler(c echo.Context) error {
	cc := c.(*model.CustomContext)
	baseProject, err := readBaseFile(cc.FullProjectDirectory)

	if err != nil {
		response := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(500, response)
	}
	cc.BaseProject = baseProject
	mapWrapper := cmap.New()
	cc.MapWrapper = &mapWrapper
	return request.DoParallel(cc, 0)
}

// serialRouteHandler process configure in serial-way.
func serialRouteHandler(c echo.Context) error {
	cc := c.(*model.CustomContext)
	baseProject, err := readBaseFile(cc.FullProjectDirectory)
	if err != nil {
		response := map[string]interface{}{
			"message": err.Error(),
		}
		return cc.JSON(500, response)
	}
	cc.BaseProject = baseProject
	mapWrapper := cmap.New()
	cc.MapWrapper = &mapWrapper
	return request.DoSerial(cc, 0)

}

func readBaseFile(fullProjectDirectory string) (model.Base, error) {
	var baseProject model.Base

	// Read base.json
	baseByte := util.ReadJsonFile(fullProjectDirectory + "/base.json")
	err := json.Unmarshal(baseByte, &baseProject)
	if err != nil {

		log.Error(err.Error())
		return baseProject, err
	}

	// load env max circular
	envMaxCircularString := os.Getenv("MAX_CIRCULAR")
	envMaxCircular, err := strconv.Atoi(envMaxCircularString)
	if err != nil {
		return baseProject, err
	}

	if baseProject.MaxCircular > envMaxCircular {
		log.Warn("project.MaxCircular > envMaxCircular")
		log.Warn("set baseProject.MaxCircular = ", envMaxCircular)
		baseProject.MaxCircular = envMaxCircular
	}
	return baseProject, err
}
