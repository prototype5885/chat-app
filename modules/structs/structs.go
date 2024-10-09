package structs

type Result struct {
	Success bool
	Message string
}

type ChatMessageResponse struct {
	IDm string // message ID
	IDu string // user ID
	Msg string // message
}

type ChannelResponse struct { // this is whats sent to the client when client requests channel
	ChannelID string
	Name      string
}

type ServerResponse struct {
	ServerID string
	OwnerID  string
	Name     string
	Picture  string
}

type ServerDeletionResponse struct {
	ServerID string
	UserID   string
}

type ServerMemberDeletionResponse struct {
	ServerID string
	UserID   string
}
