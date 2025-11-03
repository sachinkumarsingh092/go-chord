package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sync"
)

const m = 3

func hashItem(item string) uint32 {
	digest := md5.Sum([]byte(item))
	digest1 := binary.BigEndian.Uint32(digest[0:4])
	digest2 := binary.BigEndian.Uint32(digest[4:8])
	digest3 := binary.BigEndian.Uint32(digest[8:12])
	digest4 := binary.BigEndian.Uint32(digest[12:16])
	return (digest1 ^ digest2 ^ digest3 ^ digest4) % (1 << m)
}

type Node struct {
	id          uint32
	address     string
	successor   *Node
	predecessor *Node
	finger      []*Node
	keys        map[string]string
	mutex       sync.RWMutex
}

func NewNode(address string) *Node {
	node := &Node{
		id:      hashItem(address),
		address: address,
		finger:  make([]*Node, m),
		keys:    make(map[string]string),
	}
	node.successor = node
	node.predecessor = node
	for i := range node.finger {
		node.finger[i] = node
	}
	return node
}

func (n *Node) findSuccessor(id uint32) *Node {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	pred := n.findPredecessor(id)
	return pred.successor
}

func (n *Node) findPredecessor(id uint32) *Node {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	n_prime := n
	for {
		if id > n_prime.id && id <= n_prime.successor.id ||
			(n_prime.id >= n_prime.successor.id && (id > n_prime.id || id <= n_prime.successor.id)) {
			break
		}
		cpf := n_prime.closestPrecedingFinger(id)
		if cpf == n_prime {
			break
		}
		n_prime = cpf
	}
	return n_prime
}

func (n *Node) closestPrecedingFinger(id uint32) *Node {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	for i := m - 1; i >= 0; i-- {
		if n.finger[i].id > n.id && n.finger[i].id < id ||
			(n.id >= id && n.finger[i].id > n.id) {
			return n.finger[i]
		}
	}
	return n
}

func (n *Node) join(n_prime *Node) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n_prime == nil {
		return
	}
	n.successor = n_prime.findSuccessor(n.id)
	n.predecessor = n.successor.predecessor
	n.successor.predecessor = n
	if n.predecessor != nil {
		n.predecessor.successor = n
	}
	for key, value := range n.successor.keys {
		if hashItem(key) <= n.id {
			n.keys[key] = value
			delete(n.successor.keys, key)
		}
	}
	n.buildFingerTable()
}

func (n *Node) stabilize() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	x := n.successor.predecessor
	if x != nil && (x.id > n.id && x.id < n.successor.id ||
		n.successor.id <= n.id && (x.id > n.id || x.id <= n.successor.id)) {
		n.successor = x
	}
	n.successor.notify(n)
	n.fixFingers()
}

func (n *Node) notify(n_prime *Node) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.predecessor == nil || (n_prime.id > n.predecessor.id && n_prime.id < n.id) {
		n.predecessor = n_prime
	}
}

func (n *Node) buildFingerTable() {
	for i := 0; i < m; i++ {
		start := (n.id + (1 << i)) % (1 << m)
		n.finger[i] = n.findSuccessor(start)
	}
}

func (n *Node) fixFingers() {
	next := 0
	start := (n.id + (1 << next)) % (1 << m)
	n.finger[next] = n.findSuccessor(start)
	next = (next + 1) % m
}

func main() {
	node1 := NewNode("addr1")
	node2 := NewNode("addr2")
	node3 := NewNode("addr3")

	node2.join(node1)
	node3.join(node1)

	node1.stabilize()
	node2.stabilize()
	node3.stabilize()

	fmt.Printf("Node1 successor: %d\n", node1.successor.id)
	fmt.Printf("Node2 successor: %d\n", node2.successor.id)

	// Generate visualization
	nodes := []*Node{node1, node2, node3}
	if err := GenerateChordVisualization(nodes); err != nil {
		fmt.Printf("Error generating visualization: %v\n", err)
	}
}
