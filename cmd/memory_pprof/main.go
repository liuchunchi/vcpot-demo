package main

import (
	"example.com/kzg-demo/kzgtest"
	"github.com/pkg/profile"
)

// https://geektutu.com/post/hpg-pprof.html
// analyze: go tool pprof -http=:9999 mem.pprof
// depends on graphviz

// sample -> inuse_space

// – inuse_space: Means pprof is showing the amount of memory allocated
// and not yet released.
// – inuse_objects: Means pprof is showing the amount of objects allocated
// and not yet released.
// – alloc_space: Means pprof is showing the amount of memory allocated,
// regardless if it was released or not.
// – alloc_objects: Means pprof is showing the amount of objects allocated,
// regardless if they were released or not.

func main() {
	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop()
	//kzgtest.Run(200)
	kzgtest.RunLargeNSkipInterpolation(1000)
}
