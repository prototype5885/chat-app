package structs

type Result struct {
	Success bool
	Message string
}

type ServerDeletionResponse struct {
	ServerID string
	UserID   string
}

type ServerMemberListResponse struct {
	UserID  string
	Name    string
	Picture string
	Status  string
}

type ServerMemberDeletionResponse struct {
	ServerID string
	UserID   string
}

// type UserInfo struct {
// 	UserID  string
// 	Name    string
// 	Picture string
// }
