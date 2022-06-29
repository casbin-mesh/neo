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

package bschema

type BSchema interface {
	Key() []byte
	// ValueCopy returns a copy of the value of the item from the value bytes, writing it to dst slice.
	// If nil is passed, or capacity of dst isn't sufficient, a new slice would be allocated and
	// returned.
	ValueCopy(dst []byte) []byte
}
