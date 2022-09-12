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

package cascades

type Memo struct {
	Groups      []Group
	root        IndexType
	lookupTable map[string]struct{}
}

func NewMemo() *Memo {
	return &Memo{Groups: nil, root: -1}
}

func (m *Memo) Root() *Group {
	if m.root != -1 {
		return &m.Groups[m.root]
	}
	return nil
}

func (m *Memo) Init(expr SExpr) error {
	m.SetRoot(m.Insert(NoneIndex, expr))
	return nil
}

func (m *Memo) SetRoot(idx IndexType) {
	m.root = idx
}

func (m *Memo) Insert(targetGroup IndexType, expr SExpr) IndexType {
	children := make([]IndexType, 0, len(expr.Children()))
	for _, sExpr := range expr.Children() {
		children = append(children, m.Insert(NoneIndex, sExpr.Clone()))
	}

	if expr.OriginalGroup != NoneIndex {
		// The expression is extracted by PatternExtractor, no need to reinsert.
		return expr.OriginalGroup
	}

	if targetGroup == NoneIndex {
		targetGroup = m.AddGroup()
	}

	mExpr := NewMExpr(targetGroup, IndexType(m.Group(targetGroup).Len()), expr.Plan().Clone(), children)
	m.InsertMExpr(targetGroup, mExpr)
	return targetGroup
}

func (m *Memo) InsertMExpr(index IndexType, expr *MExpr) {
	key := string(expr.FingerPrint())
	_, ok := m.lookupTable[key]
	if !ok {
		m.lookupTable[key] = struct{}{}
	}
}

func (m *Memo) Group(idx IndexType) *Group {
	return &m.Groups[idx]
}

func (m *Memo) AddGroup() IndexType {
	group := Group{
		GroupIndex: IndexType(len(m.Groups)),
		MExprs:     nil,
	}
	m.Groups = append(m.Groups, group)
	return group.GroupIndex
}
