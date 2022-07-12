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

package utils

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const basicModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

func TestConfig_Parse(t *testing.T) {
	c, err := NewParse(bufio.NewReader(strings.NewReader(basicModel)))
	assert.Nil(t, err)
	rd := c.RequestDef()["r"]
	assert.Equal(t, "sub, obj, act", rd)
	roleD := c.RoleDef()["g"]
	assert.Equal(t, "_, _", roleD)
	pd := c.PolicyDef()["p"]
	assert.Equal(t, "sub, obj, act", pd)
	pe := c.PolicyEffect()["e"]
	assert.Equal(t, "some(where (p.eft == allow))", pe)
	m := c.Matchers()["m"]
	assert.Equal(t, "r.sub == p.sub && r.obj == p.obj && r.act == p.act", m)
}
