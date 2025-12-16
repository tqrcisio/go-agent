package tree

import "fmt"

type Node struct {
	Value string
	Left  *Node
	Right *Node
}

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

func InOrder(root *Node) {
	if root == nil {
		return
	}
	InOrder(root.Left)
	fmt.Printf("%s ", root.Value)
	InOrder(root.Right)
}
