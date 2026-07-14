package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"mime"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	// sha1Prefix is the prefix used by GitHub before the HMAC hexdigest.
	sha1Prefix = "sha1"
	// sha256Prefix and sha512Prefix are provided for future compatibility.
	sha256Prefix = "sha256"
	sha512Prefix = "sha512"
	// SHA256SignatureHeader is the GitHub header key used to pass the HMAC-SHA256 hexdigest.
	SHA256SignatureHeader = "X-Hub-Signature-256"
	// SHA1SignatureHeader is the GitHub header key used to pass the HMAC-SHA1 hexdigest.
	SHA1SignatureHeader = "X-Hub-Signature"
	// maxPayloadSize is the maximum size of a GitHub webhook payload.
	// GitHub documents a 25 MB limit for webhook payloads.
	maxPayloadSize = 25 * 1024 * 1024
)

// ValidatePayload validates the payload from the request using the GitHub API.
func ValidatePayload(r fiber.Ctx, secretToken []byte) (payload []byte, err error) {
	signature := r.Get(SHA256SignatureHeader)
	if signature == "" {
		signature = r.Get(SHA1SignatureHeader)
	}

	contentType, _, err := mime.ParseMediaType(r.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	return ValidatePayloadFromBody(contentType, r.Body(), signature, secretToken)
}

// ValidatePayloadFromBody validates an incoming GitHub Webhook event request body
// and returns the (JSON) payload.
// The Content-Type header of the payload can be "application/json" or "application/x-www-form-urlencoded".
// If the Content-Type is neither then an error is returned.
// secretToken is the GitHub Webhook secret token.
// If your webhook does not contain a secret token, you can pass an empty secretToken.
// Webhooks without a secret token are not secure and should be avoided.
//
// Example usage:
//
//	func (s *GitHubEventMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//		// read signature from request
//		signature := ""
//		payload, err := github.ValidatePayloadFromBody(r.Header.Get("Content-Type"), r.Body, signature, s.webhookSecretKey)
//		if err != nil { ... }
//		// Process payload...
//	}
func ValidatePayloadFromBody(contentType string, body []byte, signature string, secretToken []byte) (payload []byte, err error) {
	switch contentType {
	case "application/json":
		// If the content type is application/json,
		// the JSON payload is just the original body.
		payload = body

	case "application/x-www-form-urlencoded":
		// payloadFormParam is the name of the form parameter that the JSON payload
		// will be in if a webhook has its content type set to application/x-www-form-urlencoded.
		const payloadFormParam = "payload"

		// If the content type is application/x-www-form-urlencoded,
		// the JSON payload will be under the "payload" form param.
		form, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		payload = []byte(form.Get(payloadFormParam))

	default:
		return nil, fmt.Errorf("webhook request has unsupported Content-Type %q", contentType)
	}

	// Validate the signature if present or if one is expected (secretToken is non-empty).
	if len(secretToken) > 0 || len(signature) > 0 {
		if err := ValidateSignature(signature, body, secretToken); err != nil {
			return nil, err
		}
	}

	return payload, nil
}

// ValidateSignature validates the signature for the given payload.
// signature is the GitHub hash signature delivered in the X-Hub-Signature header.
// payload is the JSON payload sent by GitHub Webhooks.
// secretToken is the GitHub Webhook secret token.
//
// GitHub API docs: https://developer.github.com/webhooks/securing/#validating-payloads-from-github
func ValidateSignature(signature string, payload, secretToken []byte) error {
	messageMAC, hashFunc, err := messageMAC(signature)
	if err != nil {
		return err
	}
	if !checkMAC(payload, messageMAC, secretToken, hashFunc) {
		return errors.New("payload signature check failed")
	}
	return nil
}

// readPayloadBody reads the body from readable, enforcing maxPayloadSize.
func readPayloadBody(readable io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(readable, maxPayloadSize+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxPayloadSize {
		return nil, errors.New("webhook payload exceeds maximum allowed size")
	}
	return body, nil
}

// messageMAC returns the MAC method and the corresponding hash function.
func messageMAC(signature string) ([]byte, func() hash.Hash, error) {
	if signature == "" {
		return nil, nil, errors.New("missing signature")
	}
	sigParts := strings.SplitN(signature, "=", 2)
	if len(sigParts) != 2 {
		return nil, nil, fmt.Errorf("error parsing signature %q", signature)
	}

	var hashFunc func() hash.Hash
	switch sigParts[0] {
	case sha1Prefix:
		hashFunc = sha1.New
	case sha256Prefix:
		hashFunc = sha256.New
	case sha512Prefix:
		hashFunc = sha512.New
	default:
		return nil, nil, fmt.Errorf("unknown hash type prefix: %q", sigParts[0])
	}

	buf, err := hex.DecodeString(sigParts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding signature %q: %v", signature, err)
	}
	return buf, hashFunc, nil
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte, hashFunc func() hash.Hash) bool {
	expectedMAC := genMAC(message, key, hashFunc)
	return hmac.Equal(messageMAC, expectedMAC)
}

// genMAC generates the HMAC signature for a message provided the secret key
// and hashFunc.
func genMAC(message, key []byte, hashFunc func() hash.Hash) []byte {
	mac := hmac.New(hashFunc, key)
	mac.Write(message)
	return mac.Sum(nil)
}
