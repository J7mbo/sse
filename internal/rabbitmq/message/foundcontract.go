package message

//
// FoundContract represents a contract found for a given user id. It will be sent over sever-sent events to update the
// user to the fact that this file is available to be tracked.
//
type FoundContract struct {
	UserID         string `json:"userid"`
	Filename       string `json:"filename"`
	Filepath       string `json:"filepath"`
	SearchFinished bool   `json:"finished"`
}
