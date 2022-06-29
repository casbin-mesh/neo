// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
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

import "github.com/casbin-mesh/neo/pkg/db"

type neo struct {
	db db.DB
	//TODO(weny): add meta store?
}

func New(opt Options) *neo {
	return &neo{db: opt.db}
}

func (n *neo) Mutation(ctx *mutationCtx) (*mutation, error) {
	return &mutation{
		txn: n.db.NewTransaction(true),
		ctx: ctx,
	}, nil
}
