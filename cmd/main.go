package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"silent_chat/pkg/client"
	"silent_chat/pkg/config"
)

const clearScreen = "\033[2J\033[H"

func main() {
	fmt.Print(clearScreen)

	config := config.NewConfig()
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
