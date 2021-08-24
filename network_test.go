package ns_x

import (
	"container/heap"
	"github.com/bytedance/ns-x/base"
	"math/rand"
	"testing"
	"time"
)

func nop(time.Time) []base.Event {
	return nil
}

func BenchmarkEventLoop(b *testing.B) {
	network := NewNetwork([]base.Node{})
	events := make([]base.Event, b.N)
	now := time.Now()
	for i := 0; i < b.N; i++ {
		events[i] = base.NewFixedEvent(nop, now.Add(-time.Duration(rand.Int()%1000)*time.Second))
	}
	h := &base.EventHeap{Storage: events}
	heap.Init(h)
	b.ResetTimer()
	network.eventLoop(h)
}
