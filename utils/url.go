package utils

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

func IsAbsolute(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Malformed url:", err)
		return false
	}

	if u.IsAbs() {
		return true
	}

	return false
}

func ToAbsoluteUrl(base *url.URL, urlStr string) string {
	if len(urlStr) == 0 || base == nil {
		return ""
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Error parsing url:", err)
		return ""
	}

	if u.IsAbs() {
		return urlStr
	}

	return base.ResolveReference(u).String()
}

func RemoveTrailingSlash(urlStr string) string {
	return strings.ToLower(strings.TrimSuffix(urlStr, "/"))
}

func BaseUrl(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func Hostname(urlStr string) string {
	if !strings.Contains(urlStr, "/") {
		return urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host
}

func DomainZone(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	hostParts := strings.Split(strings.ToLower(u.Hostname()), ".")
	return hostParts[len(hostParts)-1]
}

func HostAlias(urlStr string) string {
	alias := Hostname(urlStr)

	// remove public domain
	suffix, _ := publicsuffix.PublicSuffix(alias)
	alias = strings.TrimSuffix(alias, "."+suffix)

	// remove common prefixes
	alias = strings.TrimPrefix(alias, "api.")

	// replace dots with underscores
	alias = strings.ReplaceAll(alias, ".", "_")
	return alias
}
