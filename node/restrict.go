package node

import (
	"github.com/bytedance/ns-x/v2/base"
	"math"
	"time"
)

// RestrictNode simulate a node with limited ability
// Once packets through a RestrictNode reaches the limit(in bps or pps), the later packets will be put in a queue
// Once the queue overflow, later packets will be discarded
type RestrictNode struct {
	*BasicNode
	ppsLimit, bpsLimit                 float64
	queueBytesLimit, queuePacketsLimit int64
	queueBytes, queuePackets           int64
	busyTime                           time.Time
}

// NewRestrictNode create a new RestrictNode with the given options
func NewRestrictNode(options ...Option) *RestrictNode {
	n := &RestrictNode{
		BasicNode:         &BasicNode{},
		ppsLimit:          -1,
		bpsLimit:          -1,
		queueBytesLimit:   -1,
		queuePacketsLimit: -1,
	}
	apply(n, options...)
	if n.ppsLimit <= 0 && n.bpsLimit <= 0 {
		panic("a restrict node must be limited in pps/bps")
	}
	return n
}

func (n *RestrictNode) Transfer(packet base.Packet, now time.Time) []base.Event {
	busy := false
	t := now
	if n.busyTime.After(now) {
		t = n.busyTime
		busy = true
	}
	if busy {
		if n.queueBytesLimit >= 0 && n.queueBytes+int64(packet.Size()) > n.queueBytesLimit {
			return nil
		}
		if n.queuePacketsLimit >= 0 && n.queuePackets+1 > n.queuePacketsLimit {
			return nil
		}
	}
	step := math.Max(1.0/n.ppsLimit, float64(packet.Size())/n.bpsLimit)
	delta := time.Duration(step * float64(time.Second))
	n.busyTime = t.Add(delta)
	if busy {
		n.queueBytes += int64(packet.Size())
		n.queuePackets++
		return base.Aggregate(
			base.NewFixedEvent(func(t time.Time) []base.Event {
				n.queueBytes -= int64(packet.Size())
				n.queuePackets--
				return n.actualTransfer(packet, n, n.GetNext()[0], t)
			}, t),
		)
	}
	return n.actualTransfer(packet, n, n.GetNext()[0], t)
}

func (n *RestrictNode) Check() {
	if len(n.GetNext()) != 1 {
		panic("restrict node can only has single connection")
	}
}

// WithPPSLimit create an option set/overwrite pps limit and queue limit in packets to nodes applied
// once flow of the node calculated in packets/second reach pps limit, further packets will be put into the queue
// once total count of packets in the queue reach the queue packets limit, further packets will be ignored
// node applied must be a RestrictNode
// set limit to -1 means unlimited
func WithPPSLimit(ppsLimit float64, queuePacketsLimit int64) Option {
	return func(node base.Node) {
		n, ok := node.(*RestrictNode)
		if !ok {
			panic("cannot set pps limit")
		}
		n.ppsLimit = ppsLimit
		n.queuePacketsLimit = queuePacketsLimit
	}
}

// WithBPSLimit create an option set/overwrite bps limit and queue limit in bytes to nodes applied
// once flow of the node calculated in bytes/second reach bps limit, further packets will be put into the queue
// once total size of packets in the queue reach the queue size limit, further packets will be ignored
// node applied must be a RestrictNode
// set limit to -1 means unlimited
func WithBPSLimit(bpsLimit float64, queueBytesLimit int64) Option {
	return func(node base.Node) {
		n, ok := node.(*RestrictNode)
		if !ok {
			panic("cannot set pps limit")
		}
		n.bpsLimit = bpsLimit
		n.queueBytesLimit = queueBytesLimit
	}
}
