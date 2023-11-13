package web

import (
	"fmt"
	"net/http"
)

type RequestParameters struct {
	Uid    string
	TaskId string
}

// ParseQueryParameters parses the query parameters. The query parameters are as follows:
//   - uid: The user's UID.
//   - taskId: The task ID.
func ParseQueryParameters(r *http.Request) (*RequestParameters, error) {
	queryParams := r.URL.Query()
	uid := queryParams.Get("uid")
	taskId := queryParams.Get("taskId")
	if uid == "" || taskId == "" {
		return nil, fmt.Errorf("uid or taskId is empty")
	}
	return &RequestParameters{
		Uid:    uid,
		TaskId: taskId,
	}, nil
}
