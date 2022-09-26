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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func setModel(e Engine, ns, path string, t *testing.T) {
	buf, err := os.ReadFile(path)
	assert.Nil(t, err)
	err = e.AddNamespaceFromString(context.Background(), ns, string(buf))
	assert.Nil(t, err)
}

func insertOne(e Engine, ns, p string, data A, t *testing.T) A {
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(p)
	inserted, err := pTab.InsertOne(context.Background(), data)
	assert.Nil(t, err)

	return inserted
}

func insertMany(e Engine, ns, p string, data []A, t *testing.T) []A {
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(p)
	inserted, err := pTab.InsertMany(context.Background(), data)
	assert.Nil(t, err)

	return inserted
}

func deleteOne(e Engine, ns, p string, data A, t *testing.T) A {
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(p)
	deleted, err := pTab.DeleteOne(context.Background(), data)
	assert.Nil(t, err)
	return deleted
}
