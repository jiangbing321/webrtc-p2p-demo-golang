package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	Join      string = "join"
	RespJoin         = "resp-join"
	Leave            = "leave"
	Offer            = "offer"
	Answer           = "answer"
	Candidate        = "candidate"
	Result           = "result"
)

type Response struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type ErrorMsg struct {
	Command string `json:"cmd"`
	Code    int    `json:"code"`
	Msg     string `json:"message`
}

type CommonMsg struct {
	Command string `json:"cmd"`
	RoomId  string `json:"roomId"`
}

type JoinMsg struct {
	CommonMsg
	UserName string `json:"userName"`
}

type BroadCastJoinMsg struct {
	JoinMsg
	UserId string `json:"userId"`
}

type RespJoinMsg struct {
	CommonMsg
	UserId string      `json:"userId"`
	Users  []*UserInfo `json:"users"`
}

type LeaveMsg struct {
	CommonMsg
	UserId   string `json:"uid"`
	UserName string `json:"userName"`
}

type OfferMsg struct {
	CommonMsg
	UserId string `json:"uid"`
	Msg    string `json:"msg"`
}

type AnswerMsg OfferMsg

type CandidateMsg OfferMsg

func parseMsg(ws *websocket.Conn, msgstr []byte) (err error) {
	msg := CommonMsg{}
	json.Unmarshal(msgstr, &msg)

	switch msg.Command {
	case Join:
		{
			msgreal := JoinMsg{}
			json.Unmarshal(msgstr, &msgreal)
			err = handleJoin(&msgreal, ws)
		}
	case Leave:
		{
			msgreal := LeaveMsg{}
			json.Unmarshal(msgstr, &msgreal)
			err = handleLeave(&msgreal, ws)
		}
	case Offer:
		{
			msgreal := OfferMsg{}
			json.Unmarshal(msgstr, &msgreal)
			err = handleOffer(&msgreal, ws)
		}
	case Candidate:
		{
			msgreal := CandidateMsg{}
			json.Unmarshal(msgstr, &msgreal)
			err = handleCandidate(&msgreal, ws)
		}
	default:
	}

	return err
}

func handleCandidate(candidateMsg *CandidateMsg, ws *websocket.Conn) error {
	broadCandidateInfo(candidateMsg, candidateMsg.RoomId, candidateMsg.UserId)
	return nil
}

func handleOffer(offerMsg *OfferMsg, ws *websocket.Conn) error {
	broadCastOffer(offerMsg, offerMsg.RoomId, offerMsg.UserId)
	return nil
}

func handleLeave(msg *LeaveMsg, ws *websocket.Conn) error {
	user := model.RemoveUser(msg.RoomId, msg.UserId, msg.UserName)
	if nil != user {
		errorMsg := ErrorMsg{
			Command: Result,
			Code:    0,
			Msg:     "success",
		}
		roomInfoMsgBody, _ := json.Marshal(errorMsg)
		sendMessage(ws, roomInfoMsgBody)
	} else {
		errorMsg := ErrorMsg{
			Command: Result,
			Code:    0,
			Msg:     "no such user...",
		}
		roomInfoMsgBody, _ := json.Marshal(errorMsg)
		sendMessage(ws, roomInfoMsgBody)
	}

	if nil != user {
		broadCastLeaveMsg(msg.RoomId, user.userId, user.UserName)
	}
	return nil
}

func handleJoin(msg *JoinMsg, ws *websocket.Conn) error {
	user, err := model.AddUser(msg.RoomId, msg.UserName, ws)
	if nil != err {
		errorMsg := ErrorMsg{
			Command: Result,
			Code:    304,
			Msg:     err.Error(),
		}
		errorMsgBody, _ := json.Marshal(errorMsg)
		sendMessage(ws, errorMsgBody)
	} else { // send current user info to the
		roomInfoMsg := RespJoinMsg{
			UserId: user.userId,
			Users:  model.GetUserListFromRoom(msg.RoomId),
			CommonMsg: CommonMsg{
				RoomId:  msg.RoomId,
				Command: RespJoin,
			},
		}
		roomInfoMsgBody, _ := json.Marshal(roomInfoMsg)
		sendMessage(ws, roomInfoMsgBody)
	}
	// broadcast msg

	if user != nil {
		broadCastJoinMsg(msg.RoomId, user.userId, user.UserName)
	}

	return nil
}

func broadCastOffer(offerMsg *OfferMsg, roomId string, userId string) {
	room := model.getRoom(roomId)
	for _, v := range room.users {
		if v.userId == userId {
			continue
		}
		msg := offerMsg
		msgJson, err := json.Marshal(msg)
		if nil == err && nil != v.ws {
			sendMessage(v.ws, msgJson)
		}
	}
}

func broadCandidateInfo(candidateMsg *CandidateMsg, roomId string, userId string) {
	room := model.getRoom(roomId)
	for _, v := range room.users {
		if v.userId == userId {
			continue
		}
		msg := candidateMsg
		msgJson, err := json.Marshal(msg)
		if nil == err && nil != v.ws {
			sendMessage(v.ws, msgJson)
		}
	}
}

func broadCastLeaveMsg(roomId string, leaveUserId string, leaveUserName string) {
	room := model.getRoom(roomId)
	for _, v := range room.users {
		if v.userId == leaveUserId {
			continue
		}
		msg := LeaveMsg{
			CommonMsg: CommonMsg{
				RoomId:  roomId,
				Command: Leave,
			},
			UserId:   leaveUserId,
			UserName: leaveUserName,
		}
		msgJson, err := json.Marshal(msg)
		if nil == err && nil != v.ws {
			sendMessage(v.ws, msgJson)
		}
	}
}

func broadCastJoinMsg(roomId string, joinUserId string, joinUserName string) {
	room := model.getRoom(roomId)
	for _, v := range room.users {
		if v.userId == joinUserId {
			continue
		}
		msg := BroadCastJoinMsg{
			JoinMsg: JoinMsg{
				CommonMsg: CommonMsg{
					RoomId:  roomId,
					Command: Join,
				},
				UserName: joinUserName,
			},
			UserId: joinUserId,
		}
		msgJson, err := json.Marshal(msg)
		if nil == err && nil != v.ws {
			sendMessage(v.ws, msgJson)
		}
	}
}

func sendMessage(ws *websocket.Conn, message []byte) {
	if ws == nil {
		return
	}
	fmt.Println("send Msg: ", string(message))
	ws.WriteMessage(websocket.TextMessage, message)
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Server(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println("connect error...", err)
		return
	}

	fmt.Println("get websocket connected...")
	ws.SetCloseHandler(func(code int, text string) error {
		fmt.Println("dissconnect code:", code, " text:", text)
		return nil
	})

	defer ws.Close()
	for {
		_, message, err := ws.ReadMessage()
		fmt.Println("Receive:", string(message))
		if err != nil {
			break
		}

		err = parseMsg(ws, message)

		if err != nil {
			fmt.Println("error:", err)
			break
		}
	}
}
