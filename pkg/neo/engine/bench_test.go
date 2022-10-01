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
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestEngine_InsertMan(t *testing.T) {
	p := "./__test_tmp__/insert_many"
	e := setupEngine(t, p)
	defer func() {
		e.Close()
		os.RemoveAll(p)
	}()
	ns := "test_namespace"
	po := "p"
	// setup basic model
	setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
	// prepares policies
	var pPolicies [][]interface{}
	var tids []primitive.ObjectID
	for j := 0; j <= 100; j++ { // 1000,000
		for i := 0; i <= 1000; i++ {
			v := []interface{}{
				fmt.Sprintf("user%d", j),
				fmt.Sprintf("data%d", i),
				"read",
			}
			pPolicies = append(pPolicies, v)
			tids = append(tids, primitive.NewObjectID())
		}
	}
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(po)
	r, err := pTab.InsertMany(context.Background(), tids, pPolicies)
	assert.Nil(t, err)
	assert.Equal(t, len(pPolicies), len(r))
	filter := map[string]interface{}{"sub": "user100", "obj": "data1000"}
	res := find(e, ns, po, filter, t)
	fmt.Println(res)

	t.Run("find", func(t *testing.T) {
		n := time.Now()
		find(e, ns, po, filter, t)
		fmt.Println(time.Since(n))
	})
}

func BenchmarkFindExecutor(t *testing.B) {
	t.Run("basic", func(t *testing.B) {
		p := "./__test_tmp__/find_bench"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)

		// prepares policies
		var pPolicies [][]interface{}
		var tids []primitive.ObjectID
		for j := 0; j <= 100; j++ { // 100,000
			for i := 0; i <= 1000; i++ {
				v := []interface{}{
					fmt.Sprintf("user%d", j),
					fmt.Sprintf("data%d", i),
					"read",
				}
				pPolicies = append(pPolicies, v)
				tids = append(tids, primitive.NewObjectID())

			}
		}
		_ = insertMany(e, ns, po, tids, pPolicies, t)
		filter := map[string]interface{}{"sub": "user100", "obj": "data100"}
		find(e, ns, po, filter, t)

		t.ResetTimer()
		t.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				find(e, ns, po, filter, t)
			}
		})
	})
}

func BenchmarkEnforce(t *testing.B) {
	t.Run("basic", func(t *testing.B) {
		p := "./__test_tmp__/enforce_bench"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)

		// prepares policies
		var pPolicies [][]interface{}
		var tids []primitive.ObjectID
		for j := 0; j <= 100; j++ { // 100,000
			for i := 0; i <= 1000; i++ {
				v := []interface{}{
					fmt.Sprintf("user%d", j), //sub
					fmt.Sprintf("data%d", i), //obj
					"read",                   //act
				}
				pPolicies = append(pPolicies, v)
				tids = append(tids, primitive.NewObjectID())
			}
		}
		_ = insertMany(e, ns, po, tids, pPolicies, t)
		t.ResetTimer()

		req := A{"user100", "data100", "read"}
		e.Namespace(ns).Table(po).EnforceOne(context.Background(), req)
		t.ResetTimer()
		tab := e.Namespace(ns).Table(po)
		t.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tab.EnforceOne(context.Background(), req)
			}
		})
	})
}

func BenchmarkInsertExecutor(t *testing.B) {
	t.Run("basic", func(t *testing.B) {
		p := "./__test_tmp__/insert_bench"
		e := setupEngine(t, p)
		defer func() {
			e.Close()
			os.RemoveAll(p)
		}()
		ns := "test_namespace"
		po := "p"
		// setup basic model
		setModel(e, ns, "../../../examples/assets/model/basic_model.conf", t)
		t.ResetTimer()
		t.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				t.StopTimer()
				uid := primitive.NewObjectID()
				oid := primitive.NewObjectID()
				v := []interface{}{
					fmt.Sprintf("user%d", uid),
					fmt.Sprintf("data%d", oid),
					"read",
				}
				t.StartTimer()
				insertOne(e, ns, po, primitive.NewObjectID(), v, t)
			}
		})
	})
}
