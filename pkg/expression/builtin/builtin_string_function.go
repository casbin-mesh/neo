// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builtin

import (
	"errors"
	"net"
	"path"
	"regexp"
	"strings"
)

var (
	keyMatch4Re *regexp.Regexp = regexp.MustCompile(`{([^/]+)}`)
)

// KeyGet returns the matched part
// For example, "/foo/bar/foo" matches "/foo/*"
// "bar/foo" will been returned
func KeyGet(key1, key2 string) string {
	i := strings.Index(key2, "*")
	if i == -1 {
		return ""
	}
	if len(key1) > i {
		if key1[:i] == key2[:i] {
			return key1[i:]
		}
	}
	return ""
}

var (
	keyGet2Re = regexp.MustCompile(`:[^/]+`)
)

// KeyGet2 returns value matched pattern
// For example, "/resource1" matches "/:resource"
// if the pathVar == "resource", then "resource1" will be returned
func KeyGet2(key1, key2, pathVar string) string {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	keys := keyGet2Re.FindAllString(key2, -1)
	key2 = keyGet2Re.ReplaceAllString(key2, "$1([^/]+)$2")
	key2 = "^" + key2 + "$"
	re2 := regexp.MustCompile(key2)
	values := re2.FindAllStringSubmatch(key1, -1)
	if len(values) == 0 {
		return ""
	}
	for i, key := range keys {
		if pathVar == key[1:] {
			return values[0][i+1]
		}
	}
	return ""
}

var (
	keyGet3Re = regexp.MustCompile(`\{[^/]+?\}`) // non-greedy match of `{...}` to support multiple {} in `/.../`
)

// KeyGet3 returns value matched pattern
// For example, "project/proj_project1_admin/" matches "project/proj_{project}_admin/"
// if the pathVar == "project", then "project1" will be returned
func KeyGet3(key1, key2 string, pathVar string) string {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	keys := keyGet3Re.FindAllString(key2, -1)
	key2 = keyGet3Re.ReplaceAllString(key2, "$1([^/]+?)$2")
	key2 = "^" + key2 + "$"
	re2 := regexp.MustCompile(key2)
	values := re2.FindAllStringSubmatch(key1, -1)
	if len(values) == 0 {
		return ""
	}
	for i, key := range keys {
		if pathVar == key[1:len(key)-1] {
			return values[0][i+1]
		}
	}
	return ""
}

// KeyMatch determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*"
func KeyMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	}
	return key1 == key2[:i]
}

var (
	keymatch2Re = regexp.MustCompile(`:[^/]+`)
)

// KeyMatch2 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*", "/resource1" matches "/:resource"
func KeyMatch2(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	key2 = keymatch2Re.ReplaceAllString(key2, "$1[^/]+$2")

	return RegexMatch(key1, "^"+key2+"$")
}

var (
	keymatch3Re = regexp.MustCompile(`\{[^/]+\}`)
)

// KeyMatch3 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*", "/resource1" matches "/{resource}"
func KeyMatch3(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	key2 = keymatch3Re.ReplaceAllString(key2, "$1[^/]+$2")

	return RegexMatch(key1, "^"+key2+"$")
}

var ()

// KeyMatch4 determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// Besides what KeyMatch3 does, KeyMatch4 can also match repeated patterns:
// "/parent/123/child/123" matches "/parent/{id}/child/{id}"
// "/parent/123/child/456" does not match "/parent/{id}/child/{id}"
// But KeyMatch3 will match both.
func KeyMatch4(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	tokens := []string{}

	re := keyMatch4Re
	key2 = re.ReplaceAllStringFunc(key2, func(s string) string {
		tokens = append(tokens, s[1:len(s)-1])
		return "([^/]+)"
	})

	re = regexp.MustCompile("^" + key2 + "$")
	matches := re.FindStringSubmatch(key1)
	if matches == nil {
		return false
	}
	matches = matches[1:]

	if len(tokens) != len(matches) {
		panic(errors.New("KeyMatch4: number of tokens is not equal to number of values"))
	}

	values := map[string]string{}

	for key, token := range tokens {
		if _, ok := values[token]; !ok {
			values[token] = matches[key]
		}
		if values[token] != matches[key] {
			return false
		}
	}

	return true
}

// KeyMatch5 determines whether key1 matches the pattern of key2 and ignores the parameters in key2.
// For example, "/foo/bar?status=1&type=2" matches "/foo/bar"
func KeyMatch5(key1 string, key2 string) bool {
	i := strings.Index(key1, "?")
	if i == -1 {
		return key1 == key2
	}

	return key1[:i] == key2
}

// RegexMatch determines whether key1 matches the pattern of key2 in regular expression.
func RegexMatch(key1 string, patten string) bool {
	res, err := regexp.MatchString(patten, key1)
	if err != nil {
		panic(err)
	}
	return res
}

// IPMatch determines whether IP address ip1 matches the pattern of IP address ip2, ip2 can be an IP address or a CIDR pattern.
// For example, "192.168.2.123" matches "192.168.2.0/24"
func IPMatch(ip1 string, ip2 string) bool {
	objIP1 := net.ParseIP(ip1)
	if objIP1 == nil {
		panic("invalid argument: ip1 in IPMatch() function is not an IP address.")
	}

	_, cidr, err := net.ParseCIDR(ip2)
	if err != nil {
		objIP2 := net.ParseIP(ip2)
		if objIP2 == nil {
			panic("invalid argument: ip2 in IPMatch() function is neither an IP address nor a CIDR.")
		}

		return objIP1.Equal(objIP2)
	}

	return cidr.Contains(objIP1)
}

// GlobMatch determines whether key1 matches the pattern of key2 using glob pattern
func GlobMatch(key1 string, key2 string) (bool, error) {
	return path.Match(key2, key1)
}
