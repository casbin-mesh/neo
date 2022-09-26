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
	"os"
	"testing"
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
	for i := 0; i < 1_000_000; i++ {
		v := []interface{}{
			fmt.Sprintf("user%d", i),
			fmt.Sprintf("data%d", i),
			"read",
		}
		pPolicies = append(pPolicies, v)
	}
	// open namespace
	nsHandle := e.Namespace(ns)
	// the default policies table
	pTab := nsHandle.Table(po)
	_, _ = pTab.InsertMany(context.Background(), pPolicies)
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
		for i := 0; i < 10_000_000; i++ {
			v := []interface{}{
				fmt.Sprintf("user%d", i),
				fmt.Sprintf("data%d", i),
				"read",
			}
			pPolicies = append(pPolicies, v)
		}
		_ = insertMany(e, ns, po, pPolicies, t)

		t.ResetTimer()
		filter := map[string]interface{}{"sub": "user1000"}
		t.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				find(e, ns, po, filter, t)
			}
		})
	})
}
