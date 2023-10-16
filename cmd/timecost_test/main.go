package main

import (
	"example.com/kzg-demo/kzgtest"
)

func main() {
	for _, n := range []int{3, 5, 10, 20, 50, 100, 200, 500, 1000} {
		kzgtest.Run(n)
	}
}
