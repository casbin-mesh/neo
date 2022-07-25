package index

import (
	"github.com/casbin-mesh/neo/pkg/storage/mem/index/art"
	"sync"
)

type secondaryIndex[T any] struct {
	tree    *art.Tree[T]
	idlayer indirectLayer[T]
	mu      sync.Mutex
}

func (sIndex *secondaryIndex[T]) build_secondary_index(idlayer *indirectLayer[T], iterator *art.Iterator[T]) {
	sIndex.tree = iterator.Tree
	sIndex.mu.Lock()
	defer sIndex.mu.Unlock()
	for t := iterator.Itr; t.Next(); {
		sIndex.tree.Insert(t.Key(), t.Value())
		//todo : add into indirect layer
	}
}
