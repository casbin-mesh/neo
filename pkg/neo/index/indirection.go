package index

import (
	"sync"
)
import "github.com/casbin-mesh/neo/pkg/storage/mem/index/art"

type Mapper[T any] struct {
	treeLeaf  *art.LEAF[T]
	chainHead VersionChainHead[T]
}

type indirectLayer[T any] struct {
	table      []Mapper[T]
	sync.Mutex //to avoid datarace
}

func (ilayer *indirectLayer[T]) add_into_indirlayer(lf *art.LEAF[T], vsch VersionChainHead[T]) {
	ilayer.Lock()
	defer ilayer.Unlock()
	new_map := Mapper[T]{
		treeLeaf:  lf,
		chainHead: vsch,
	}
	ilayer.table = append(ilayer.table, new_map)
}
