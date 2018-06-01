package main

import (
	"github.com/gorilla/websocket"
	"log"
	"encoding/json"
	"sync"
	// "fmt"
)

const (
	UserDisconnectMessage = 0
	PingMessage = 1
	UserConnectedMessage = 2
	GetClientsInRoomMessage = 3
	JoinPartyMessage = 4
	TapMessage = 5
	WinnerMessage = 6
	CreateUserMessage = 7
)


/* To figure out if they wanna broadcast to all or broadcast to all except them */
type Message struct {
	mtype string
	msg   []byte
}

//Message struct
type NetworkMessage struct {
	Username string `json:"username"`
	MessageType int `json:"messageType"`
	PositionX int `json:"positionX"`
	PositionY int `json:"positionY"`
	Message string `json:"message"`
	IPAddress string `json:"ipaddress"`
}

/* Reads and writes messages from client */
type Client struct {
	conn *websocket.Conn
	out  chan NetworkMessage
}

/* Reads and pumps to out channel */
func (c *Client) ReadLoop(wg *sync.WaitGroup) {
	defer close(c.out)
	defer wg.Done()
	for {
		wg.Add(1)
		var msg NetworkMessage
		// err := c.conn.ReadJSON(&msg)
		_, p, err := c.conn.ReadMessage()
		if p, ok := err.(*websocket.CloseError); ok {
			// log.Printf("readLoop() - ReadMessage() | err: ", err)
		    switch p.Code {
		    	case websocket.CloseNormalClosure:
		    		log.Println("Normal Closure")
		    		msg.IPAddress = c.conn.RemoteAddr().String()
		    		c.out <- msg
		    		break
		    	case websocket.CloseGoingAway:
		    		log.Println("Going Away Closure")
		    		msg.IPAddress = c.conn.RemoteAddr().String()
		    		c.out <- msg
		    		close(c.out)
		    		break
		    	case websocket.CloseNoStatusReceived:
		    		log.Println("No Status Received Closure")
		    		msg.IPAddress = c.conn.RemoteAddr().String()
		    		c.out <- msg
		    		close(c.out)
		    	default:
		    		// o := <-c.out
		    		log.Printf("err code: %d | %s", p.Code, c.conn.RemoteAddr().String())
		    		msg.IPAddress = c.conn.RemoteAddr().String()
		    		c.out <- msg
		    		close(c.out)
		    }
		    break
		    // return
		}
		// if err != nil {
		    
		// }
		if err := json.Unmarshal(p, &msg); err != nil {
		    log.Printf("readLoop() - Unmarshal() | err: ", err)
		    _, message, err := c.conn.ReadMessage()
				if err != nil {
					// log.Println("read:", err)
					break
				}
				log.Printf("Unknown Message and MesssageType: %s", message)
				// if string(message) == "ping" {
				// 	message = []byte("pong")
				// 	msg := NetworkMessage{"", "", "", PingMessage, string(message), c.conn.RemoteAddr().String()}
				// 	log.Printf("PING")
				// 	c.out <- msg
				// }
				break
		}

		switch msg.MessageType {
			case UserDisconnectMessage:
				log.Printf("ReadLoop | User Disconnect/Lost Connection to Server | (Message: %s)", msg.Message)
				msg.IPAddress = c.conn.RemoteAddr().String()
				break
			case PingMessage:
				log.Printf("ReadLoop | PingMessage: %s", msg.Message)
				msg.IPAddress = c.conn.RemoteAddr().String()
				break
			case UserConnectedMessage:
				log.Printf("ReadLoop | UserConnectedMessage -  (Message: %s)", msg.Message)
				msg.IPAddress = c.conn.RemoteAddr().String()
				break
			case GetClientsInRoomMessage:
				log.Printf("ReadLoop | GetClientsInRoomMessage: %s", msg.Message)
				msg.IPAddress = c.conn.RemoteAddr().String()
				break
				// tmp_r := r.GetClients()
				// log.Printf("%#v", tmp_r)
				// var s = ""
				// for k, v := range tmp_r { 
				    // s += fmt.Sprintf("%s : %s\n", k, v)
				// }
				// r.SendTo(id, out.Message)
			// 5 TapMessage
			case TapMessage:
				log.Printf("ReadLoop | TapMessage: %s", msg.Message)
				msg.IPAddress = c.conn.RemoteAddr().String()
				break
			case WinnerMessage:
				log.Printf("ReadLoop | WinnerMessage: %s", msg.Message)
				msg.IPAddress = c.conn.RemoteAddr().String()
				break
			case CreateUserMessage:
				log.Printf("ReadLoop | WinnerMessage: %s", msg.Message)
				msg.IPAddress = c.conn.RemoteAddr().String()
			default:
				log.Printf("ReadLoop | Unrecognized MessageType: %d", msg.MessageType)
		}
		c.out <- msg
	}
	wg.Wait()
}

// type NetworkMessage struct {
// 	Username string `json:"username"`
// 	MessageType int `json:"messageType"`
// 	PositionX int `json:"positionX"`
// 	PositionY int `json:"positionY"`
// 	Message string `json:"message"`
// 	IPAddress string `json:"ipaddress"`
// }

/* Writes a network message to the client */
// func (c *Client) WriteMessage(msg string) {
// 	log.Printf("writing message: %s to client.")
// 	network_msg := NetworkMessage{"", "", "", 100, msg, c.conn.RemoteAddr().String()}

// 	err := c.conn.WriteJSON(network_msg)
// 	if err != nil {
// 		log.Println("write:", err)
// 	}
// }

/*


/* Constructor */
func NewClient(conn *websocket.Conn) *Client {
	client := new(Client)
	client.conn = conn
	client.out = make(chan NetworkMessage)
	return client
}
