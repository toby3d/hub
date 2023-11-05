package domain

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"

	"source.toby3d.me/toby3d/hub/internal/common"
)

type Algorithm struct {
	algorithm string
}

var (
	AlgorithmUnd    = Algorithm{algorithm: ""}       // "und"
	AlgorithmSHA1   = Algorithm{algorithm: "sha1"}   // "sha1"
	AlgorithmSHA256 = Algorithm{algorithm: "sha256"} // "sha256"
	AlgorithmSHA384 = Algorithm{algorithm: "sha384"} // "sha384"
	AlgorithmSHA512 = Algorithm{algorithm: "sha512"} // "sha512"
)

var ErrSyntaxAlgorithm = errors.New("bad algorithm syntax")

var stringsAlgorithms = map[string]Algorithm{
	AlgorithmSHA1.algorithm:   AlgorithmSHA1,
	AlgorithmSHA256.algorithm: AlgorithmSHA256,
	AlgorithmSHA384.algorithm: AlgorithmSHA384,
	AlgorithmSHA512.algorithm: AlgorithmSHA512,
}

func ParseAlgorithm(algorithm string) (Algorithm, error) {
	if alg, ok := stringsAlgorithms[algorithm]; ok {
		return alg, nil
	}

	return AlgorithmUnd, fmt.Errorf("%w: %s", ErrSyntaxAlgorithm, algorithm)
}

func (a Algorithm) Hash() hash.Hash {
	switch a {
	default:
		return nil
	case AlgorithmSHA1:
		return sha1.New()
	case AlgorithmSHA256:
		return sha256.New()
	case AlgorithmSHA384:
		return sha512.New384()
	case AlgorithmSHA512:
		return sha512.New()
	}
}

func (a *Algorithm) UnmarshalForm(src []byte) error {
	var err error
	if *a, err = ParseAlgorithm(string(src)); err != nil {
		return fmt.Errorf("Algorithm: %w", err)
	}

	return nil
}

func (a Algorithm) String() string {
	if a.algorithm != "" {
		return a.algorithm
	}

	return common.Und
}

func (a Algorithm) GoString() string {
	return "domain.Algorithm(" + a.String() + ")"
}
