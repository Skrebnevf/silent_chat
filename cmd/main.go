package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"silent_chat/pkg/client"
	"silent_chat/pkg/config"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\033[2J\033[H")

	asciiArt := `
  ____ ___ _     _____ _   _ _____    ____ _   _    _  _____ 
 / ___|_ _| |   | ____| \ | |_   _|  / ___| | | |  / \|_   _|
 \___ \| || |   |  _| |  \| | | |   | |   | |_| | / _ \ | |  
  ___) | || |___| |___| |\  | | |   | |___|  _  |/ ___ \| |  
 |____/___|_____|_____|_| \_| |_|    \____|_| |_/_/   \_\_|  
                                                                                                                                           
`
	fmt.Println(asciiArt)

	fmt.Print("Type host: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)

	fmt.Print("Type port: ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)

	if host == "" || port == "" {
		fmt.Println(
			"Error: Host and port cannot be empty. Please provide valid values.",
		)
		os.Exit(1)
	}

	config := config.NewConfig()
	config.Addr = host + ":" + port
	config.ExpectedFP = os.Getenv("CHAT_SERVER_FINGERPRINT")

	if config.ExpectedFP == "" {
		fmt.Println("\nWARNING: CHAT_SERVER_FINGERPRINT is not set.")
		fmt.Println("Please set env and try again")
	} else {
		fmt.Printf("\nSecure mode enabled.\nExpected server fingerprint: %s\n\n", config.ExpectedFP)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	client := &client.Client{
		Reader:    reader,
		Addr:      config.Addr,
		Config:    config,
		Connected: false,
	}

	go func() {
		<-sigChan
		fmt.Println("\nCtrl+C pressed. Closing chat...")
		if client.Conn != nil {
			if errClose := client.Conn.Close(); errClose != nil {
				log.Printf("client close err: %v", errClose)
			}
		}
		os.Exit(0)
	}()

	if err := client.Run(); err != nil {
		fmt.Printf("Client terminated: %v\n", err)
		os.Exit(1)
	}
}
