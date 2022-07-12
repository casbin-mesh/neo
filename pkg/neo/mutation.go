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

import (
	"bufio"
	"github.com/casbin-mesh/neo/pkg/db"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"io"
	"strings"
)

type mutation struct {
	readTs   uint64
	commitTs uint64
	ctx      *mutationCtx
	txn      db.Txn
}

type KeyValue interface {
	Key() []byte
	Value
}

type Value interface {
	ValueCopy() []byte
}

type ParseModelOptions struct {
	// Warm if true, engine caches this model definitions, and builds functions that will inject into matchers
	Warm bool
}

func (m *mutation) parseAndBuildMatcher(key, def string) {
	// parse matcher
	// inject functions
}

// ParseModel parses model string from user inputs
func (m *mutation) ParseModel(reader io.Reader, opts ParseModelOptions) error {
	c, err := utils.NewParse(bufio.NewReader(reader))
	if err != nil {
		return err
	}
	rd := c.RequestDef()
	ns := m.ctx.namespace

	// TODO: uses yaac to parse model. for now, we treat all definitions as strings.

	// ---------------------------- Request definitions
	// key patten: | namespace \x00 | name{prefix r} \x00 |
	// parses request schema definitions
	for key, def := range rd {
		// TODO(weny): cache the builder
		builder := bschema.NewReaderWriter(ns, []byte(key))
		fields := strings.Split(def, ",")
		for _, field := range fields {
			builder.Append(bsontype.String, []byte(field))
		}

		err = m.txn.Set(builder.EncodeKey(), builder.EncodeVal())
		if err != nil {
			return err
		}
	}

	// ---------------------------- Policy definitions
	// key patten: | namespace \x00 | name{prefix p} \x00 |
	// parses policies schema definitions
	pd := c.PolicyDef()
	for key, def := range pd {
		// TODO(weny): cache the builder
		builder := bschema.NewReaderWriter(ns, []byte(key))
		fields := strings.Split(def, ",")
		for _, field := range fields {
			builder.Append(bsontype.String, []byte(field))
		}
		err = m.txn.Set(builder.EncodeKey(), builder.EncodeVal())
		if err != nil {
			return err
		}
		if opts.Warm {

		}
	}

	// ---------------------------- Optional Role definitions
	// key patten: | namespace \x00 | name{prefix g} \x00 |
	// parses policies schema definitions
	gd := c.RoleDef()
	for key, def := range gd {
		// TODO(weny): cache the builder
		builder := bschema.NewReaderWriter(ns, []byte(key))
		fields := strings.Split(def, ",")
		for _, field := range fields {
			builder.Append(bsontype.String, []byte(field))
		}
		err = m.txn.Set(builder.EncodeKey(), builder.EncodeVal())
		if err != nil {
			return err
		}
		if opts.Warm {

		}
	}

	// ---------------------------- Optional Effect definitions
	// key patten: | namespace \x00 | name{prefix e} \x00 |
	// parses effect definitions
	// TODO: cache effect
	pe := c.PolicyEffect()
	for key, def := range pe {
		if err = m.txn.Set(utils.CString(ns, []byte(key)), []byte(def)); err != nil {
			return err
		}
		if opts.Warm {

		}
	}

	// ---------------------------- Matcher definitions
	// key patten: | namespace \x00 | name{prefix m} \x00 |
	// parses matcher schema definitions
	// TODO: cache matcher
	ms := c.Matchers()
	for key, def := range ms {
		if err = m.txn.Set(utils.CString(ns, []byte(key)), []byte(def)); err != nil {
			return err
		}
		if opts.Warm {
			m.parseAndBuildMatcher(key, def)
		}
	}
	return nil
}

// AddPolicy add policy
// storageKey: {namespace}\x00{schemaName}\x00{objectID}
func (m *mutation) AddPolicy(schemaName []byte, reader btuple.Reader) {
	// builds key values
	// updates secondary indexes
}

func (m *mutation) RemovePolicy(schemaName []byte, target btuple.Reader) (removed bool) {
	// search in secondary indexes
	// case1: found
	//	remove it from secondary indexes TODO: shall we?
	//	What if another txn tries to search the older value in secondary indexes
	//  which should be indexed.
	//  TODO: Or we should search the value in KV by the primary key retrieved from the secondary indexes

	//  remove value from kv
	// case2: not found, return false
	return false
}

func (m *mutation) UpdatePolicy(schemaName []byte, old, new btuple.Reader) (updated bool) {
	// search in secondary indexes
	// case1: found
	// 	TODO: Or we should search the value in KV by the primary key retrieved from the secondary indexes
	//	update value in kv
	// case2: not found, return false
	return false
}

func (m *mutation) CommitAt(commitTs uint64) error {
	// commit mutation txn
	// commit indexes changes
	return m.txn.CommitAt(commitTs, nil)
}

func (m *mutation) Abort() error {
	// discard txn changes
	// discard indexes changes
	return nil
}
