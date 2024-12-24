package websocket

import (
	"encoding/json"
	"proto-chat/modules/database"
	log "proto-chat/modules/logging"
	"proto-chat/modules/macros"
)

func (c *Client) onUpdateUserDataRequest(packetJson []byte) {
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
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), "change user data", c.UserID)
	}

	response := UpdateUserDataRequest{
		NewDN: false,
		NewP:  false,
		NewST: false,
	}

	// if display name was changed
	if req.NewDN {
		log.Trace("Changing display name of user ID [%d] to [%s]", c.UserID, req.DisplayName)
		success := database.UpdateUserValue(c.UserID, req.DisplayName, "display_name")
		if !success {
			c.WriteChan <- macros.RespondFailureReason("Failed changing display name")
			return
		} else {
			type DisplayName struct {
				UserID      uint64
				DisplayName string
			}

			var newDisplayName = DisplayName{
				UserID:      c.UserID,
				DisplayName: req.DisplayName,
			}

			jsonBytes, err := json.Marshal(newDisplayName)
			if err != nil {
				macros.ErrorSerializing(err.Error(), "change member display name", c.UserID)
			}

			// get what servers are the user part of, so message will broadcast to members of these servers
			// this should make sure users who don't have visual on the user who changed display name won't get the message
			serverIDs := database.GetJoinedServersList(c.UserID)
			if len(serverIDs) != 0 {
				// if user is in servers
				broadcastChan <- BroadcastData{
					MessageBytes:    macros.PreparePacket(updateMemberDisplayName, jsonBytes),
					Type:            updateMemberDisplayName,
					AffectedServers: serverIDs,
				}
			}
			response.NewDN = true
			response.DisplayName = req.DisplayName
		}
	}
	// if pronouns were changed
	if req.NewP {
		log.Trace("Changing pronouns of user ID [%d] to [%s]", c.UserID, req.Pronouns)
		success := database.UpdateUserValue(c.UserID, req.Pronouns, "pronouns")
		if !success {
			c.WriteChan <- macros.RespondFailureReason("Failed changing pronouns")
		} else {
			response.NewP = true
			response.Pronouns = req.Pronouns
		}
	}
	// if status text was changed
	if req.NewST {
		// setUserStatusText(c.UserID, req.StatusText)
	}

	if req.NewDN || req.NewP || req.NewST {
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			macros.ErrorSerializing(err.Error(), "change user data response", c.UserID)
		}

		broadcastChan <- BroadcastData{
			MessageBytes:   macros.PreparePacket(updateUserData, jsonBytes),
			Type:           updateUserData,
			AffectedUserID: c.UserID,
		}
	}
}

func (c *Client) onUpdateUserStatusValue(packetJson []byte) {
	const jsonType string = "change status value"

	type UpdateUserStatusRequest struct {
		Status byte
	}

	var updateUserStatusRequest = UpdateUserStatusRequest{}

	if err := json.Unmarshal(packetJson, &updateUserStatusRequest); err != nil {
		macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
		c.WriteChan <- macros.ErrorDeserializing(err.Error(), jsonType, c.UserID)
	}
	// setUserStatus(c.UserID, updateUserStatusRequest.Status)
}
