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

package engine

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/stretchr/testify/assert"
	"os"
)

func setModel(e Engine, ns, path string, t assert.TestingT) {
	buf, err := os.ReadFile(path)
	assert.Nil(t, err)
	err = e.AddNamespaceFromString(context.Background(), ns, string(buf))
	assert.Nil(t, err)
}

func insertOne(e Engine, ns, p string, tid primitive.ObjectID, data A, t assert.TestingT) A {
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(p)
	inserted, err := pTab.InsertOne(context.Background(), tid, data)
	assert.Nil(t, err)

	return inserted
}

func tryInsertOne(e Engine, ns, p string, tid primitive.ObjectID, data A, t assert.TestingT) (A, error) {
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(p)
	return pTab.InsertOne(context.Background(), tid, data)
}

func insertMany(e Engine, ns, p string, tids []primitive.ObjectID, data []A, t assert.TestingT) []A {
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(p)
	inserted, err := pTab.InsertMany(context.Background(), tids, data)
	assert.Nil(t, err)

	return inserted
}

func find(e Engine, ns, p string, filter interface{}, t assert.TestingT) []M {
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(p)
	updated, err := pTab.Find(context.Background(), filter)
	assert.Nil(t, err)
	return updated
}
