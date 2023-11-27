package ethereum

import (
	"fmt"
)

var headNode = NewSymTreeNode("Head", "-1")
var count int
var nodeMap = make(map[string]*SymTreeNode)

type SymTreeNode struct {
	id string
	op string
	data []string
	children []*SymTreeNode
}

func GetHeadNode() *SymTreeNode {
	return headNode
}

func NewSymTreeNode(op string, id string) *SymTreeNode{
	return &SymTreeNode{
		id: id,
		op: op,
		data: make([]string, 0),
		children: make([]*SymTreeNode, 0),
	}
}

func BuildSymTreeWithNode(op string, id string, parentId string){
	// create a new node
	node := NewSymTreeNode(op, id)
	nodeMap[id] = node

	// find the parent node
	parentNode := nodeMap[parentId]

	// insert
	if parentNode == nil {
		parentNode = headNode
	}
	parentNode.children = append(parentNode.children, node)
}

func SymTreeLog(node *SymTreeNode){
	fmt.Println("+++++++", node.op)
	count++
	for i, child := range node.children {
		if len(node.children) > 1 {
			fmt.Println(node.op, "-", i)
		}
		SymTreeLog(child)
	}
}
func SymTreeCount(){
	fmt.Println("Number of nodes:", count)
}
