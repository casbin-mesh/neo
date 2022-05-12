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

package art

const MaxPrefixLen = 10

// Key Type.
type Key []byte

func (key Key) Clone() Key {
	cloned := make(Key, len(key))
	copy(cloned, key)
	return cloned
}

// Value type.
type Value []byte

func (value Value) Clone() Key {
	cloned := make(Key, len(value))
	copy(cloned, value)
	return cloned
}
