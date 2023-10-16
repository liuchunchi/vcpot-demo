package main

import (
	"example.com/kzg-demo/kzgtest"
	"github.com/pkg/profile"
)

// https://geektutu.com/post/hpg-pprof.html
// analyze: go tool pprof -http=:9999 cpu.pprof
// depends on graphviz

func main() {
	defer profile.Start().Stop()
	kzgtest.Run(200)
	//kzgtest.RunLargeNSkipInterpolation(1000)
}
