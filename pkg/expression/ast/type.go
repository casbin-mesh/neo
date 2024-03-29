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

package ast

type Type uint

const (
	BOOLEAN Type = iota + 1
	INT
	FLOAT
	STRING
	TUPLE
	ERROR
	MEMBER_ACCESSOR
	IDENTIFIER
	NULL
)

var (
	typeToString = []string{
		"BOOLEAN",
		"INT",
		"FLOAT",
		"STRING",
		"TUPLE",
		"ERROR",
		"MEMBER_ACCESSOR",
		"IDENTIFIER",
		"NULL",
	}
)

func (t Type) String() string {
	return typeToString[t]
}
