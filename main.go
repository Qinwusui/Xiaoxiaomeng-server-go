package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Message struct {
	Data string
	From string
	To   string
}
type Connection struct {
	Name string
	conn *websocket.Conn
	lock sync.Mutex
}
type Connections struct {
	Devices map[*Connection]bool
	lock    sync.Mutex
}

var connections = Connections{
	Devices: make(map[*Connection]bool, 20),
	lock:    sync.Mutex{},
}

var Upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var receiveChan = make(chan Message, 30)

func main() {
	log.Println("请暴露3456端口")
	router := mux.NewRouter()
	router.HandleFunc("/ws", WsHandler)
	router.HandleFunc("/api/{command}", CommandHandler)
	go func() {
		for message := range receiveChan {
			connections.lock.Lock()
			rawMsg := strings.Replace(message.Data, "isNoReply:", "", -1)
			isNoReply := strings.Contains(message.Data, "isNoReply:")
			log.Println("收到", message.From, "发来的消息:", message.Data, "需要回复状态:", isNoReply)
			for i := range connections.Devices {
				if isNoReply {
					if i.Name != message.From {
						i.Send(Message{From: message.From, Data: rawMsg, To: i.Name})
					}
				} else {
					i.Send(Message{From: message.From, Data: rawMsg, To: i.Name})
				}

			}
			connections.lock.Unlock()
		}
	}()
	if err := http.ListenAndServe(":3456", router); err != nil {
		log.Fatal(err)
	}
}
func CommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Header)
	connections.lock.Lock()
	defer connections.lock.Unlock()
	for i := range connections.Devices {
		c := strings.Split(r.URL.Path, "/")
		cmd := c[len(c)-1]
		err := i.conn.WriteMessage(websocket.TextMessage, []byte(cmd))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
func WsHandler(w http.ResponseWriter, r *http.Request) {

	userName, pwd, ok := r.BasicAuth()
	device := r.PathValue("device")
	if device == "" {
		device = r.Header.Get("Device")
	}
	if device == "" {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("错误的设备")
		return
	}
	if userName != "wusui" || pwd != "Qinsansui233..." || !ok {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("错误的设备", userName, pwd)
		return
	}

	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(device, "已连接")
	conn := &Connection{
		conn: ws,
		Name: device,
		lock: sync.Mutex{},
	}
	go conn.Receive(receiveChan)

	connections.lock.Lock()
	if !connections.Devices[conn] {
		connections.Devices[conn] = true
	} else {
		_ = conn.conn.Close()
		delete(connections.Devices, conn)
	}
	connections.lock.Unlock()
}
func (c *Connection) Send(message Message) {

	err := c.conn.WriteMessage(websocket.TextMessage, []byte(message.Data))
	if err != nil {
		log.Println(err)
		return
	}
}
func (c *Connection) Receive(receiveChan chan<- Message) {
	for {
		messageType, i, err := c.conn.ReadMessage()
		if err != nil {
			receiveChan <- Message{Data: err.Error()}
			break
		}
		if messageType == websocket.TextMessage {
			var str = string(i)
			receiveChan <- Message{Data: str, From: c.Name, To: "any"}
		}
	}
}