package gemini

import (
	"encoding/json"
	"google.golang.org/api/googleapi"
	"net/http"
)

type (
	RemoteError struct {
		Err     *googleapi.Error
		errList []APIErrorContainer
	}

	APIErrorContainer struct {
		Errors APIError `json:"error"`
	}

	APIError struct {
		Code    int           `json:"code"`
		Message string        `json:"message"`
		Status  string        `json:"status"`
		Details []interface{} `json:"details"`
	}
)

func NewRemoteError() *RemoteError {
	return &RemoteError{}
}

func (remoteErr *RemoteError) Error() string {
	return remoteErr.Err.Error()
}

var _ error = (*RemoteError)(nil)

func (remoteErr *RemoteError) hasModelNotFoundError() bool {
	return remoteErr.Err != nil && remoteErr.Err.Code == http.StatusNotFound
}

func (remoteErr *RemoteError) unmarshalErrorBody() {
	_ = json.Unmarshal([]byte(remoteErr.Err.Body), &remoteErr.errList)
}
