module github.com/borschtapp/krip

go 1.26.0

retract [v1.0.0, v1.3.9] // broken semver, minor versions includes breaking changes

require (
	github.com/PuerkitoBio/goquery v1.11.0
	github.com/astappiev/microdata v1.0.2
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/mmcdole/gofeed v1.3.0
	github.com/sosodev/duration v1.4.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.51.0
)

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/astappiev/fixjson v1.0.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mmcdole/goxpp v1.1.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
