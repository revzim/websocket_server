package main

import (
	"log"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
	"math/rand"
)

type GameMap struct {
	sizeX int
	sizeY int
}


var GAMEMAP = GameMap{320, 320}

/* Has a name, clients, count which holds the actual coutn and index which acts as the unique id */
type Room struct {
	name    string
	clients map[int]*Client
	count   int
	index   int
}

/* Add a conn to clients map so that it can be managed */
func (r *Room) Join(conn *websocket.Conn) int {
	r.index++
	r.clients[r.index] = NewClient(conn)
	t := time.Now()
	log.Printf("\n%s - New Client %s joined room name: %s\n\n", t.Format("2006-01-02T15:04:05.999999-07:00"),r.clients[r.index].conn.RemoteAddr(), r.name)
	r.count++
	return r.index
}

/* Removes client from room */
func (r *Room) Leave(id int) {
	r.count--
	delete(r.clients, id)
}


func (r *Room) GetClients() map[int]*Client{
	return r.clients
}


/* Send to specific client */
func (r *Room) SendTo(id int, messageType int, posX int, posY int, msg string, ip string) {
	network_msg := NetworkMessage{"", messageType, posX, posY, msg, ip}

	log.Printf("Sending %d message to Client[%d]", messageType, id)
	err := r.clients[id].conn.WriteJSON(network_msg)
	if err != nil {
		log.Println("BroadcastAllExcept write error: ", err)
	}
}

/* Broadcast to every client */
func (r *Room) BroadcastAll(id int, messageType int, posX int, posY int, message string, ip string) {
	for _, client := range r.clients {
		log.Printf("BroadcastAll prepping... | Client[%d] - mtype: %d | msg: %s", id, messageType, message)

		out := <- r.clients[id].out
		username := out.Username
		ip := out.IPAddress
		network_msg := NetworkMessage{username, messageType, posX, posY, message, ip}

		err := client.conn.WriteJSON(network_msg)
		if err != nil {
			log.Println("BroadcastAll write error: ", err)
		}
	}
}

/* Broadcast to all except use for chat */
func (r *Room) BroadcastAllExcept(senderid int, messageType int, posX int, posY int, message string, ip string) {
	if r.count > 0 {

		for id, client := range r.clients {
			if id != senderid {
				log.Printf("BroadcastAllExcept prepping...\nClient[%d] IP: %s - messageType: %d | msg: %s", id, ip, messageType, message)

				network_msg := NetworkMessage{"", messageType, posX, posY, message, ip}

				err := client.conn.WriteJSON(network_msg)
				if err != nil {
					log.Println("BroadcastAllExcept write error: ", err)
					r.Leave(id)
				}
			}
		}
	}
	
}

/* Handle messages */
/*
UserDisconnectMessage = 0
PingMessage = 1
UserConnectedMessage = 2
GetClientsInRoomMessage = 3
JoinPartyMessage = 4
TapMessage = 5
*/
func (r *Room) HandleMsg(id int, wg *sync.WaitGroup) {

	defer wg.Done()
	for {
		if r.clients[id] == nil {
			break
		}
		// grab channel
		// add wait group
		wg.Add(1)
		out := <-r.clients[id].out
		log.Printf("HandleMsg - Out: %s", out)

		switch out.MessageType {
			case 0:
				wg.Add(1)
				log.Printf("User Disconnect/Lost Connection to Server | (Message: %s)", out.Message)
				r.Leave(id)
				r.BroadcastAllExcept(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			case PingMessage:
				wg.Add(1)
				log.Printf("PingMessage - client[%d] sent ping. Prepping Pong message...", id)
				r.SendTo(id, out.MessageType, out.PositionX, out.PositionY, "pong", out.IPAddress)
			case UserConnectedMessage:
				wg.Add(1)
				log.Printf("UserConnectedMessage - client[%d] connected | (Message: %s)", id, out.Message)
				//tell everyone else connected that a new user connected
				r.BroadcastAllExcept(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			case GetClientsInRoomMessage:
				wg.Add(1)
				log.Printf("GetClientsInRoomMessage - client[%d] requested users. | (Message: %s)", id, out.Message)
				/*
				 * special count condition for 1v1
				 */
				if r.count == 2 {
					tmp_r := r.GetClients()
					var s = ""
					for index, element := range tmp_r {
						if element.conn.RemoteAddr().String() != out.IPAddress {
							log.Printf("%s - Client[%d] is your opponent.", element.conn.RemoteAddr().String(), index)
							s = element.conn.RemoteAddr().String()
						}
						
						// s += fmt.Sprintf("[%s]", element.conn.RemoteAddr().String())
					}
					out.Message = s
					r.SendTo(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
				}
				// if r.count > 1 {
				// 	tmp_r := r.GetClients()
				// 	var s = ""
				// 	for index, element := range tmp_r {
				// 		log.Printf("Client[%d]: %#v", index, element.conn.RemoteAddr().String())
				// 		s += fmt.Sprintf("[%s]", element.conn.RemoteAddr().String())
				// 	}
				// 	out.Message = s
				// 	r.SendTo(id, out.MessageType, out.Message)
				// }

			case TapMessage:
				wg.Add(1)
				log.Printf("TapMessage - client[%d] connected | (Message: %s)", id, out.Message)
				//tell everyone else connected that a new user connected
				r.BroadcastAllExcept(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			case WinnerMessage:
				wg.Add(1)
				log.Printf("WinnerMessage - client[%d] won! | (Message: %s)", id, out.Message)
				r.BroadcastAll(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			default:
				log.Printf("Unrecognized MessageType: %d", out.MessageType)
		}

	}

	//after done with channel, allivate waitgroup
	wg.Wait()
}


/* Handle messages */
/*
UserDisconnectMessage = 0
PingMessage = 1
UserConnectedMessage = 2
GetClientsInRoomMessage = 3
JoinPartyMessage = 4
TapMessage = 5
*/
func (r *Room) HandleMainLobbyMsg(id int, wg *sync.WaitGroup) {

	defer wg.Done()
	for {
		if r.clients[id] == nil {
			break
		}
		// grab channel
		// add wait group
		wg.Add(1)
		out := <-r.clients[id].out
		// log.Printf("Mainlobby - HandleMsg - Out: %s\n", out)
		
		switch out.MessageType {
			case 0:
				wg.Add(1)
				log.Printf("Mainlobby - User Disconnect/Lost Connection to Server | (IP: %s | User: %s disconnected.)\n", out.IPAddress, out.Username)
				r.Leave(id)
				r.BroadcastAllExcept(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			case PingMessage:
				wg.Add(1)
				log.Printf("Mainlobby - PingMessage - client[%d] sent ping. Prepping Pong message...\n", id)
				r.SendTo(id, out.MessageType, out.PositionX, out.PositionY, "pong", out.IPAddress)
			case UserConnectedMessage:
				wg.Add(1)
				log.Printf("Mainlobby - UserConnectedMessage - client[%d] connected | (Message: %s)\n", id, out.Message)
				//tell everyone else connected that a new user connected
				r.BroadcastAllExcept(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			case GetClientsInRoomMessage:
				wg.Add(1)
				log.Printf("Mainlobby - GetClientsInRoomMessage - client[%d] requested users. | (Message: %s)\n", id, out.Message)
				/*
				 * special count condition for 1v1
				 */
				if r.count >= 1 {
					// r.GetLobbyClients(id, out.MessageType, out.Message)
					tmp_r := r.GetClients()
					var s = ""
					for index, element := range tmp_r {
						if element.conn.RemoteAddr().String() != out.IPAddress {
							log.Printf("Mainlobby - %s - Client[%d] is connected.\n", element.conn.RemoteAddr().String(), index)
							s += fmt.Sprintf("%s+", element.conn.RemoteAddr().String())
						}
						
						// s += fmt.Sprintf("[%s]", element.conn.RemoteAddr().String())
					}
					out.Message = s
					r.SendTo(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
				}
				// if r.count > 1 {
				// 	tmp_r := r.GetClients()
				// 	var s = ""
				// 	for index, element := range tmp_r {
				// 		log.Printf("Client[%d]: %#v", index, element.conn.RemoteAddr().String())
				// 		s += fmt.Sprintf("[%s]", element.conn.RemoteAddr().String())
				// 	}
				// 	out.Message = s
				// 	r.SendTo(id, out.MessageType, out.Message)
				// }

			case TapMessage:
				wg.Add(1)
				log.Printf("Mainlobby - TapMessage - client[%d] %s tapped| (Message: %s\n)", id, out.Username,out.Message)
				//tell everyone else connected that a new user connected
				r.BroadcastAllExcept(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			case WinnerMessage:
				wg.Add(1)
				log.Printf("Mainlobby - WinnerMessage - client[%d] won! | (Message: %s\n)", id, out.Message)
				r.BroadcastAll(id, out.MessageType, out.PositionX, out.PositionY, out.Message, out.IPAddress)
			case CreateUserMessage:
				wg.Add(1)
				log.Printf("Mainlobby - CreateUserMessage - client[%d] created | (Message: %s\n)", id, out.Message)
				var posX = 0
				var posY = 0
				rng := rand.New(rand.NewSource(time.Now().UnixNano()))
				posX = rng.Intn(GAMEMAP.sizeX - (-GAMEMAP.sizeX)) + (-GAMEMAP.sizeX)
				posY = rng.Intn(GAMEMAP.sizeX - (-GAMEMAP.sizeX)) + (-GAMEMAP.sizeX)
				log.Printf("Position: %d, %d", posX, posY)
				out.Username = out.Message
				r.BroadcastAll(id, out.MessageType, posX, posY, out.Message, out.IPAddress)
			default:
				log.Printf("Mainlobby - Unrecognized MessageType: %d\n", out.MessageType)
				break
		}

	}

	//after done with channel, allivate waitgroup
	wg.Wait()
}

// func GetLobbyClients() {
// 	tmp_r := r.GetClients()
// 	var s = ""
// 	for index, element := range tmp_r {
// 		if element.conn.RemoteAddr().String() != out.IPAddress {
// 			log.Printf("Mainlobby - %s - Client[%d] is connected.\n", element.conn.RemoteAddr().String(), index)
// 			s = element.conn.RemoteAddr().String()
// 		}
		
// 		// s += fmt.Sprintf("[%s]", element.conn.RemoteAddr().String())
// 	}
// }

/* Constructor */
func NewRoom(name string) *Room {
	room := new(Room)
	room.name = name
	room.clients = make(map[int]*Client)
	room.count = 0
	room.index = 0
	return room
}
