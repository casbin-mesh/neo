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
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
)

func setupEngine(t assert.TestingT, path string) Engine {
	e, err := NewEngineFromPath(path)
	assert.Nil(t, err)
	return e
}

func TestEngine_AddNamespaceFromString(t *testing.T) {
	p := "./__test_tmp__/add_namespace_from_string"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	setModel(e, "test_namespace", "../../../examples/assets/model/basic_model.conf", t)
}

func TestEngine_InsertOne(t *testing.T) {
	p := "./__test_tmp__/namespace_insert_one"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	// setup basic model
	setModel(e, "test_namespace", "../../../examples/assets/model/basic_model.conf", t)
	inserted := insertOne(e, "test_namespace", "p", A{"alice", "data1", "read"}, t)
	assert.Equal(t, []interface{}{"alice", "data1", "read", "allow"}, inserted)
}

func TestEngine_InsertMany(t *testing.T) {
	p := "./__test_tmp__/namespace_insert_many"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	// setup basic model
	setModel(e, "test_namespace", "../../../examples/assets/model/basic_model.conf", t)
	inserted := insertMany(e, "test_namespace", "p", []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
	assert.Equal(t, [][]interface{}{{"alice", "data1", "read", "allow"}, {"bob", "data2", "write", "allow"}}, inserted)
	fmt.Printf("%v\n", inserted)
}

func TestEngine_DeleteOne(t *testing.T) {
	p := "./__test_tmp__/namespace_delete_one"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	ns := "test_namespace"
	po := "p"
	// setup basic model
	setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
	_ = insertMany(e, ns, po, []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
	deleted := deleteOne(e, ns, po, A{"alice", "data1", "read", "allow"}, t)
	assert.Equal(t, []interface{}{"alice", "data1", "read", "allow"}, deleted)
}

func TestEngine_UpdateOne(t *testing.T) {
	p := "./__test_tmp__/namespace_update_one"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	ns := "test_namespace"
	po := "p"
	// setup basic model
	setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
	_ = insertMany(e, ns, po, []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
	deleted := updateOne(e, ns, po, A{"alice", "data1", "read", "allow"}, A{"alice", "data1", "write", "allow"}, t)
	// returns the old one
	assert.Equal(t, []interface{}{"alice", "data1", "write", "allow"}, deleted)
}

func TestEngine_EnforceOne(t *testing.T) {
	p := "./__test_tmp__/namespace_enforce_one"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	ns := "test_namespace"
	po := "p"
	// setup basic model
	setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
	_ = insertMany(e, ns, po, []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
}

func TestEngine_Find(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		p := "./__test_tmp__/namespace_find_one_basic"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
		_ = insertMany(e, ns, po, []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
		found := find(e, ns, po, nil, t)
		assert.Equal(t, "[map[act:read eft:allow obj:data1 sub:alice] map[act:write eft:allow obj:data2 sub:bob]]", fmt.Sprintf("%v", found))
	})
	t.Run("with filter", func(t *testing.T) {
		p := "./__test_tmp__/namespace_find_one_with_filter"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
		_ = insertMany(e, ns, po, []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
		found := find(e, ns, po, map[string]interface{}{"sub": "alice"}, t)
		assert.Equal(t, "[map[act:read eft:allow obj:data1 sub:alice]]", fmt.Sprintf("%v", found))
		fmt.Println(found)
	})
}

func TestDecode(t *testing.T) {
	var data []interface{}
	data = append(data, 1)
	data = append(data, "string")
	m := bson.M{"data": data}
	b, err := bson.Marshal(m)
	assert.Nil(t, err)
	fmt.Printf("%v", b)
}
