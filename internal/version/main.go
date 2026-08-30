// Command version prints the SDK version, so the release workflow can check it
// against the tag being pushed.
package main

import (
	"fmt"

	"github.com/fopost/fopost-go"
)

func main() {
	fmt.Print(fopost.Version)
}
