package ns_x

import (
	"container/heap"
	"github.com/bytedance/ns-x/v2/base"
	"github.com/bytedance/ns-x/v2/tick"
	"go.uber.org/atomic"
	"runtime"
	"sync"
)

// Network Indicates a simulated network, which contains some simulated nodes
type Network struct {
	nodes   []base.Node
	clock   tick.Clock
	buffer  *base.EventBuffer
	running *atomic.Bool
	wg      *sync.WaitGroup
}

// NewNetwork creates a network with the given nodes, connections of nodes should be already established.
func NewNetwork(nodes []base.Node, clock tick.Clock) *Network {
	return &Network{
		nodes:   nodes,
		clock:   clock,
		buffer:  base.NewEventBuffer(),
		running: atomic.NewBool(false),
		wg:      &sync.WaitGroup{},
	}
}

// fetch events from nodes in the network, and put them into given heap
func (n *Network) fetch(packetHeap heap.Interface) {
	n.buffer.Reduce(func(packet base.Event) {
		heap.Push(packetHeap, packet)
	})
}

// drain the given heap if possible, and process the events available
func (n *Network) drain(packetHeap *base.EventHeap) {
	now := n.clock()
	for !packetHeap.IsEmpty() {
		p := packetHeap.Peek()
		t := p.Time()
		if t.After(now) {
			break
		}
		events := p.Action()(t)
		heap.Pop(packetHeap)
		for _, event := range events {
			heap.Push(packetHeap, event)
		}
	}
}

// block until clear the given heap
func (n *Network) clear(packetHeap *base.EventHeap) {
	for !packetHeap.IsEmpty() {
		n.drain(packetHeap)
	}
}

// eventLoop Main polling loop of network
func (n *Network) eventLoop(packetHeap *base.EventHeap) {
	println("network main loop start")
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	n.wg.Add(1)
	defer n.wg.Done()
	for n.running.Load() {
		n.fetch(packetHeap)
		n.drain(packetHeap)
	}
	n.clear(packetHeap)
	println("network main loop end")
}

// Start the network to enable event process
func (n *Network) Start(events ...base.Event) {
	if n.running.Load() {
		return
	}
	n.running.Store(true)
	for _, node := range n.nodes {
		node.Check()
	}
	h := &base.EventHeap{Storage: events}
	heap.Init(h)
	go n.eventLoop(h)
}

// Stop the network, release resources
func (n *Network) Stop() {
	n.running.Store(false)
	n.wg.Wait()
}

// Event insert the given event
func (n *Network) Event(events ...base.Event) {
	n.buffer.Insert(events...)
}

// Nodes return all nodes managed by the network
func (n *Network) Nodes() []base.Node {
	return n.nodes
}
