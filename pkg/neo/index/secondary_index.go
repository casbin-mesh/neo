package index

import "github.com/casbin-mesh/neo/pkg/storage/mem/index/art"

type secondaryIndex[T any] struct {
	tree    art.Tree[T]
	idlayer indirectLayer[T]
}
