package types

import (
	"fmt"
	interpolation "github.com/SadPencil/go-lagrange-interpolation"
	"github.com/SadPencil/go-lagrange-interpolation/field"
	gozkg "github.com/protolambda/go-kzg"
	"github.com/protolambda/go-kzg/bls"
	"math/big"
)

type Controller struct {
	Polynomial  []bls.Fr
	KzgSettings *gozkg.KZGSettings
}

func (c *Controller) Setup(nodesPrivateData []uint32) (ks *gozkg.KZGSettings, polynomial []bls.Fr, err error) {
	// modulus: subgroup size of bls12381 (order of bls.Fr)
	modulusHex := "0x73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001"
	modulus := new(big.Int)
	_, success := modulus.SetString(modulusHex, 0)
	if !success {
		return nil, nil, fmt.Errorf("failed to parse modulus string")
	}
	points := make([]*interpolation.XYPoint, 0)
	for i := 0; i < len(nodesPrivateData); i++ {
		points = append(points, &interpolation.XYPoint{
			X: &field.Field{Modulus: modulus, Value: big.NewInt(int64(nodesPrivateData[i]))},
			Y: &field.Field{Modulus: modulus, Value: big.NewInt(int64(i + 1))}}, // starts from 1
		)
	}
	interpolatingPolynomial, err := interpolation.LagrangeInterpolation(points)
	if err != nil {
		return nil, nil, err
	}

	// special case: abort if the degree is too small
	if interpolatingPolynomial.Degree() <= 1 {
		return nil, nil, fmt.Errorf("the polynomial degree is too low. Try using random private data")
	}

	polynomial = make([]bls.Fr, len(interpolatingPolynomial.Coefficients))
	for i := 0; i < len(interpolatingPolynomial.Coefficients); i++ {
		bls.SetFr(&polynomial[i], interpolatingPolynomial.Coefficients[i].Value.String())
	}

	// 16 = 2^4
	fs := gozkg.NewFFTSettings(4)
	// should be no less than 2^4+1 points
	// also, should be no less than polynomial degree + 1
	s1, s2 := gozkg.GenerateTestingSetup("1927409816240961209460912649124", max(uint64(len(polynomial)), 16+1))
	ks = gozkg.NewKZGSettings(fs, s1, s2)

	c.Polynomial = polynomial
	c.KzgSettings = ks
	return ks, polynomial, nil
}

func (c *Controller) Commit() *bls.G1Point {
	commitment := c.KzgSettings.CommitToPoly(c.Polynomial)
	return commitment
}

type PublicStorage struct {
	PolynomialCommitment *bls.G1Point
	KzgSettings          *gozkg.KZGSettings
}

type Node struct {
	Secret     uint64
	SecretFr   *bls.Fr
	Polynomial []bls.Fr
}
