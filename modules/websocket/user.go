package websocket

// func (c *Client) onUserInfoRequest(packetJson []byte, packetType byte) []byte {
// 	const jsonType string = "user info"

// 	type UserInfoRequest struct {
// 		UserID uint64
// 	}

// 	var userInfoRequest UserInfoRequest

// 	if err := json.Unmarshal(packetJson, &userInfoRequest); err != nil {
// 		macros.ErrorDeserializing(err.Error(), jsonType, c.userID)
// 	}

// 	var userInfoResponse = structs.UserInfo{
// 		UserID: strconv.FormatUint(userInfoRequest.UserID, 10),
// 	}

// 	userInfoResponse.Name, userInfoResponse.Picture = database.GetUserInfo(userInfoRequest.UserID)

// 	jsonBytes, err := json.Marshal(userInfoResponse)
// 	if err != nil {
// 		macros.ErrorSerializing(err.Error(), jsonType, c.userID)
// 	}

// 	return macros.PreparePacket(packetType, jsonBytes)
// }
