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

package neo

import (
	badgerAdapter "github.com/casbin-mesh/neo/pkg/db/adapter/badger"
	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	engine *neo
)

const DBPATH = "./__db_test__"

func TestMain(m *testing.M) {
	var db, err = badgerAdapter.OpenManaged(badger.DefaultOptions(DBPATH))
	if err != nil {
		panic(err)
	}
	engine = New(Options{db: db})
	code := m.Run()

	// clean up
	err = db.Close()
	if err != nil {
		panic(err)
	}
	err = os.RemoveAll(DBPATH)
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestNeo_Mutation(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		m, err := engine.NewMutationAt(0, []byte("test"))
		defer m.CommitAt(1)
		assert.Nil(t, err)
	})

}
