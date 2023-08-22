package websocket

import (
	"errors"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type UserInfo struct {
	UserName      string `json:"userName"`
	userId        string
	roomId        string
	SdpInfo       string `json:"sdp"`
	CandidateInfo string `json:"candidate"`
	ws            *websocket.Conn
}
type RoomInfo struct {
	roomId string
	users  map[string]*UserInfo
}

type webRtcServerContext struct {
	rooms map[string]*RoomInfo // room id==> RoomInfo
}

var model = &webRtcServerContext{
	rooms: make(map[string]*RoomInfo),
}

func (ctx *webRtcServerContext) init() {

}

func (ctx *webRtcServerContext) getUser(roomId string, userId string) *UserInfo {
	room := ctx.getRoom(roomId)
	if room != nil && room.users != nil {
		return room.users[userId]
	} else {
		return nil
	}
}

func (ctx *webRtcServerContext) getUserByName(roomId string, userName string) *UserInfo {
	room := ctx.getRoom(roomId)
	if room != nil && room.users != nil {
		for _, v := range room.users {
			if v.UserName == userName {
				return v
			}
		}
	}
	return nil
}

func (ctx *webRtcServerContext) getRoom(roomId string) *RoomInfo {
	return ctx.rooms[roomId]
}

func (ctx *webRtcServerContext) AddUser(roomId string, userName string, ws *websocket.Conn) (*UserInfo, error) {
	room := ctx.getRoom(roomId)
	if nil == room {
		room = ctx.CreateRoom(roomId)
	}

	userCount := len(room.users)
	if userCount >= 2 {
		return nil, errors.New("The room is full")
	}

	// check
	user := ctx.getUserByName(roomId, userName)
	if nil != user {
		return nil, errors.New("The user " + userName + " is alreay in room")
	}

	userId, _ := uuid.NewUUID()
	user = &UserInfo{
		userId:   userId.String(),
		roomId:   roomId,
		UserName: userName,
		ws:       ws,
	}
	room.users[userId.String()] = user

	return user, nil
}

func (ctx *webRtcServerContext) RemoveUser(roomId string, userId string, userName string) *UserInfo {
	room := ctx.getRoom(roomId)
	if nil == room {
		return nil
	}
	user := room.users[userId]
	if user == nil {
		user = ctx.getUserByName(roomId, userName)
	}
	if user != nil {
		delete(room.users, user.userId)
	}
	return user
}

func (ctx *webRtcServerContext) CreateRoom(roomId string) *RoomInfo {
	room := &RoomInfo{
		roomId: roomId,
		users:  make(map[string]*UserInfo),
	}

	ctx.rooms[roomId] = room
	return room
}

func (ctx *webRtcServerContext) RemoveRoom(roomId string) {
	delete(ctx.rooms, roomId)
}

func (ctx *webRtcServerContext) GetRoom(roomId string) *RoomInfo {
	return ctx.rooms[roomId]
}

func (ctx *webRtcServerContext) GetUserListFromRoom(roomId string) []*UserInfo {
	room := ctx.getRoom(roomId)
	userList := []*UserInfo{}
	if nil != room {
		for _, v := range room.users {
			userList = append(userList, v)
		}
	}
	return userList
}

func init() {
	model.init()
}
