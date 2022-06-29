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

package index

type Setter func(old interface{}) (new interface{})

type Mutation interface {
	Set(key []byte, s Setter) error
	Delete(key []byte) error
	Commit() error // commit all keys
	Abort() error  // remove all keys
}

type mutation struct {
	uncommittedKeys [][]byte // used to check all uncommitted keys
	//TODO(weny): indexes tree
}
