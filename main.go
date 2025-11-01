package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
)

const m = 3

func hashItem(item string, numNodes int) uint32 {
	digest := md5.Sum([]byte(item))
	digest1 := binary.BigEndian.Uint32(digest[0:4])
	digest2 := binary.BigEndian.Uint32(digest[4:8])
	digest3 := binary.BigEndian.Uint32(digest[8:12])
	digest4 := binary.BigEndian.Uint32(digest[12:16])
	return (digest1 ^ digest2 ^ digest3 ^ digest4) % (1 << numNodes)
}

type Node struct {
	id          uint32
	address     string
	successor   *Node
	predecessor *Node
	finger      []*Node
}

func NewNode(address string, numNodes int) *Node {
	node := &Node{
		id:      hashItem(address, numNodes),
		address: address,
		finger:  make([]*Node, m),
	}

	node.successor = node
	node.predecessor = node

	// finger table
	for i := 0; i < m; i++ {
		node.finger[i] = &Node{
			id: (node.id + (1 << i)) % (1 << m),
		}
	}

	return node
}

func (n *Node) findSuccessor(id uint32) *Node {
	return n.findPredecessor(id).successor
}

func (n *Node) findPredecessor(id uint32) *Node {
	for {
		if id >= n.id && id < n.successor.id {
			return n
		} else {
			return n.closestPrecedingFinger(id)
		}
	}

	return n
}

func (n *Node) closestPrecedingFinger(id uint32) *Node {
	for i := 0; i < m; i++ {
		if n.finger[i].id > n.id && n.finger[i].id < id {
			return n.finger[i]
		}
	}
	return n
}

// /////////////////////
type Ring struct {
	nodes      []*Node
	nodeHashes []uint32
}

func NewRing(nodes []string) *Ring {
	numNodes := len(nodes)
	ring := &Ring{}

	for i := 0; i < numNodes; i++ {
		ring.nodes = append(ring.nodes, NewNode(nodes[i], numNodes))
	}

	return ring
}

///////////////////////

func main() {
	nodes := []string{"lubna", "lubbu", "booboo"}
	ring := NewRing(nodes)
	for i := 0; i < len(nodes); i++ {
		fmt.Println("Node -> ", ring.nodes[i].id)
		fmt.Println("Succesor -> ", ring.nodes[i].findSuccessor(1).id)
		for _, finger := range ring.nodes[i].finger {
			fmt.Println("- ", finger.id)
		}
	}

	//if err := GenerateChordVisualization(ring); err != nil {
	//	fmt.Printf("Error generating visualization: %v\n", err)
	//}
}
