package websocket

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestModel(t *testing.T) {
	msg := CommonMsg{
		Command: "join",
		RoomId:  "32434",
	}
	msgstr, _ := json.Marshal(msg)
	fmt.Println("result:", string(msgstr))
}
