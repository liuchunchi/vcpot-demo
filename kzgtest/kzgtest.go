package kzgtest

import (
	"example.com/kzg-demo/types"
	"example.com/kzg-demo/utils"
	"fmt"
	gozkg "github.com/protolambda/go-kzg"
	"github.com/protolambda/go-kzg/bls"
	"math/rand"
	"time"
)

func RunLargeNSkipInterpolation(n int) {
	var startTime time.Time
	var duration time.Duration

	nodesCount := max(n, 3)
	fmt.Printf("[n=%v]\n", nodesCount)

	polynomial := make([]bls.Fr, nodesCount+1)
	for i := 0; i < nodesCount+1; i++ {
		polynomial[i] = *bls.RandomFr()
	}
	// 2^10; seems to be irrelevant to performance??
	fs := gozkg.NewFFTSettings(10)
	// should be no less than 2^10+1 points
	// also, should be no less than polynomial degree + 1
	s1, s2 := gozkg.GenerateTestingSetup("1927409816240961209460912649124", max(uint64(len(polynomial)), 1024+1))
	ks := gozkg.NewKZGSettings(fs, s1, s2)

	public := types.PublicStorage{
		KzgSettings:          ks,
		PolynomialCommitment: ks.CommitToPoly(polynomial),
	}

	nodes := make([]types.Node, nodesCount)

	for i := 0; i < nodesCount; i++ {
		nodes[i].Polynomial = polynomial
	}

	for i := 0; i < nodesCount; i++ {
		data := rand.Int31()

		secretFr := new(bls.Fr)
		bls.AsFr(secretFr, uint64(data))
		nodes[i].SecretFr = secretFr
		nodes[i].Secret = uint64(data)
	}

	proofs := make([]*bls.G1Point, nodesCount)
	startTime = time.Now()
	for i := 0; i < nodesCount; i++ {
		if i%10 == 0 {
			fmt.Printf("[n=%v] prove (%v/%v)\n", nodesCount, i+1, nodesCount)
		}
		proofs[i] = public.KzgSettings.ComputeProofSingle(nodes[i].Polynomial, nodes[i].Secret)
	}
	duration = time.Since(startTime)

	fmt.Printf("[n=%v] prove %v\n", nodesCount, utils.DurationDivideBy(duration, nodesCount))

	ys := make([]*bls.Fr, nodesCount)
	for i := 0; i < nodesCount; i++ {
		ys[i] = new(bls.Fr)
		bls.AsFr(ys[i], uint64(i+1)) // incorrect answer
		//bls.EvalPolyAt(ys[i], nodes[i].Polynomial, nodes[i].SecretFr) // correct answer
	}

	startTime = time.Now()
	for i := 0; i < nodesCount; i++ {
		if i%10 == 0 {
			fmt.Printf("[n=%v] verify (%v/%v)\n", nodesCount, i+1, nodesCount)
		}
		public.KzgSettings.CheckProofSingle(public.PolynomialCommitment, proofs[i], nodes[i].SecretFr, ys[i])
	}
	duration = time.Since(startTime)

	fmt.Printf("[n=%v] verify %v\n", nodesCount, utils.DurationDivideBy(duration, nodesCount))

	fmt.Printf("done\n")
}

func Run(n int) {
	var startTime time.Time
	var duration time.Duration

	nodesCount := max(n, 3)
	fmt.Printf("[n=%v] begin setup\n", nodesCount)
	controller := types.Controller{}
	nodes := make([]types.Node, nodesCount)
	nodesPrivateData := make([]uint32, 0)

	for i := 0; i < nodesCount; i++ {
		data := rand.Int31()

		secretFr := new(bls.Fr)
		bls.AsFr(secretFr, uint64(data))
		nodes[i].SecretFr = secretFr
		nodes[i].Secret = uint64(data)
		nodesPrivateData = append(nodesPrivateData, uint32(data))
	}

	startTime = time.Now()
	_, _, err := controller.Setup(nodesPrivateData)
	if err != nil {
		panic("%v")
	}
	duration = time.Since(startTime)
	fmt.Printf("[n=%v] setup %v\n", nodesCount, duration)

	public := types.PublicStorage{
		KzgSettings:          controller.KzgSettings,
		PolynomialCommitment: controller.Commit(),
	}

	for i := 0; i < nodesCount; i++ {
		nodes[i].Polynomial = controller.Polynomial
	}

	proofs := make([]*bls.G1Point, nodesCount)
	startTime = time.Now()
	for i := 0; i < nodesCount; i++ {
		proofs[i] = public.KzgSettings.ComputeProofSingle(nodes[i].Polynomial, nodes[i].Secret)
	}
	duration = time.Since(startTime)

	fmt.Printf("[n=%v] prove %v\n", nodesCount, utils.DurationDivideBy(duration, nodesCount))

	startTime = time.Now()
	for i := 0; i < nodesCount; i++ {
		y := new(bls.Fr)
		bls.AsFr(y, uint64(i+1))
		public.KzgSettings.CheckProofSingle(public.PolynomialCommitment, proofs[i], nodes[i].SecretFr, y)
	}
	duration = time.Since(startTime)

	fmt.Printf("[n=%v] verify %v\n", nodesCount, utils.DurationDivideBy(duration, nodesCount))

	fmt.Printf("done\n")
}
