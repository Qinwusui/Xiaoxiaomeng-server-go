package asr

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var appId float64
var appKey string
var devPid float64

func init() {
	bts, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var config map[string]any
	err = json.Unmarshal(bts, &config)
	if err != nil {
		panic(err)
	}
	appId = config["asr"].(map[string]any)["appId"].(float64)
	devPid = config["asr"].(map[string]any)["devPid"].(float64)
	appKey = config["asr"].(map[string]any)["appKey"].(string)
}

func Start() (conn *websocket.Conn, err error) {
	dialer := websocket.Dialer{
		HandshakeTimeout:  10 * time.Second,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
	}
	conn, _, err = dialer.Dial("wss://vop.baidu.com/realtime_asr?sn=21312", http.Header{})
	if err != nil {
		return
	}
	m := map[string]any{
		"type": "START",
		"data": map[string]any{
			"appid":   appId,
			"appkey":  appKey,
			"dev_pid": devPid,
			"cuid":    "qe2e",
			"format":  "pcm",
			"sample":  16000,
		},
	}
	bts, err := json.MarshalIndent(m, " ", "  ")
	log.Println(string(bts))
	if err != nil {
		return
	}
	conn.WriteMessage(websocket.TextMessage, bts)
	return
}

func IsConnected(conn *websocket.Conn) bool {
	// 发送 Ping 消息探测连接
	err := conn.WriteMessage(websocket.PingMessage, []byte{})
	if err != nil {
		return false
	}

	// 尝试读取（非阻塞模式）
	conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	if err != nil {
		return websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway)
	}
	conn.SetReadDeadline(time.Time{}) // 重置超时
	return true
}
