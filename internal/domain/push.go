package domain

import (
	"crypto/hmac"
	"encoding/hex"
	"hash"
	"net/http"

	"source.toby3d.me/toby3d/hub/internal/common"
)

type Push struct {
	Self        Topic
	ContentType string
	Content     []byte
}

func (p Push) SetXHubSignatureHeader(req *http.Request, alg Algorithm, secret Secret) {
	if alg == AlgorithmUnd || secret.secret == "" {
		return
	}

	h := func() hash.Hash { return alg.Hash() }
	req.Header.Set(common.HeaderXHubSignature, alg.algorithm+"="+hex.EncodeToString(hmac.New(h,
		[]byte(secret.secret)).Sum(p.Content)))
}
