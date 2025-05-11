package main

import (
	"Xiaoxiaomeng-server/location"
	"Xiaoxiaomeng-server/openai"
	"Xiaoxiaomeng-server/weather"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Message struct {
	Data []byte
	From string
	To   string
}
type Connection struct {
	Name    string
	conn    *websocket.Conn
	lock    sync.Mutex
	asrConn *websocket.Conn
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
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "haoue")
	})
	router.HandleFunc("/ws", WsHandler)
	router.HandleFunc("/api/{command}", CommandHandler).Methods("GET")
	router.HandleFunc("/api/weather/get", GetWeatherHandler)
	router.HandleFunc("/api/location/get", GetLocationHandler)
	router.HandleFunc("/api/chat", ChatHandler).Methods("POST")

	if err := http.ListenAndServe(":3456", router); err != nil {
		log.Fatal(err)
	}
}
func AsrHandler(w http.ResponseWriter, r *http.Request) {

}
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	userName, pwd, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		return
	}
	if userName != "wusui" && pwd != "Qinsansui233..." {
		log.Println("用户名密码不对")
		w.WriteHeader(401)
		return
	}
	defer r.Body.Close()
	bts, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	var respMap map[string]any
	err = json.Unmarshal(bts, &respMap)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		return
	}
	key := respMap["key"]
	if key == nil {
		w.WriteHeader(400)
		return
	}

	content := respMap["content"]
	if content == nil {
		log.Println("content为空")
		w.WriteHeader(400)
		return
	}
	model := respMap["model"].(string)
	data, err := openai.Chat(content.(string), key.(string), model)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		return
	}
	bts, err = json.Marshal(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(401)
		return
	}
	w.WriteHeader(200)
	w.Write(bts)
}
func GetLocationHandler(w http.ResponseWriter, r *http.Request) {
	userName, pwd, ok := r.BasicAuth()
	if !ok {
		log.Println("认证失败")
		w.WriteHeader(401)
		return
	}
	if userName != "wusui" && pwd != "Qinsansui233..." {
		log.Println("用户名密码不对")
		w.WriteHeader(401)
		return
	}

	loc := r.URL.Query().Get("location")
	if loc == "" {
		w.WriteHeader(400)
		log.Println("缺少location")
		return
	}
	l, err := location.GetLocation(loc)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)

		return
	}
	w.WriteHeader(200)
	fmt.Fprintln(w, string(l))
}
func GetWeatherHandler(w http.ResponseWriter, r *http.Request) {
	userName, pwd, ok := r.BasicAuth()
	if !ok {
		log.Println("认证失败")
		w.WriteHeader(401)
		return
	}
	if userName != "wusui" && pwd != "Qinsansui233..." {
		log.Println("用户名密码不对")
		w.WriteHeader(401)
		return
	}

	location := r.URL.Query().Get("location")
	if location == "" {
		w.WriteHeader(400)
		log.Println("缺少location")
		return
	}
	w.WriteHeader(200)
	we, err := weather.GetNowWeather(location)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(200)
	fmt.Fprintln(w, string(we))

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
	log.Println(device)
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
	// asrClient, err := asr.Start()
	if err != nil {
		panic(err)
	}
	conn := &Connection{
		conn: ws,
		Name: device,
		lock: sync.Mutex{},
		// asrConn: asrClient,
	}
	connections.lock.Lock()
	connections.Devices[conn] = true
	connections.lock.Unlock()
	go conn.Receive(receiveChan)
	// go func(c *Connection) {
	// 	for {
	// 		msgType, msg, err := c.asrConn.ReadMessage()
	// 		if err != nil {
	// 			log.Println("从百度读取消息失败", err)
	// 			time.Sleep(100 * time.Millisecond)

	// 			c.conn.Close()
	// 			// c.asrConn.Close()
	// 			break
	// 		}
	// 		if msgType == websocket.TextMessage {
	// 			log.Println("百度语音转文本返回数据", string(msg))
	// 			m	 := map[string]any{}
	// 			err := json.Unmarshal(msg, &m)
	// 			if err != nil {
	// 				continue
	// 			}

	// 			if t, ok := m["type"].(string); ok && t == "FIN_TEXT" {
	// 				if result, ok := m["result"].(string); ok && result != "" {
	// 					log.Println("百度识别到的最终文本", result)
	// 					log.Println("发送到Deepseek", result)
	// 					go func(r string) {
	// 						data, err := openai.Chat(r, "", "")
	// 						if err != nil {
	// 							log.Println("发送到deepseek失败了", err)
	// 						}
	// 						if d, ok := data.(map[string]any); ok {
	// 							if content, ok := d["content"].(string); ok {
	// 								log.Println("deepseek返回的数据", content)
	// 								c.conn.WriteMessage(websocket.TextMessage, []byte(content))
	// 							}
	// 						}
	// 					}(result)
	// 				}
	// 			}

	// 		}
	// 	}
	// }(conn)
	for msg := range receiveChan {
		data := string(msg.Data)
		log.Println("收到消息", data)
		var isNoReply bool
		//如果字符串中包含isNoReply，则不发送给自己
		if strings.Contains(data, "_isNoReply") {
			isNoReply = true
		}
		for i := range connections.Devices {
			isNeedRemoveConnection := i.SendPing()
			if isNeedRemoveConnection {
				i.conn.Close()
				delete(connections.Devices, i)

				log.Println("连接已关闭", i.Name)
				continue
			}

			if msg.From == i.Name && isNoReply {
				log.Println("不发送给自己")
				continue
			}
			data = strings.ReplaceAll(data, "_isNoReply", "")

			msg.Data = []byte(data)
			i.Send(msg)
		}
	}
}
func (c *Connection) SendPing() bool {
	err := c.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
	if err != nil {
		log.Println(err)
		if c.conn != nil {
			_ = c.conn.Close()
			delete(connections.Devices, c)
		}
		log.Println("连接已关闭", c.Name)
		return true
	}
	return false
}
func (c *Connection) Send(message Message) {
	log.Println("发送消息到", c.Name, string(message.Data))
	err := c.conn.WriteMessage(websocket.TextMessage, message.Data)
	if err != nil {
		log.Println(err)
		if c.conn != nil {
			_ = c.conn.Close()
			delete(connections.Devices, c)
		}
		log.Println("连接已关闭", c.Name)
		return
	}
}
func (c *Connection) Receive(receiveChan chan<- Message) {

	for {
		messageType, i, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			c.conn.Close()
			// c.asrConn.Close()
			break
		}
		//二进制流，直接转发到百度语音识别接口
		if messageType == websocket.BinaryMessage {
			// err :x= c.asrConn.WriteMessage(messageType, i)

			// if err != nil {
			// 	log.Println("语音发送到百度失败", err)

			// 	continue
			// }
		}
		//文本类型，解析成json对象后检查json的类型
		if messageType == websocket.TextMessage {
			receiveChan <- Message{Data: i, From: c.Name, To: "any"}
		}
	}
}
