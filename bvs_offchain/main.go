package main

import (
	"context"

	"github.com/satlayer/hello-world-bvs/bvs_offchain/node"
)

// main is the entry point of the program.
//
// It initializes a new node and runs it.
// No parameters.
// No return values.
// main is the entry point of the program.
//
// It initializes a new node and runs it.
// No parameters.
// No return values.
func main() {
	ctx := context.Background()
	n := node.NewNode()
	n.Run(ctx)
}
