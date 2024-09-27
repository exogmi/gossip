# Gossip IRC Server

Gossip is a Golang-based IRC server designed to provide a persistent user experience, eliminating the need for external bouncers.

## Features

- **Client Connection Management:**
  - Accepts incoming TCP connections from IRC clients
  - Handles multiple clients per user transparently
  - Maintains user sessions even when clients disconnect
  - Delivers missed messages upon client reconnection

- **User Management:**
  - Tracks user presence and state (nickname, real name, last activity timestamp)
  - Associates multiple client connections with a single user
  - Persists user state across client disconnections

- **Channel Management:**
  - Supports creation and joining of channels
  - Manages user lists within channels
  - Stores and retrieves channel message history

- **Message Handling:**
  - Parses incoming IRC messages according to the IRC protocol
  - Handles standard IRC commands (e.g., NICK, USER, JOIN, PART, PRIVMSG, NOTICE)
  - Stores all messages with timestamps, regardless of user connection status
  - Delivers missed messages to reconnecting clients

- **SSL Support:**
  - Optional SSL/TLS encryption for client connections

## Usage

### Building the Server

To build the Gossip IRC server, run:

```bash
go build -o gossip_server cmd/server/main.go
```

### Running the Server

To start the server with default settings:

```bash
./gossip_server
```

### Command-line Options

- `-host`: Host to listen on (default: "localhost")
- `-port`: Port to listen on (default: 6667)
- `-ssl-port`: SSL Port to listen on (default: 6697)
- `-ssl-cert`: Path to SSL certificate file
- `-ssl-key`: Path to SSL key file
- `-use-ssl`: Enable SSL support
- `-verbosity`: Logging verbosity (info, debug, trace)

Example with SSL enabled:

```bash
./gossip_server -use-ssl -ssl-cert /path/to/cert.pem -ssl-key /path/to/key.pem -verbosity debug
```

## Connecting to the Server

You can connect to the Gossip IRC server using any standard IRC client. Point your client to the server's address and port (default: 6667, or 6697 for SSL).

## Future Enhancements

- Persistent storage for server state
- User authentication and registration
- Operator privileges
- Support for more advanced IRC features
- Performance optimization for handling a large number of concurrent users and channels

## Contributing

Contributions to the Gossip IRC Server are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## License

[MIT License](LICENSE)
