package gemini

import (
	"encoding/json"
	"github.com/BooleanCat/go-functional/v2/it/itx"
	"github.com/go-errors/errors"
	"google.golang.org/api/googleapi"
	"jcheng.org/jcllm/llm"
	"net/http"
)

type GoogleAPIErrorRoot struct {
	Error GoogleAPIError `json:"error"`
}

type GoogleAPIError struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Status  string           `json:"status"`
	Details []map[string]any `json:"details"`
}

func isModelNotFoundError(gerr *googleapi.Error) bool {
	return gerr.Code == http.StatusNotFound
}

func isNotAuthorizedError(gerr *googleapi.Error) bool {
	errList := unmarshalErrorBody(gerr)
	_, apiKeyInvalid := itx.FromSlice(errList).Find(func(elem GoogleAPIErrorRoot) bool {
		_, ok := itx.FromSlice(elem.Error.Details).Find(func(m map[string]any) bool {
			return m["reason"] == "API_KEY_INVALID"
		})
		return ok
	})
	return apiKeyInvalid
}

func unmarshalErrorBody(gerr *googleapi.Error) []GoogleAPIErrorRoot {
	var errList []GoogleAPIErrorRoot
	_ = json.Unmarshal([]byte(gerr.Body), &errList)
	return errList
}

func mapStreamReadError(err error) error {
	var gerr *googleapi.Error
	// If not a googleapi.Error, simply return it
	if !errors.As(err, &gerr) {
		return err
	}

	// Get details out of Google API errors
	if isModelNotFoundError(gerr) {
		return llm.ErrModelNotFound
	} else if isNotAuthorizedError(gerr) {
		return llm.ErrAPIKeyInvalid
	}
	errMsg := itx.FromSlice([]string{gerr.Message, gerr.Body, gerr.Error()}).Collect()[0]
	return errors.WrapPrefix(err, errMsg, 0)
}
