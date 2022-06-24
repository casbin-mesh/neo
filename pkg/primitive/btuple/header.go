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

package btuple

import "encoding/binary"

const SizeOfHeader = 8

// header
// tuple type | len | offset(only for large tuple)
type header struct {
	typ BTupleType // 1B
	len uint32     // 4B address 2^32-1 elements
}

func NewHeader(typ BTupleType, len uint32) *header {
	return &header{typ: typ, len: len}
}

// writeTo return written size.
// padding to 8 Bytes
func (h *header) writeTo(dst []byte) int {
	dst[0] = byte(h.typ)
	// BigEndian
	binary.BigEndian.PutUint32(dst[1:], h.len)
	return SizeOfHeader
}

func (h *header) decode(src []byte) {
	h.typ = BTupleType(src[0])
	// BigEndian
	h.len = binary.BigEndian.Uint32(src[1:])
}
