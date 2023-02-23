package testdata

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
	"github.com/borschtapp/krip/utils"
)

var HtmlExt = ".html"
var JsonExt = ".json"
var NewExt = ".new"

var PackageDir = currentPath() + "/"
var WebsitesDir = PackageDir + "websites/"
var RecipesDir = PackageDir + "recipes/"

func currentPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

func WalkTestdataWebsites(fn func(name string, path string)) {
	_ = filepath.Walk(WebsitesDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), HtmlExt) {
			fn(info.Name(), path)
		}
		return nil
	})
}

func WalkTestdataRecipes(fn func(name string, recipe model.Recipe)) {
	_ = filepath.Walk(RecipesDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), JsonExt) {
			recipe := model.Recipe{}

			file, _ := os.ReadFile(path)
			_ = json.Unmarshal(file, &recipe)

			fn(info.Name(), recipe)
		}
		return nil
	})
}

func AssertRecipe(t *testing.T, recipe *model.Recipe) {
	AssertJson(t, recipe, RecipesDir+utils.HostAlias(recipe.Url))
}

func AssertJson(t *testing.T, data any, filePath string) {
	jsonObj, err := json.MarshalIndent(data, "", "  ")
	assert.NoError(t, err)

	fileName := filePath + JsonExt
	expectedObj, err := os.ReadFile(fileName)
	assert.NoError(t, err)

	if !assert.JSONEq(t, string(expectedObj), string(jsonObj)) {
		if _, ok := os.LookupEnv("RECIPE_OVERRIDE"); !ok {
			fileName += NewExt
		}

		assert.NoError(t, os.WriteFile(fileName, jsonObj, 0644))
	}
}

func OptionallyMockRequests(t *testing.T) {
	if _, ok := os.LookupEnv("RECIPE_ONLINE"); !ok {
		MockRequests(t)
	}
}

func MockRequests(t *testing.T) {
	httpmock.Activate()

	httpmock.RegisterNoResponder(func(req *http.Request) (*http.Response, error) {
		data, err := mockResponse(req.URL.String(), req.Header.Get("Accept"))

		if err != nil {
			// if no file found, try to remove subdomain (required for gousto)
			req.URL.Host = req.URL.Host[strings.Index(req.URL.Host, ".")+1:]
			data, err = mockResponse(req.URL.String(), req.Header.Get("Accept"))
		}

		if err != nil {
			return httpmock.NewStringResponse(http.StatusInternalServerError, "HttpMock: "+err.Error()), nil
		} else {
			response := httpmock.NewBytesResponse(http.StatusOK, data)
			response.Request = req
			if req.Header.Get("Accept") == "application/json" {
				response.Header.Set("Content-Type", "application/json")
			} else {
				response.Header.Set("Content-Type", "text/html; charset=utf-8")
			}
			return response, nil
		}
	})

	t.Cleanup(httpmock.Deactivate)
}

func mockResponse(requestUrl string, accept string) ([]byte, error) {
	fileName := WebsitesDir + utils.HostAlias(requestUrl)
	switch accept {
	case "", "text/html":
		fileName += HtmlExt
	case "application/json":
		fileName += JsonExt
	default:
		return nil, errors.New("unknown accept type")
	}

	return os.ReadFile(fileName)
}
