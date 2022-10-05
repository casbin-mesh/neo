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
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/executor"
	"github.com/casbin-mesh/neo/pkg/primitive"
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

func NewObjectIds(n int) []primitive.ObjectID {
	tid := make([]primitive.ObjectID, 0, n)
	for i := 0; i < n; i++ {
		tid = append(tid, primitive.NewObjectID())
	}
	return tid
}

func TestEngine_InsertOne(t *testing.T) {
	t.Run("should insert one", func(t *testing.T) {
		p := "./__test_tmp__/namespace_insert_one"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		// setup basic model
		setModel(e, "test_namespace", "../../../examples/assets/model/basic_model.conf", t)
		inserted := insertOne(e, "test_namespace", "p", primitive.NewObjectID(), A{"alice", "data1", "read"}, t)
		assert.Equal(t, []interface{}{"alice", "data1", "read", "allow"}, inserted)
	})
	t.Run("should failed to insert duplicate record", func(t *testing.T) {
		p := "./__test_tmp__/namespace_insert_one_duplicates"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		policyTableName := "p"
		// setup basic model
		setModel(e, "test_namespace", "../../../examples/assets/model/basic_model.conf", t)
		oid := primitive.NewObjectID()
		inserted := insertOne(e, ns, policyTableName, oid, A{"alice", "data1", "read"}, t)
		assert.Equal(t, []interface{}{"alice", "data1", "read", "allow"}, inserted)

		inserted, err := tryInsertOne(e, ns, policyTableName, oid, A{"alice", "data1", "read"}, t)
		assert.Equal(t, executor.ErrPrimaryKeyDuplicates, err)
	})
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
	inserted := insertMany(e, "test_namespace", "p", NewObjectIds(2), []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
	assert.Equal(t, [][]interface{}{{"alice", "data1", "read", "allow"}, {"bob", "data2", "write", "allow"}}, inserted)
	fmt.Printf("%v\n", inserted)
}

func TestEngine_Update(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		p := "./__test_tmp__/namespace_update_basic"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
		_ = insertMany(e, ns, po, NewObjectIds(2), []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)

		filter := M{"sub": "alice"}
		cnt, err := e.Namespace(ns).Table(po).Update(context.Background(), filter, M{"$set": M{"sub": "leo"}})
		assert.Equal(t, 1, cnt)
		assert.Nil(t, err)

		r, err := e.Namespace(ns).Table(po).Find(context.Background(), M{"sub": "leo"})
		expected := `[map[act:read eft:allow obj:data1 sub:leo]]`
		assert.Nil(t, err)
		assert.Equal(t, expected, fmt.Sprintf("%v", r))
	})
	t.Run("basic2", func(t *testing.T) {
		p := "./__test_tmp__/namespace_update_basic2"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
		_ = insertMany(e, ns, po, NewObjectIds(2), []A{{"alice", "data1", "read"}, {"alice", "data2", "write"}}, t)

		filter := M{"sub": "alice"}
		cnt, err := e.Namespace(ns).Table(po).Update(context.Background(), filter, M{"$set": M{"sub": "leo"}})
		assert.Equal(t, 2, cnt)
		assert.Nil(t, err)

		r, err := e.Namespace(ns).Table(po).Find(context.Background(), M{"sub": "leo"})
		expected := `[map[act:read eft:allow obj:data1 sub:leo] map[act:write eft:allow obj:data2 sub:leo]]`
		assert.Nil(t, err)
		assert.Equal(t, expected, fmt.Sprintf("%v", r))
	})
}

func TestEngine_Delete(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		p := "./__test_tmp__/namespace_delete_basic2"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
		oids := NewObjectIds(2)
		_ = insertMany(e, ns, po, oids, []A{{"alice", "data1", "read"}, {"alice", "data2", "write"}}, t)

		deleted, err := e.Namespace(ns).Table(po).Delete(context.Background(), M{"sub": "alice"})
		assert.Nil(t, err)
		assert.Equal(t, 2, len(deleted))
		assert.Equal(t, deleted, oids)
	})
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
	_ = insertMany(e, ns, po, NewObjectIds(2), []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)

	result, err := e.Namespace(ns).Table(po).EnforceOne(context.Background(), A{"alice", "data1", "read"})
	assert.Nil(t, err)
	assert.True(t, result)

	result, err = e.Namespace(ns).Table(po).EnforceOne(context.Background(), A{"alice", "data1", "write"})
	assert.Nil(t, err)
	assert.False(t, result)
}

func TestEngine_Analyze(t *testing.T) {
	p := "./__test_tmp__/ns_analyze"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	ns := "test_namespace"
	po := "p"
	// setup basic model
	setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
	_ = insertMany(e, ns, po, NewObjectIds(2), []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)

	result, err := e.Namespace(ns).Table(po).Analyze(context.Background(), A{"alice", "data1", "read"})
	assert.Nil(t, err)
	expected := `MatcherPlan | Type: AllowOverride
└─LimitPlan | Limit:1
  └─TableRowIdScan
    └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (r.obj == p.obj)) && (r.act == p.act)) && (p.eft == "allow"))`
	assert.Equal(t, expected, result)
	fmt.Println(result)
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
		_ = insertMany(e, ns, po, NewObjectIds(2), []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
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
		_ = insertMany(e, ns, po, NewObjectIds(2), []A{{"alice", "data1", "read"}, {"bob", "data2", "write"}}, t)
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
