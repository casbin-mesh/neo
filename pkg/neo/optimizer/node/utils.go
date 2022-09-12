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

package node

import (
	"fmt"
	"strings"
)

func childrenIdent(children string) string {
	lines := strings.Split(children, "\n")
	for i, line := range lines {
		if i != 0 && line != "" {
			lines[i] = "  " + line
		}
	}
	return strings.Join(lines, "\n")
}

func childrenIdentWithLine(children string) string {
	lines := strings.Split(children, "\n")
	for i, line := range lines {
		if i != 0 && line != "" {
			lines[i] = "│ " + line
		}
	}
	return strings.Join(lines, "\n")
}

func treeFormat(parent string, children ...string) string {
	result := parent + "\n"
	for i, child := range children {
		if i != len(children)-1 {
			result += childrenIdentWithLine(fmt.Sprintf("├─%s\n", child))
		} else {
			result += childrenIdent(fmt.Sprintf("└─%s\n", child))
		}
	}
	for strings.HasSuffix(result, "\n") {
		result = result[:len(result)-1]
	}
	return result
}
