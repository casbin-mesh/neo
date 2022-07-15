package index

import "sync"

type indirectLayer[T any] struct {
	table      map[string]Value[T]
	sync.Mutex //to avoid datarace
}

func (il *indirectLayer[T]) add_into_indirLayer(str string, val Value[T]) {
	oldVal, ok := il.table[str]
	if ok {
		val.next = &oldVal
	} else {
		il.table[str] = val
	}
}
