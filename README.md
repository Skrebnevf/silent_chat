# Silent Chat Client

A simple and secure chat client written in Go. It connects to a server over TLS, authenticates users, and provides a terminal-based user interface for real-time messaging.

## Features

- **Secure Connection**: Uses TLS for encrypted communication.
- **Authentication**: Password-based authentication with SHA-256 hashing.
- **Terminal UI**: Built with Bubble Tea for a clean, interactive chat experience.
- **Privacy Features**: Sends fake messages periodically to enhance privacy.
- **Auto-Reconnect**: Automatically retries connections on failure.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Skrebnevf/silent_chat.git
   cd silent_chat
   ```

2. Build the application:
   ```bash
   go build ./cmd
   ```

## Usage

1. Set the server fingerprint environment variable for security:
   ```bash
   export CHAT_SERVER_FINGERPRINT=<server-fingerprint>
   ```

2. Run the client:
   ```bash
   ./silent_chat
   ```

3. Enter your username, password, server host, and port in the authentication screen.

The client will connect to the server, authenticate, and open the chat interface. Type messages and press Enter to send.

## Server

This client connects to the server available at: [https://github.com/Skrebnevf/silent_chat_server](https://github.com/Skrebnevf/silent_chat_server)

## Requirements

- Go 1.25.1 or later
- Terminal that supports ANSI escape codes

## License

This project is open source. See the server repository for more details.
