package websocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"proto-chat/modules/attachments"
	"proto-chat/modules/clients"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
	"proto-chat/modules/snowflake"
	"strconv"
	"time"
)

// when client is requesting to add a new channel, type 31
func (c *WsClient) onAddChannelRequest(packetJson []byte, packetType byte) {
	type AddChannelRequest struct {
		Name     string
		ServerID uint64
	}

	var channelRequest = AddChannelRequest{}

	if err := json.Unmarshal(packetJson, &channelRequest); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	var errorMessage = fmt.Sprintf("Error adding channel called [%s]", channelRequest.Name)

	// check if client is authorized to add channel to given server
	var ownerID uint64 = database.GetServerOwner(channelRequest.ServerID)
	if ownerID != c.UserID {
		log.Hack("User [%d] is trying to add a channel to server ID [%d] that they dont own", c.UserID, channelRequest.ServerID)
		c.WriteChan <- macros.RespondFailureReason("%s", errorMessage)
		return
	}

	var channelID uint64 = snowflake.Generate()

	// insert into database
	var channel = database.Channel{
		ChannelID: channelID,
		ServerID:  channelRequest.ServerID,
		Name:      channelRequest.Name,
	}

	err := database.Insert(channel)
	if err != nil {
		c.WriteChan <- macros.RespondFailureReason("%s", errorMessage)
		return
	}

	type ChannelResponse struct { // this is what's sent to the client when client requests channel
		ChannelID uint64
		Name      string
	}

	// serialize response about success
	var channelResponse = ChannelResponse{
		ChannelID: channelID,
		Name:      channelRequest.Name,
	}

	messagesBytes, err := json.Marshal(channelResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, messagesBytes),
		Type:            packetType,
		AffectedServers: []uint64{channelRequest.ServerID},
	}
}

func (c *WsClient) onChannelDeleteRequest(packetJson []byte, packetType byte) {
	type ChannelToDelete struct {
		ChannelID uint64
	}

	var req ChannelToDelete

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	serverID := database.GetServerIdOfChannel(req.ChannelID)

	ownerID := database.GetServerOwner(serverID)

	if ownerID != c.UserID {
		c.WriteChan <- macros.RespondFailureReason("Denied deleting channel ID [%d]", req.ChannelID)
		return
	}

	channelDeletion := database.ChannelDelete{
		ChannelID: req.ChannelID,
		ServerID:  serverID,
	}

	success := database.Delete(channelDeletion)
	if !success {
		c.WriteChan <- macros.RespondFailureReason("Error deleting channel ID [%d]", req.ChannelID)
		return
	}

	messagesBytes, err := json.Marshal(channelDeletion)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, messagesBytes),
		Type:            packetType,
		AffectedServers: []uint64{serverID},
	}
}

func (c *WsClient) onChannelDataUpdateRequest(packetJson []byte, packetType byte) {
	type UpdateChannelDataRequest struct {
		ChannelID uint64
		Name      string
		NewCN     bool
	}

	var req UpdateChannelDataRequest

	err := json.Unmarshal(packetJson, &req)
	if err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	serverID := database.GetServerIdOfChannel(req.ChannelID)
	ownerID := database.GetServerOwner(serverID)

	if ownerID != c.UserID {
		log.Hack("User ID [%d] tries to update data of channel ID [%d] without permission", c.UserID, req.ChannelID)
		c.WriteChan <- macros.RespondFailureReason("Denied updating data of channel ID [%d]", req.ChannelID)
		return
	}

	// update channel name
	if req.NewCN {
		success := database.ChangeChannelName(req.ChannelID, req.Name)
		if !success {
			log.Hack("Couldn't change name of channel ID [%d] to [%s] requested by user ID [%d]", req.ChannelID, req.Name, c.UserID)
			c.WriteChan <- macros.RespondFailureReason("Failed changing name of channel ID [%d]", req.ChannelID)
			return
		}

		jsonBytes, err := json.Marshal(req)
		if err != nil {
			macros.ErrorSerializing(err.Error(), packetType, c.UserID)
			return
		}

		broadcastChan <- BroadcastData{
			MessageBytes:    macros.PreparePacket(packetType, jsonBytes),
			Type:            packetType,
			AffectedServers: []uint64{serverID},
		}
	}
}

// when client requests list of server they are in, type 32
func (c *WsClient) onChannelListRequest(packetJson []byte, packetType byte) {
	type ChannelListRequest struct {
		ServerID uint64
	}

	var channelListRequest ChannelListRequest

	if err := json.Unmarshal(packetJson, &channelListRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	var serverID uint64 = channelListRequest.ServerID

	var isMember bool = database.ConfirmServerMembership(c.UserID, serverID)
	if isMember {
		success := clients.SetCurrentServerID(c.SessionID, serverID)
		if !success {
			log.Impossible("Failed setting current server ID to [%d] for user ID [%d] in onChannelListRequest", serverID, c.UserID)
			return
		}
		var jsonBytes []byte = database.GetChannelList(serverID)
		c.WriteChan <- macros.PreparePacket(packetType, jsonBytes)
	} else {
		c.WriteChan <- macros.RespondFailureReason("Rejected sending channel list of server ID [%d]", serverID)
	}
}

func (c *WsClient) onAddChatMessageRequest(packetJson []byte, packetType byte) {
	type ClientChatMsg struct {
		ChannelID uint64
		Message   string
		AttTok    string
	}

	var req ClientChatMsg

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	var rejectMessage = fmt.Sprintf("Denied sending chat message to channel ID [%d]", req.ChannelID)

	// check if user is member of the server which the channel belongs to
	var serverID uint64 = database.GetServerIdOfChannel(req.ChannelID)
	if serverID == 0 {
		c.WriteChan <- macros.RespondFailureReason("%s", rejectMessage)
		return
	}
	if !database.ConfirmServerMembership(c.UserID, serverID) {
		c.WriteChan <- macros.RespondFailureReason("%s", rejectMessage)
		return
	}

	attachmentToken, err := base64.StdEncoding.DecodeString(req.AttTok)
	if err != nil {
		log.Hack("User ID [%d] sent an attachmentToken base64 string that can't be decoded", c.UserID)
		c.WriteChan <- macros.RespondFailureReason("%s", rejectMessage)
		return
	}

	var uploadedAttachments []attachments.UploadedAttachment
	if len(attachmentToken) > 0 {
		uploadedAttachments = attachments.GetWaitingAttachment([64]byte(attachmentToken))
	}

	var messageID = snowflake.Generate()

	hasAttachments := false
	if req.AttTok != "" {
		hasAttachments = true
	}

	err = database.Insert(database.Message{
		MessageID:      messageID,
		ChannelID:      req.ChannelID,
		UserID:         c.UserID,
		Message:        req.Message,
		HasAttachments: hasAttachments,
	})
	if err != nil {
		log.FatalError(err.Error(), "Fatal error inserting message ID [%d] into database of user ID [%d]", messageID, c.UserID)
		return
	}

	log.Trace("Message ID [%d] will have [%d] attachmentList", messageID, len(uploadedAttachments))
	for i := 0; i < len(uploadedAttachments); i++ {
		attachment := database.Attachment{
			Hash:      uploadedAttachments[i].Hash[:],
			MessageID: messageID,
			Name:      uploadedAttachments[i].Name,
		}
		err := database.Insert(attachment)
		if err != nil {
			log.FatalError(err.Error(), "Fatal error inserting attachment of message ID [%d] into database of user ID [%d]", messageID, c.UserID)
			return
		}
	}

	var attachmentList []database.AttachmentResponse
	for i := 0; i < len(uploadedAttachments); i++ {
		attachmentResp := database.AttachmentResponse{
			Hash: uploadedAttachments[i].Hash[:],
			Name: uploadedAttachments[i].Name,
		}
		attachmentList = append(attachmentList, attachmentResp)
	}

	type ChatMessageResponse struct {
		MsgID  uint64
		UserID uint64
		Msg    string
		Att    []database.AttachmentResponse
	}

	var serverChatMsg = ChatMessageResponse{
		MsgID:  messageID,
		UserID: c.UserID,
		Msg:    req.Message,
		Att:    attachmentList,
	}

	jsonBytes, err := json.Marshal(serverChatMsg)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, jsonBytes),
		Type:            packetType,
		AffectedChannel: req.ChannelID,
	}
}

// when client is requesting chat history for a channel, type 2
func (c *WsClient) onChatHistoryRequest(packetJson []byte, packetType byte) {
	type ChatHistoryRequest struct {
		ChannelID     uint64
		FromMessageID uint64
		Older         bool
	}

	var req ChatHistoryRequest

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	success := clients.SetCurrentChannelID(c.SessionID, req.ChannelID)
	if !success {
		log.Impossible("Failed setting current channel ID to [%d] for user ID [%d] in onChatHistoryRequest", req.ChannelID, c.UserID)
		return
	}
	const rejectionMessage = "Denied chat history request"
	// check if user is member of server channel is part of
	serverID := database.GetServerIdOfChannel(req.ChannelID)
	if serverID == 0 {
		c.WriteChan <- macros.RespondFailureReason(rejectionMessage)
		return
	}
	if !database.ConfirmServerMembership(c.UserID, serverID) {
		c.WriteChan <- macros.RespondFailureReason(rejectionMessage)
		return
	}

	var jsonBytes []byte = database.GetChatHistory(req.ChannelID, req.FromMessageID, req.Older, c.UserID)
	if jsonBytes == nil {
		c.WriteChan <- macros.RespondFailureReason(rejectionMessage)
		return
	}

	c.WriteChan <- macros.PreparePacket(packetType, jsonBytes)
}

// when client wants to delete a message they own, type 3
func (c *WsClient) onChatMessageDeleteRequest(packetJson []byte, packetType byte) {
	type MessageToDelete struct {
		MessageID uint64
	}

	var req = MessageToDelete{}

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	// get the channel ID where the message was deleted,
	// so can broadcast it to affected Clients
	var channelID uint64 = database.DeleteChatMessage(req.MessageID, c.UserID)
	if channelID == 0 {
		c.WriteChan <- macros.RespondFailureReason("Denied to delete chat message")
		return
	}

	responseBytes, err := json.Marshal(req)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
	}

	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, responseBytes),
		Type:            packetType,
		AffectedChannel: channelID,
	}
}

func (c *WsClient) onAddFriendRequest(packetJson []byte, packetType byte) {
	type AddFriendRequest struct {
		UserID uint64
	}

	var req = AddFriendRequest{}

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	log.Trace("User ID [%d] wants to add [%d] as friend", c.UserID, req.UserID)

	// this is just extra check locally, database already doesn't allow 1 user being friends with itself
	if c.UserID == req.UserID {
		c.WriteChan <- macros.RespondFailureReason("You can't be friends with yourself")
		return
	}

	// TODO check if blocked

	// areFriends := database.CheckIfFriends(c.UserID, req.UserID)
	// if areFriends {
	// 	c.WriteChan <- macros.RespondFailureReason("You are already friends with user ID [%d]", req.UserID)
	// 	return
	// }

	friendship := database.Friendship{
		FriendsSince: time.Now().Unix(),
	}

	// make sure the smaller ID is first one
	if c.UserID < req.UserID {
		friendship.FirstUserID = c.UserID
		friendship.SecondUserID = req.UserID
	} else {
		friendship.FirstUserID = req.UserID
		friendship.SecondUserID = c.UserID
	}

	err := database.Insert(friendship)
	if err != nil {
		log.Warn("Error adding user ID [%d] as friend for [%d]", req.UserID, c.UserID)
		c.WriteChan <- macros.RespondFailureReason("Error adding user ID [%d] as friend", req.UserID)
		return
	}

	res := database.FriendshipSimple{
		UserID:     c.UserID,
		ReceiverID: req.UserID,
	}

	msgBytes, err := json.Marshal(res)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	broadcastData := BroadcastData{
		MessageBytes:   macros.PreparePacket(packetType, msgBytes),
		Type:           packetType,
		AffectedUserID: []uint64{c.UserID, req.UserID},
	}

	broadcastChan <- broadcastData
}

func (c *WsClient) onBlockUserRequest(packetJson []byte, packetType byte) {
	type BlockUserRequest struct {
		UserID uint64
	}

	var req = BlockUserRequest{}

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	log.Trace("[User %d] wants to block user [%d]", c.UserID, req.UserID)

	block := database.BlockUser{
		UserID:        c.UserID,
		BlockedUserID: req.UserID,
	}

	err := database.Insert(block)
	if err != nil {
		log.Warn("Error blocking user ID [%d] for [%d]", req.UserID, c.UserID)
		c.WriteChan <- macros.RespondFailureReason("Error blocking user ID [%d]", req.UserID)
		return
	}

	msgBytes, err := json.Marshal(req)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	broadcastChan <- BroadcastData{
		MessageBytes:   macros.PreparePacket(packetType, msgBytes),
		Type:           packetType,
		AffectedUserID: []uint64{c.UserID},
	}
}

func (c *WsClient) onUnfriendRequest(packetJson []byte) {
	type UnfriendRequest struct {
		UserID uint64
	}

	var req = UnfriendRequest{}

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), UNFRIEND, c.UserID)
		return
	}

	log.Trace("[User %d] wants to unfriend user [%d]", c.UserID, req.UserID)

	unfriend := database.FriendshipSimple{}

	// make sure the smaller ID is first one
	if c.UserID < req.UserID {
		unfriend.UserID = c.UserID
		unfriend.ReceiverID = req.UserID
	} else {
		unfriend.UserID = req.UserID
		unfriend.ReceiverID = c.UserID
	}

	success := database.Delete(unfriend)
	if !success {
		log.Warn("[User %d] failed to unfriend user [%d]", c.UserID, req.UserID)
		c.WriteChan <- macros.RespondFailureReason("Error unfriending user ID [%d]", req.UserID)
		return
	}

	res := database.FriendshipSimple{
		UserID:     c.UserID,
		ReceiverID: req.UserID,
	}

	msgBytes, err := json.Marshal(res)
	if err != nil {
		macros.ErrorSerializing(err.Error(), UNFRIEND, c.UserID)
		return
	}

	broadcastData := BroadcastData{
		MessageBytes:   macros.PreparePacket(UNFRIEND, msgBytes),
		Type:           UNFRIEND,
		AffectedUserID: []uint64{c.UserID, req.UserID},
	}

	broadcastChan <- broadcastData
}

// func (c *WsClient) onFriendListRequest(packetJson []byte) []byte {

// }

func (c *WsClient) onServerMemberListRequest(packetJson []byte, packetType byte) {
	type MemberListRequest struct {
		ServerID uint64
	}

	var req MemberListRequest

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	members := database.GetServerMembersList(req.ServerID)

	// check if members are online or not
	for i := 0; i < len(members); i++ {
		members[i].Online = clients.CheckIfUserIsOnline(members[i].UserID)
	}

	membersJson, err := json.Marshal(members)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	c.WriteChan <- macros.PreparePacket(packetType, membersJson)
}

func (c *WsClient) onLeaveServerRequest(packetJson []byte, packetType byte) {
	type LeaveServerRequest struct {
		ServerID uint64
	}

	var req LeaveServerRequest

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	var resp = database.ServerMemberShort{
		ServerID: req.ServerID,
		UserID:   c.UserID,
	}

	deleted := database.Delete(resp)
	if !deleted {
		c.WriteChan <- macros.RespondFailureReason("Failed leaving server ID [%d]", req.ServerID)
		return
	}

	responseBytes, err := json.Marshal(resp)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	// to make sure client won't receive messages after leaving
	success := clients.SetCurrentServerID(c.SessionID, 2000)
	if !success {
		return
	}
	success = clients.SetCurrentChannelID(c.SessionID, 0)
	if !success {
		return
	}

	packet := macros.PreparePacket(packetType, responseBytes)

	// have to send manually to deleter as they won't be part of broadcast list
	c.WriteChan <- packet

	broadcastChan <- BroadcastData{
		MessageBytes:    packet,
		AffectedServers: []uint64{req.ServerID},
		Type:            packetType,
	}
}

func (c *WsClient) onAddServerRequest(packetJson []byte) {
	type AddServerRequest struct {
		Name string
	}

	var addServerRequest = AddServerRequest{}

	if err := json.Unmarshal(packetJson, &addServerRequest); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), ADD_SERVER, c.UserID)
		return
	}

	const defaultPic = ""

	serverID := database.AddNewServer(c.UserID, addServerRequest.Name, defaultPic)

	var serverResponse = database.JoinedServer{
		ServerID: serverID,
		Owned:    true,
		Name:     addServerRequest.Name,
		Picture:  defaultPic,
	}

	messagesBytes, err := json.Marshal(serverResponse)
	if err != nil {
		macros.ErrorSerializing(err.Error(), ADD_SERVER, c.UserID)
		return
	}
	c.WriteChan <- macros.PreparePacket(ADD_SERVER, messagesBytes)
}

func (c *WsClient) onServerDeleteRequest(jsonBytes []byte, packetType byte) {
	type ServerToDelete struct {
		ServerID uint64
	}

	var req ServerToDelete

	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	serverDeletion := database.ServerDelete{
		ServerID: req.ServerID,
		UserID:   c.UserID,
	}

	success := database.Delete(serverDeletion)
	if !success {
		c.WriteChan <- macros.RespondFailureReason("Failed deleting server ID [%d]", req.ServerID)
		return
	}

	messagesBytes, err := json.Marshal(serverDeletion)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	members := database.GetServerMembersList(req.ServerID)
	onlineMembers := clients.FilterOnlineMembers(members)

	packet := macros.PreparePacket(packetType, messagesBytes)

	// have to send manually to deleter as they won't be part of broadcast list
	c.WriteChan <- packet

	broadcastChan <- BroadcastData{
		MessageBytes:   packet,
		Type:           packetType,
		AffectedUserID: onlineMembers,
	}
}

func (c *WsClient) onServerInviteRequest(packetJson []byte, packetType byte) {
	type ServerInviteRequest struct {
		ServerID     uint64
		TargetUserID uint64
		SingleUse    bool
		Expiration   uint64
	}

	var req = ServerInviteRequest{}

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	log.Trace("User ID [%d] is requesting to generate an invite link for server ID [%d]", c.UserID, req.ServerID)

	inviteID := snowflake.Generate()

	var serverInvite = database.ServerInvite{
		InviteID:     inviteID,
		ServerID:     req.ServerID,
		TargetUserID: req.TargetUserID,
		SingleUse:    req.SingleUse,
		Expiration:   req.Expiration,
	}

	err := database.Insert(serverInvite)
	if err != nil {
		log.Fatal("Error creating invite for server ID [%d] for user ID [%d]", req.ServerID, c.UserID)
		return
	}

	messagesBytes, err := json.Marshal(strconv.FormatUint(inviteID, 10))
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}
	c.WriteChan <- macros.PreparePacket(packetType, messagesBytes)
}

func (c *WsClient) onServerDataUpdateRequest(packetJson []byte, packetType byte) {
	type UpdateServerDataRequest struct {
		ServerID uint64
		Name     string
		NewSN    bool
	}

	var req UpdateServerDataRequest

	err := json.Unmarshal(packetJson, &req)
	if err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	// update server name
	if req.NewSN {
		success := database.ChangeServerName(c.UserID, req.ServerID, req.Name)
		if !success {
			log.Hack("Couldnt' change name of server ID [%d] to [%s] requested by user ID [%d], possibly because they are not the owner", req.ServerID, req.Name, c.UserID)
			c.WriteChan <- macros.RespondFailureReason("Failed changing name of server ID [%d]", req.ServerID)
			return
		}

		jsonBytes, err := json.Marshal(req)
		if err != nil {
			macros.ErrorSerializing(err.Error(), packetType, c.UserID)
			return
		}

		members := database.GetServerMembersList(req.ServerID)
		onlineMembers := clients.FilterOnlineMembers(members)

		broadcastChan <- BroadcastData{
			MessageBytes:   macros.PreparePacket(packetType, jsonBytes),
			Type:           packetType,
			AffectedUserID: onlineMembers,
		}
	}
}

func (c *WsClient) onInitialDataRequest(packetType byte) {
	initialData, success := database.GetInitialData(c.UserID)
	if !success {
		return
	}

	jsonUserID, err := json.Marshal(initialData)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	c.WriteChan <- macros.PreparePacket(packetType, jsonUserID)
}

func (c *WsClient) onImageHostAddressRequest(packetType byte) {
	log.Trace("User ID [%d] is requesting address of image host server", c.UserID)
	imageHostJson, err := json.Marshal(ImageHost)
	if err != nil {
		log.FatalError(err.Error(), "Error serializing ImageHost [%s]", ImageHost)
	}
	c.WriteChan <- macros.PreparePacket(packetType, imageHostJson)
}

func (c *WsClient) onUpdateUserDataRequest(packetJson []byte, packetType byte) {
	type UpdateUserDataRequest struct {
		DisplayName string
		Pronouns    string
		StatusText  string
		NewDN       bool
		NewP        bool
		NewST       bool
	}

	var req UpdateUserDataRequest

	err := json.Unmarshal(packetJson, &req)
	if err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	type UpdateUserDataResponse struct {
		UserID      uint64
		DisplayName string
		Pronouns    string
		StatusText  string
		NewDN       bool
		NewP        bool
		NewST       bool
	}

	response := UpdateUserDataResponse{
		UserID: c.UserID,
		NewDN:  false,
		NewP:   false,
		NewST:  false,
	}

	// if display name was changed
	if req.NewDN {
		log.Trace("Changing display name of user ID [%d] to [%s]", c.UserID, req.DisplayName)

		if len(req.DisplayName) > 32 {
			c.WriteChan <- macros.RespondFailureReason("Display name can't be longer than 32 bytes")
			return
		}

		success := database.UpdateUserValue(c.UserID, req.DisplayName, "display_name")
		if !success {
			c.WriteChan <- macros.RespondFailureReason("Failed changing display name")
			return
		} else {
			response.NewDN = true
			response.DisplayName = req.DisplayName
		}
	}
	// if pronouns were changed
	if req.NewP {
		log.Trace("Changing pronouns of user ID [%d] to [%s]", c.UserID, req.Pronouns)

		if len(req.Pronouns) > 16 {
			c.WriteChan <- macros.RespondFailureReason("Pronouns can't be longer than 16 bytes")
			return
		}

		success := database.UpdateUserValue(c.UserID, req.Pronouns, "pronouns")
		if !success {
			c.WriteChan <- macros.RespondFailureReason("Failed changing pronouns")
			return
		} else {
			response.NewP = true
			response.Pronouns = req.Pronouns
		}
	}
	// if status text was changed
	if req.NewST {
		log.Trace("Changing status text of user ID [%d] to [%s]", c.UserID, req.StatusText)

		if len(req.StatusText) > 32 {
			c.WriteChan <- macros.RespondFailureReason("Status text can't be longer than 32 bytes")
			return
		}

		success := setUserStatusText(c.UserID, req.StatusText)
		if !success {
			c.WriteChan <- macros.RespondFailureReason("Failed changing status text")
			return
		} else {
			response.NewST = true
			response.StatusText = req.StatusText
		}
	}

	if req.NewDN || req.NewP || req.NewST {
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			macros.ErrorSerializing(err.Error(), packetType, c.UserID)
			return
		}

		// broadcast it to every session of the user who changed their info
		broadcastChan <- BroadcastData{
			MessageBytes:   macros.PreparePacket(packetType, jsonBytes),
			Type:           packetType,
			AffectedUserID: []uint64{c.UserID},
		}

		// get what servers are the user part of, so message will broadcast to members of these servers
		// this should make sure users who don't have visual on the user who changed user data won't get the message
		serverIDs := database.GetJoinedServersList(c.UserID)
		// if user isn't in any servers, don't broadcast this
		if len(serverIDs) != 0 {
			broadcastChan <- BroadcastData{
				MessageBytes:    macros.PreparePacket(UPDATE_MEMBER_DATA, jsonBytes),
				Type:            UPDATE_MEMBER_DATA,
				AffectedServers: serverIDs,
			}
		}
	}
}

func (c *WsClient) onUpdateUserStatusValue(packetJson []byte, packetType byte) {
	type UpdateUserStatusRequest struct {
		Status byte
	}

	var req = UpdateUserStatusRequest{}

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}
	// change status in database
	success := database.UpdateUserValue(c.UserID, string(req.Status), "status")
	if !success {
		log.Warn("Failed to update user status value.")
		return
	}

	type NewStatus struct {
		UserID uint64
		Status byte
	}

	var newStatus = NewStatus{
		UserID: c.UserID,
		Status: req.Status,
	}

	jsonBytes, err := json.Marshal(newStatus)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	serverIDs := database.GetJoinedServersList(c.UserID)

	// prepare broadcast data that will be sent to affected users
	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, jsonBytes),
		Type:            packetType,
		AffectedServers: serverIDs,
	}
}

func (c *WsClient) onChatMessageTyping(packetJson []byte, packetType byte) {
	type StartedTyping struct {
		Typing bool
	}

	var req StartedTyping

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	channelID := clients.GetCurrentChannelID(c.SessionID)

	type Typing struct {
		Typing    bool
		UserID    uint64
		ChannelID uint64
	}

	resp := Typing{
		Typing:    req.Typing,
		UserID:    c.UserID,
		ChannelID: channelID,
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		macros.ErrorSerializing(err.Error(), packetType, c.UserID)
		return
	}

	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, jsonBytes),
		Type:            packetType,
		AffectedChannel: channelID,
	}
}

func (c *WsClient) onChatMessageEditRequest(packetJson []byte, packetType byte) {
	type EditedMessage struct {
		MessageID uint64
		Message   string
	}

	var req EditedMessage

	if err := json.Unmarshal(packetJson, &req); err != nil {
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), packetType, c.UserID)
		return
	}

	channelID := database.EditChatMessage(req.MessageID, c.UserID, req.Message)
	if channelID == 0 {
		log.Hack("Could not edit chat message ID [%d] requested by user ID [%d], possibly unauthorized", req.MessageID, c.UserID)
		return
	}

	broadcastChan <- BroadcastData{
		MessageBytes:    macros.PreparePacket(packetType, packetJson),
		Type:            packetType,
		AffectedChannel: channelID,
	}
}
