## Functional Description: Golang IRC Server

This document describes the functionality of a Golang-based IRC server designed to provide a persistent user experience, eliminating the need for external bouncers.

**Core Features:**

* **Client Connection Management:**
    * Accepts incoming TCP connections from IRC clients.
    * Handles multiple clients per user transparently.
    * Maintains user sessions even when clients disconnect.
    * Delivers missed messages upon client reconnection.

* **User Management:**
    * Tracks user presence and state (nickname, real name, last activity timestamp).
    * Associates multiple client connections with a single user.
    * Persists user state across client disconnections.

* **Channel Management:**
    * Supports creation and joining of channels.
    * Manages user lists within channels.
    * Stores and retrieves channel message history.

* **Message Handling:**
    * Parses incoming IRC messages according to the IRC protocol.
    * Handles standard IRC commands (e.g., NICK, USER, JOIN, PART, PRIVMSG, NOTICE).
    * Stores all messages with timestamps, regardless of user connection status.
    * Delivers missed messages to reconnecting clients.

* **Persistence (Future Enhancement):**
    * Stores server state (users, channels, messages) persistently (e.g., using a database).  Initially, this will be in-memory only.


**Non-Standard Features:**

* **Connection Resilience:** Users are not considered "offline" when a client disconnects. Their presence is maintained on the server.  All messages directed to them are stored and delivered upon reconnection.
* **Built-in Message Buffering:**  The server acts as a built-in bouncer, eliminating the need for external bouncing software. All messages are stored and replayed to users upon reconnection.


**Use Cases:**

* **Standard IRC Client Interaction:** Users can connect using any standard IRC client.
* **Reconnecting After Disconnection:** Users can reconnect after network interruptions or client crashes and seamlessly resume their session, receiving any missed messages.
* **Multiple Client Connections:**  Users can connect from multiple devices simultaneously, seeing the same messages and participating in channels across all connections.


**Future Enhancements:**

* **Persistent Storage:** Store server state (users, channels, messages) in a database for persistence across server restarts.
* **User Authentication:** Implement user authentication and registration.
* **Operator Privileges:** Implement operator commands and privileges for channel management.
* **Advanced IRC Features:** Support for more advanced IRC features like modes, server linking, and extensions.
* **SSL/TLS Support:** Secure client connections using SSL/TLS.
* **Performance Optimization:**  Optimize for handling a large number of concurrent users and channels.


This functional description outlines the core capabilities and intended behavior of the Golang IRC server. The focus is on providing a robust and resilient IRC experience with built-in message buffering and persistent user sessions.
