package main

import(
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"strings"
)

var(
	clients = make(map[string]net.Conn) //Map to store all connected clients
	usernames = make(map[string]string) //Stores username for each client
	mutex = sync.Mutex{} //Synchronizes access to the clients map
	broadcast = make(chan string) // Channel for broadcasting all messages
	shutdown = make(chan os.Signal, 1) //Channel to catch OS signals
)

func main(){
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil{
		fmt.Println("Error starting server", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server started on 0.0.0.0:8080")

	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go handleBroadcast()

	go func(){
		<- shutdown
		fmt.Println("\nShutting down server....")

		mutex.Lock()
		for addr, conn := range clients{
			fmt.Println("Closing connection: ",addr)
			conn.Close()
		}
		mutex.Unlock()

		os.Exit(0)
	}()

	//Loop for continuously accepting new connections
	for{
		conn, err := listener.Accept()
		if err != nil{
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		fmt.Println("New client connected: ", conn.RemoteAddr())

		mutex.Lock()
		clients[conn.RemoteAddr().String()] = conn
		mutex.Unlock()

		go handleClient(conn) //Opens a new go thread (concurrency) to run a separate instance of the chat
	}
}

func handleClient(conn net.Conn){
	defer func(){
		mutex.Lock()
		username := usernames[conn.RemoteAddr().String()]
		delete(clients, conn.RemoteAddr().String())
		delete(usernames, conn.RemoteAddr().String())
		mutex.Unlock()
		conn.Close()
		fmt.Printf("%s has disconnected from the server.\n", username)
		broadcast <- fmt.Sprintf("** %s has left the chat **\n", username)
	}()

	reader := bufio.NewReader(conn)


	//Prompt for a username
	conn.Write([]byte("Enter your username: "))
	username, err := reader.ReadString('\n')
	if err != nil{
		return
	}
	username = strings.TrimSpace(username)

	//Store the client and its username
	mutex.Lock()
	clients[conn.RemoteAddr().String()] = conn
	usernames[conn.RemoteAddr().String()] = username
	mutex.Unlock()

	//Welcome message and broadcast
	conn.Write([]byte(fmt.Sprintf("\nWelcome, %s! Type /exit to leave the chat. \n", username)))
	broadcast <- fmt.Sprintf("\n** %s has joined the chat **\r\n", username)


	for{
		message, err := reader.ReadString('\n')
		if err != nil{
			return
		}
		message = strings.TrimSpace(message)

		//Check for special commands
		if message == "/exit"{
			conn.Write([]byte("See you next time!\n"))
			return
		}

		formattedMessage := fmt.Sprintf("[%s]: %s\r\n",username, message)
		fmt.Print(formattedMessage)

		broadcast <- formattedMessage //sends the message to the broadcast channel
	}
}


//function to listen for all incoming messages and send them to all clients
func handleBroadcast(){
	for{
		message := <- broadcast

		mutex.Lock()
		for _, conn := range clients{
			_, err := conn.Write([]byte(message))
			if err != nil{
				fmt.Println("Error sending message to the client: ", err)
			}
		}
		mutex.Unlock()
	}
}