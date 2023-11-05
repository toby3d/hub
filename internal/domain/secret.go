package domain

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"math/rand"
	"net/url"
	"testing"

	"source.toby3d.me/toby3d/hub/internal/common"
)

// Secret describes a subscriber-provided cryptographically random unique secret
// string that will be used to compute an HMAC digest for authorized content
// distribution. If not supplied, the HMAC digest will not be present for
// content distribution requests. This parameter SHOULD only be specified when
// the request was made over HTTPS [RFC2818]. This parameter MUST be less than
// 200 bytes in length.
//
// [RFC2818]: https://tools.ietf.org/html/rfc2818
type Secret struct {
	secret string
}

var ErrSyntaxSecret = NewError("secret MUST be less than 200 bytes in length")

var lengthMax = 200

func ParseSecret(raw string) (*Secret, error) {
	if len(raw) >= lengthMax {
		return nil, ErrSyntaxSecret
	}

	return &Secret{secret: raw}, nil
}

func (s Secret) IsSet() bool {
	return s.secret != ""
}

func (s Secret) AddQuery(q url.Values) {
	if s.secret == "" {
		return
	}

	q.Add(common.HubSecret, s.secret)
}

func (s Secret) String() string {
	return s.secret
}

// TestSecret returns a valid random generated Secret.
func TestSecret(tb testing.TB) *Secret {
	tb.Helper()

	src := make([]byte, rand.Intn(lengthMax/2))
	if _, err := cryptorand.Read(src); err != nil {
		tb.Fatal(err)
	}

	return &Secret{secret: base64.URLEncoding.EncodeToString(src)}
}

// TestSecret returns a invalid random generated Secret.
func TestSecretInvalid(tb testing.TB) *Secret {
	tb.Helper()

	src := make([]byte, lengthMax*2)
	if _, err := cryptorand.Read(src); err != nil {
		tb.Fatal(err)
	}

	return &Secret{secret: base64.URLEncoding.EncodeToString(src)}
}
