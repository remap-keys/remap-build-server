package web

import (
	"fmt"
	"net/http"
	"remap-keys.app/remap-build-server/common"
)

// ParseQueryParameters parses the query parameters. The query parameters are as follows:
//   - uid: The user's UID.
//   - taskId: The task ID.
func ParseQueryParameters(r *http.Request) (*common.RequestParameters, error) {
	queryParams := r.URL.Query()
	uid := queryParams.Get("uid")
	taskId := queryParams.Get("taskId")
	if uid == "" || taskId == "" {
		return nil, fmt.Errorf("uid or taskId is empty")
	}
	return &common.RequestParameters{
		Uid:    uid,
		TaskId: taskId,
	}, nil
}
