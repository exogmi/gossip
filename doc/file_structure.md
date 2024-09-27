Project Structure:

```
./
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── server/
│   │   └── server.go
│   ├── network/
│   │   ├── listener.go
│   │   └── client_session.go
│   ├── protocol/
│   │   ├── parser.go
│   │   └── protocol_handler.go
│   ├── state/
│   │   ├── state_manager.go
│   │   ├── user_manager.go
│   │   ├── channel_manager.go
│   │   └── message_store.go
│   └── models/
│       ├── user.go
│       ├── channel.go
│       └── message.go
├── config/
│   └── config.go
├── test/
│   ├── server_test.go
│   ├── network_test.go
│   ├── protocol_test.go
│   ├── state_test.go
│   └── models_test.go
└── go.mod
```
