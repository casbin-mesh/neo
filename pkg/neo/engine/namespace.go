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

type namespace struct {
	engine Engine
	dbName string
}

func (n namespace) Table(name string) Table {
	return &table{
		dbName:    n.dbName,
		tableName: name,
		engine:    n.engine,
	}
}

func NewNamespace(engine Engine, dbname string) Namespace {
	return &namespace{engine, dbname}
}
