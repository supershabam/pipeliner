package main

import (
	"fmt"

	"github.com/supershabam/pipeliner"
)

//go:generate pipeliner -operator=from -file=from_strings.go -func=fromStrings -from=string
func main() {
	ctx := pipeliner.FirstError()
	for i := range fromStrings(ctx, []string{"one", "two", "three"}) {
		fmt.Println(i)
	}
}
