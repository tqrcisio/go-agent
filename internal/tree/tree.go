package tree

import "fmt"

// Node represents a node in the binary tree
type Node struct {
	Value string
	Left  *Node
	Right *Node
}

// Insert adds a new value to the binary tree
func Insert(root *Node, value string) *Node {
	if root == nil {
		return &Node{Value: value}
	}
	if value < root.Value {
		root.Left = Insert(root.Left, value)
	} else {
		root.Right = Insert(root.Right, value)
	}
	return root
}

// InOrder traverses the tree and prints the values
func InOrder(root *Node) {
	if root == nil {
		return
	}
	InOrder(root.Left)
	fmt.Printf("%s ", root.Value)
	InOrder(root.Right)
}
