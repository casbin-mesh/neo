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

package optimizer

import (
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"testing"
)

func TestPredicateAccessorMember(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		root := parser.MustParseFromString("g(r.sub,p.sub) && r.obj == p.obj && r.act == p.act")
		pred := Optimize(root)
		result := GetPredicateAccessorMembers(pred, IncludeAccessorOnly)
		slices.Sort(result)
		assert.Equal(t, []string{"act", "obj"}, result)
	})
}
