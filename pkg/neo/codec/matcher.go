// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
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

package codec

import (
	"github.com/casbin-mesh/neo/fb"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	flatbuffers "github.com/google/flatbuffers/go"
)

// MatcherInfoKey s_m{id}
func MatcherInfoKey(matcherId uint64) []byte {
	buf := make([]byte, 0, 11)
	buf = append(buf, mSchemaPrefix...)
	buf = append(buf, matcherPrefix...)
	buf = appendUint64(buf, matcherId)
	return buf
}

func EncodeMatcherInfo(info *model.MatcherInfo) []byte {
	builder := flatbuffers.NewBuilder(1024)
	LName := builder.CreateString(info.Name.L)
	OName := builder.CreateString(info.Name.O)
	raw := builder.CreateString(info.Raw)

	// name
	fb.CIStrStart(builder)
	fb.CIStrAddL(builder, LName)
	fb.CIStrAddO(builder, OName)
	name := fb.CIStrEnd(builder)

	fb.MatcherInfoStart(builder)
	fb.MatcherInfoAddId(builder, info.ID)
	fb.MatcherInfoAddName(builder, name)
	fb.MatcherInfoAddRaw(builder, raw)
	orc := fb.MatcherInfoEnd(builder)
	builder.Finish(orc)

	return builder.FinishedBytes()
}
