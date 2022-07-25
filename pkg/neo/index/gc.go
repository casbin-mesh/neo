package index

type garbageCollector[T any] struct {
	epManger  *epochManager[T]
	waterMark uint64
}

func (gc *garbageCollector[T]) runGC() {
	for _, epch := range gc.epManger.epoches {
		if epch.activeTxnsCnt == 0 {
			gc.cleanEpoch(&epch)
		}
	}
}

func (gc *garbageCollector[T]) cleanEpoch(epch *epoch[T]) {

}
