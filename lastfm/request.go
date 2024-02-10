package lastfm

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"slices"
)

type Request url.Values

const (
	paramMethod     = "method"
	paramApiKey     = "api_key"
	paramSessionKey = "sk"
	paramSignature  = "api_sig"
)

func newRequest(method string) Request {
	return map[string][]string{paramMethod: {method}}
}

func (r Request) method() string {
	return r[paramMethod][0]
}

func (r Request) set(k, v string) {
	r[k] = []string{v}
}

func (r Request) sign(apiKey, apiSecret, sessionKey string) {
	keys := make([]string, 0, len(r))

	// Add API Key and Session Key if defined
	r.set(paramApiKey, apiKey)
	if sessionKey != "" {
		r.set(paramSessionKey, sessionKey)
	}

	for k := range r {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	// Append keys and values in sorted order
	sig := md5.New()
	for _, k := range keys {
		sig.Write([]byte(k))
		sig.Write([]byte(r[k][0]))
	}
	// Append API Secret
	sig.Write([]byte(apiSecret))

	// Sign the request
	r.set(paramSignature, hex.EncodeToString(sig.Sum(nil)))
}

func (r Request) encode() string {
	return url.Values(r).Encode()
}
