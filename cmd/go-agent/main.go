package main

import (
	"fmt"
	"go-agent/internal/tree"
)

func main() {
	var root *tree.Node
	values := []string{"M", "G", "S", "B", "J", "Q", "Z"}

	fmt.Println("Inserting values:", values)
	for _, v := range values {
		root = tree.Insert(root, v)
	}

	fmt.Print("In-order traversal (sorted result): ")
	tree.InOrder(root)
	fmt.Println()
}
