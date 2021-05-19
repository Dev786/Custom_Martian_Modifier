package body

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/martian/v3/log"
	"github.com/google/martian/v3/parse"
)

func init() {
	parse.Register("body.ResponseErrorModifier", modifierFromJSON)
}

// Modifier substitutes the body on an HTTP response.
type ResponseErrorModifier struct {
	contentType string
	body        []byte
	boundary    string
}

type ResponseErrorModifierJSON struct {
	ContentType string               `json:"contentType"`
	Body        []byte               `json:"body"` // Body is expected to be a Base64 encoded string.
	Scope       []parse.ModifierType `json:"scope"`
}

// NewModifier constructs and returns a body.Modifier.
func NewModifier(b []byte, contentType string) *ResponseErrorModifier {
	log.Debugf("body.NewModifier: len(b): %d, contentType %s", len(b), contentType)
	return &ResponseErrorModifier{
		contentType: contentType,
		body:        b,
		boundary:    randomBoundary(),
	}
}

// modifierFromJSON takes a JSON message as a byte slice and returns a
// body.Modifier and an error.
//
// Example JSON Configuration message:
// {
//   "scope": ["request", "response"],
//   "contentType": "text/plain",
//   "body": "c29tZSBkYXRhIHdpdGggACBhbmQg77u/" // Base64 encoded body
// }
func modifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &ResponseErrorModifierJSON{}
	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	mod := NewModifier(msg.Body, msg.ContentType)
	return parse.NewResult(mod, msg.Scope)
}

// SetBoundary set the boundary string used for multipart range responses.
func (m *ResponseErrorModifier) SetBoundary(boundary string) {
	m.boundary = boundary
}

// ModifyResponse sets the Content-Type header and overrides the response body.
func (m *ResponseErrorModifier) ModifyResponse(res *http.Response) error {
	log.Debugf("body.ModifyResponse: request: %s", res.Request.URL)
	// Replace the existing body, close it first.
	defer res.Body.Close()

	var recreatedResponse interface{}
	err := json.NewDecoder(res.Body).Decode(&recreatedResponse)
	if err != nil {
		fmt.Println("Error in decoding the response")
	} else {
		fmt.Printf("%s", recreatedResponse)
	}
	return nil
}

// randomBoundary generates a 30 character string for boundaries for mulipart range
// requests. This func panics if io.Readfull fails.
// Borrowed from: https://golang.org/src/mime/multipart/writer.go?#L73
func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}
