package gemini

import (
	"google.golang.org/api/googleapi"
	"net/http"
	"testing"
)

func TestModelNotFoundError(t *testing.T) {
	gerr := &googleapi.Error{Code: http.StatusNotFound}
	if isNotAuthorizedError(gerr) {
		t.Fatal("expected googleapi.Error NOT to be a NotAuthorizedError")
	}
	if !isModelNotFoundError(gerr) {
		t.Fatal("expected googleapi.Error to be a ModelNotFoundError")
	}
}

func TestNotAuthorizedError(t *testing.T) {
	gerr := &googleapi.Error{
		Body: `[{
  "error": {
    "code": 400,
    "message": "API key not valid. Please pass a valid API key.",
    "status": "INVALID_ARGUMENT",
    "details": [
      {
        "@type": "type.googleapis.com/google.rpc.ErrorInfo",
        "reason": "API_KEY_INVALID",
        "domain": "googleapis.com",
        "metadata": {
          "service": "generativelanguage.googleapis.com"
        }
      },
      {
        "@type": "type.googleapis.com/google.rpc.LocalizedMessage",
        "locale": "en-US",
        "message": "API key not valid. Please pass a valid API key."
      }
    ]
  }
}]`,
	}
	if isModelNotFoundError(gerr) {
		t.Fatal("expected googleapi.Error NOT to be a ModelNotFoundError")
	}
	if !isNotAuthorizedError(gerr) {
		t.Fatal("expected googleapi.Error to be a NotAuthorizedError")
	}
}
