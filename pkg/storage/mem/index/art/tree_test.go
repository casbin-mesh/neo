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

package art

import (
	"bufio"
	"bytes"
	"encoding/binary"
	gBtree "github.com/google/btree"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/btree"

	"os"
	"testing"
)

func TestArtTree_Insert(t *testing.T) {
	tree := NewArtTree()
	// insert one key
	inserted, updated := tree.Insert([]byte("I'm Key"), Value("I'm Value"))
	assert.False(t, updated)
	assert.Equal(t, Value("I'm Value"), inserted)

	// search it
	value, found := tree.Search(Key("I'm Key"))
	assert.Equal(t, Value("I'm Value"), value)
	assert.True(t, found)
	//insert another key
	inserted, updated = tree.Insert(Key("I'm Key2"), Value("I'm Value2"))
	assert.False(t, updated)
	assert.Equal(t, Value("I'm Value2"), inserted)

	// search it
	value, found = tree.Search(Key("I'm Key2"))
	assert.Equal(t, Value("I'm Value2"), value)

	// should be found
	value, found = tree.Search(Key("I'm Key"))
	assert.Equal(t, Value("I'm Value"), value)
}

type Tests struct {
	key   Key
	value Value
}

func TestArtTree_Insert3(t *testing.T) {
	tree := NewArtTree()
	tree.Insert(Key("sharedKey::1"), Value("value1"))
	tree.Insert(Key("sharedKey::2"), Value("value1"))
	tree.Insert(Key("sharedKey::3"), Value("value1"))
	tree.Insert(Key("sharedKey::4"), Value("value1"))

	tree.Insert(Key("sharedKey::1::created_at"), Value("created_at_value1"))

	tree.Insert(Key("sharedKey::1::name"), Value("name_value1"))

	value, found := tree.Search(Key("sharedKey::1::created_at"))
	assert.True(t, found)
	assert.Equal(t, Value("created_at_value1"), value)
}

func TestArtTree_Insert2(t *testing.T) {
	tree := NewArtTree()
	sets := []Tests{{
		Key("sharedKey::1"), Value("value1"),
	}, {
		Key("sharedKey::2"), Value("value2"),
	}, {
		Key("sharedKey::3"), Value("value3"),
	}, {
		Key("sharedKey::4"), Value("value4"),
	}, {
		Key("sharedKey::1::created_at"), Value("created_at_value1"),
	}, {
		Key("sharedKey::1::name"), Value("name_value1"),
	},
	}
	for _, set := range sets {
		tree.Insert(set.key, set.value)
	}
	for _, set := range sets {
		value, found := tree.Search(set.key)
		assert.True(t, found)
		assert.Equal(t, set.value, value)
	}
}

func TestArtTree_Update(t *testing.T) {
	tree := NewArtTree()
	key := Key("I'm Key")

	// insert an entry
	tree.Insert(key, Value("I'm Value"))

	// should be found
	value, found := tree.Search(key)
	assert.Equal(t, Value("I'm Value"), value)
	assert.Truef(t, found, "The inserted key should be found")

	// try update inserted key
	_, updated := tree.Insert(key, Value("Value Updated"))
	//assert.Equal(t, Value("Value Updated"), value)
	assert.Truef(t, updated, "The inserted key should be updated")

	value, found = tree.Search(key)
	assert.Truef(t, found, "The inserted key should be found")
	assert.Equal(t, Value("Value Updated"), value)
}

func TestArtTree_InsertSimilarPrefix(t *testing.T) {
	tree := NewArtTree()
	tree.Insert(Key{1}, []byte{1})
	tree.Insert(Key{1, 1}, []byte{1, 1})

	v, found := tree.Search(Key{1, 1})
	assert.True(t, found)
	assert.Equal(t, Value([]byte{1, 1}), v)
}

func TestArtTree_InsertMoreKey(t *testing.T) {
	tree := NewArtTree()
	keys := []Key{{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, {1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1}, {2, 1, 1}}
	for _, key := range keys {
		tree.Insert(key, Value(key))
	}
	for _, key := range keys {
		value, found := tree.Search(key)
		assert.Equal(t, Value(key), value)
		assert.True(t, found)
	}
}

func TestArtTree_Remove(t *testing.T) {
	tree := NewArtTree()
	_, deleted := tree.Remove(Key("wrong-key"))
	assert.False(t, deleted)

	tree.Insert(Key("sharedKey::1"), Value("value1"))
	tree.Insert(Key("sharedKey::2"), Value("value2"))

	value, deleted := tree.Remove(Key("sharedKey::2"))
	assert.Equal(t, Value("value2"), value)
	assert.True(t, deleted)

	value, deleted = tree.Remove(Key("sharedKey::3"))
	assert.Nil(t, value)
	assert.False(t, deleted)

	tree.Insert(Key("sharedKey::3"), Value("value3"))

	value, deleted = tree.Remove(Key("sharedKey"))
	assert.Nil(t, value)
	assert.False(t, deleted)

	tree.Insert(Key("sharedKey::4"), Value("value3"))

	value, deleted = tree.Remove(Key("sharedKey::5::xxx"))
	assert.Nil(t, value)
	assert.False(t, deleted)

	value, deleted = tree.Remove(Key("sharedKey::4xsfdasd"))
	assert.Nil(t, value)
	assert.False(t, deleted)

	tree.Insert(Key("sharedKey::4::created_at"), Value("value3"))
	value, deleted = tree.Remove(Key("sharedKey::4::created_at"))
	assert.True(t, deleted)
}

func TestArtTree_Search(t *testing.T) {
	tree := NewArtTree()
	value, found := tree.Search(Key("wrong-key"))
	assert.Nil(t, value)
	assert.False(t, found)

	tree.Insert(Key("sharedKey::1"), Value("value1"))

	value, found = tree.Search(Key("sharedKey"))
	assert.Nil(t, value)
	assert.False(t, found)
	value, found = tree.Search(Key("sharedKey::2"))
	assert.Nil(t, value)
	assert.False(t, found)

	tree.Insert(Key("sharedKey::2"), Value("value1"))

	value, found = tree.Search(Key("sharedKey::3"))
	assert.Nil(t, value)
	assert.False(t, found)

	value, found = tree.Search(Key("sharedKey"))
	assert.Nil(t, value)
	assert.False(t, found)
}

type CheckPoint struct {
	name       string
	totalNodes int
	expected   Kind
}

func TestArtTree_Grow(t *testing.T) {
	checkPoints := []CheckPoint{
		{totalNodes: 5, expected: Node16, name: "node4 growing test"},
		{totalNodes: 17, expected: Node48, name: "node16 growing test"},
		{totalNodes: 49, expected: Node256, name: "node256 growing test"},
	}
	for _, point := range checkPoints {
		tree := NewArtTree()
		g := NewKeyValueGenerator()
		for i := 0; i < point.totalNodes; i++ {
			tree.Insert(g.next())
		}
		assert.Equal(t, int(point.totalNodes), tree.size)
		assert.Equalf(t, point.expected, tree.root.kind, "exected kind %s got %s", point.expected, tree.root.kind)
		g.resetCur()
		for i := 0; i < point.totalNodes; i++ {
			k, v := g.next()
			got, found := tree.Search(k)
			assert.True(t, found, "should found inserted (%v,%v) in test %s", k, v, point.name)
			assert.Equal(t, v, got, "should equal inserted (%v,%v) in test %s", k, v, point.name)
		}
	}
}

func TestArtTree_Shrink(t *testing.T) {
	tree := NewArtTree()
	g := NewKeyValueGenerator()
	// fill up an 256 node
	for i := 0; i < node256Max; i++ {
		tree.Insert(g.next())
	}
	// check inserted
	g.resetCur()
	for i := 0; i < node256Max; i++ {
		k, v := g.next()
		got, found := tree.Search(k)
		assert.True(t, found)
		assert.Equal(t, v, got)
	}
	// deleting nodes
	for i := 255; i >= 0; i-- {
		assert.Equal(t, i+1, tree.size)
		k, v := g.prev()
		old, deleted := tree.Remove(k)
		assert.True(t, deleted)
		assert.Equal(t, v, old)
		// n.go L439-449
		switch tree.Size() {
		case 36:
			assert.Equal(t, Node48, tree.root.kind)
		case 11:
			assert.Equal(t, Node16, tree.root.kind)
		case 2:
			assert.Equal(t, Node4, tree.root.kind)
		case 0:
			assert.Nil(t, tree.root)

		}
	}
}

func TestArtTree_ShrinkConcatenating(t *testing.T) {
	tree := NewArtTree()
	tree.Insert(Key("sharedKey::1"), Value("value1"))
	tree.Insert(Key("sharedKey::2"), Value("value1"))
	tree.Insert(Key("sharedKey::3"), Value("value1"))
	tree.Insert(Key("sharedKey::4"), Value("value1"))

	tree.Insert(Key("sharedKey::1::nested::name"), Value("created_at_value1"))
	tree.Insert(Key("sharedKey::1::nested::job"), Value("name_value1"))

	tree.Insert(Key("sharedKey::1::nested::name::firstname"), Value("created_at_value1"))
	tree.Insert(Key("sharedKey::1::nested::name::lastname"), Value("created_at_value1"))

	tree.Remove(Key("sharedKey::1::nested::name"))

	_, found := tree.Search(Key("sharedKey::1::nested::name"))
	assert.False(t, found)
}

func TestArtTree_LargeKeyShrink(t *testing.T) {
	tree := NewArtTree()
	g := NewLargeKeyValueGenerator([]byte("this a very long sharedKey::"))
	// fill up an 256 node
	for i := 0; i < node256Max; i++ {
		tree.Insert(g.next())
	}
	// check inserted
	g.resetCur()
	for i := 0; i < node256Max; i++ {
		k, v := g.next()
		got, found := tree.Search(k)
		assert.True(t, found)
		assert.Equal(t, v, got)
	}
	// deleting nodes
	for i := 255; i >= 0; i-- {
		assert.Equal(t, i+1, tree.size)
		k, v := g.prev()
		old, deleted := tree.Remove(k)
		assert.True(t, deleted)
		assert.Equal(t, v, old)
		// n.go L439-449
		switch tree.Size() {
		case 36:
			assert.Equal(t, Node48, tree.root.kind)
		case 11:
			assert.Equal(t, Node16, tree.root.kind)
		case 2:
			assert.Equal(t, Node4, tree.root.kind)
		case 0:
			assert.Nil(t, tree.root)

		}
	}
}

type largeKeyValueGenerator struct {
	cur       int64
	generator func([]byte) []byte
	prefix    []byte
}

func NewLargeKeyValueGenerator(prefix []byte) *largeKeyValueGenerator {
	return &largeKeyValueGenerator{
		cur: 0,
		generator: func(input []byte) []byte {
			return input
		},
		prefix: prefix,
	}
}

func (g *largeKeyValueGenerator) get(cur int64) (Key, Value) {
	prefixLen := len(g.prefix)
	var buf = make([]byte, prefixLen+8)
	copy(buf[:], g.prefix)
	binary.PutVarint(buf[prefixLen:], cur)
	return buf, g.generator(buf)
}

func (g *largeKeyValueGenerator) prev() (Key, Value) {
	g.cur--
	k, v := g.get(g.cur)
	return k, v
}

func (g *largeKeyValueGenerator) next() (Key, Value) {
	k, v := g.get(g.cur)
	g.cur++
	return k, v
}

func (g *largeKeyValueGenerator) reset() {
	g.cur = 0
}

func (g *largeKeyValueGenerator) resetCur() {
	g.cur = 0
}

type keyValueGenerator struct {
	cur       int
	generator func([]byte) []byte
}

func (g keyValueGenerator) getValue(key Key) Value {
	return g.generator(key)
}

func (g *keyValueGenerator) prev() (Key, Value) {
	g.cur--
	var buf [8]byte
	binary.PutVarint(buf[:], int64(g.cur))
	k, v := []byte{byte(g.cur)}, g.generator(buf[:])
	return k, v
}

func (g *keyValueGenerator) next() (Key, Value) {
	var buf [8]byte
	binary.PutVarint(buf[:], int64(g.cur))
	k, v := []byte{byte(g.cur)}, g.generator(buf[:])
	g.cur++
	return k, v
}

func (g *keyValueGenerator) setCur(c int) {
	g.cur = c
}

func (g *keyValueGenerator) resetCur() {
	g.setCur(0)
}

func NewKeyValueGenerator() *keyValueGenerator {
	return &keyValueGenerator{cur: 0, generator: func(input []byte) []byte {
		return input
	}}
}

func TestArtTree_InsertOneAndDeleteOne(t *testing.T) {
	tree := NewArtTree()
	g := NewKeyValueGenerator()
	k, v := g.next()

	// insert one
	tree.Insert(k, v)

	// delete inserted
	oldValue, deleted := tree.Remove(k)
	assert.Equal(t, v, oldValue)
	assert.True(t, deleted)

	// should be not found
	got, found := tree.Search(k)
	assert.Nil(t, got)
	assert.False(t, found)

	// insert another one
	k, v = g.next()
	tree.Insert(k, v)

	// try to delete a non-exist key
	oldValue, deleted = tree.Remove(Key("wrong-key"))
	assert.Nil(t, oldValue)
	assert.False(t, deleted)
}

func TestArtTest_InsertAndDelete(t *testing.T) {
	tree := NewArtTree()
	g := NewKeyValueGenerator()
	// insert 1000
	for i := 0; i < 100; i++ {
		_, _ = tree.Insert(g.next())
	}
	g.resetCur()
	// check inserted kv
	for i := 0; i < 100; i++ {
		k, v := g.next()
		got, found := tree.Search(k)
		assert.Equalf(t, v, got, "should insert key-value (%v:%v)", k, v)
		assert.True(t, found)
	}
}

func TestArtTree_InsertLargeKeyAndDelete(t *testing.T) {
	tree := NewArtTree()
	g := NewLargeKeyValueGenerator([]byte("largeThanMax"))
	// insert 1000
	for i := 0; i < 100; i++ {
		_, _ = tree.Insert(g.next())
	}
	g.reset()
	// check inserted kv
	for i := 0; i < 100; i++ {
		k, v := g.next()
		got, found := tree.Search(k)
		assert.Equalf(t, v, got, "should insert key-value (%v:%v)", k, v)
		assert.True(t, found)
	}
}

// Benchmark
func loadTestFile(path string) [][]byte {
	file, err := os.Open(path)
	if err != nil {
		panic("Couldn't open " + path)
	}
	defer file.Close()

	var words [][]byte
	reader := bufio.NewReader(file)
	for {
		if line, err := reader.ReadBytes(byte('\n')); err != nil {
			break
		} else {
			if len(line) > 0 {
				words = append(words, line[:len(line)-1])
			}
		}
	}
	return words
}

type KV struct {
	Key   []byte
	Value []byte
}

func (kv KV) Less(than gBtree.Item) bool {
	return bytes.Compare(kv.Key, than.(KV).Key) < 0
}

func Compare(a, b KV) bool {
	return bytes.Compare(a.Key, b.Key) < 0
}

func BenchmarkWordsBTreeInsert(b *testing.B) {
	words := loadTestFile("../../../../../test/assets/words.txt")
	var kv []KV
	for _, word := range words {
		kv = append(kv, KV{Key: word, Value: word})
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree := btree.NewGeneric[KV](Compare)
		for _, pair := range kv {
			tree.Set(pair)
		}
	}
}

func BenchmarkWordsArtInsert(b *testing.B) {
	words := loadTestFile("../../../../../test/assets/words.txt")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree := NewArtTree()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkWordsMapInsert(b *testing.B) {
	words := loadTestFile("../../../../../test/assets/words.txt")
	var strWords []string
	for _, word := range words {
		strWords = append(strWords, string(word))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := make(map[string]string)
		for _, w := range strWords {
			m[w] = w
		}
	}
}

func BenchmarkWordsBTreeSearch(b *testing.B) {
	words := loadTestFile("../../../../../test/assets/words.txt")
	var kv []KV
	for _, word := range words {
		kv = append(kv, KV{Key: word, Value: word})
	}
	tree := btree.NewGeneric[KV](Compare)

	for n := 0; n < b.N; n++ {
		for _, pair := range kv {
			tree.Set(pair)
		}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, pair := range kv {
			tree.Get(pair)
		}
	}
}

func BenchmarkWordsArtSearch(b *testing.B) {
	words := loadTestFile("../../../../../test/assets/words.txt")
	tree := NewArtTree()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, w := range words {
			tree.Search(w)
		}
	}
}

func BenchmarkWordsMapSearch(b *testing.B) {
	words := loadTestFile("../../../../../test/assets/words.txt")
	var strWords []string
	for _, word := range words {
		strWords = append(strWords, string(word))
	}
	m := make(map[string]string)
	for n := 0; n < b.N; n++ {
		for _, w := range strWords {
			m[w] = w
		}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, w := range strWords {
			_ = m[w]
		}
	}
}
