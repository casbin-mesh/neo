package index

import (
	"sync"
)
import "github.com/casbin-mesh/neo/pkg/storage/mem/index/art"

type mapper[T any] struct {
	treeLeaf  art.LEAF[T]
	chainHead VersionChainHead[T]
}

type indirectLayer[T any] struct {
	table      []mapper[T]
	sync.Mutex //to avoid datarace
}
