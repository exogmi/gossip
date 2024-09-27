## Gossip Server architecture

1. Core Components:
   - Server: Manages the overall server lifecycle.
   - ClientManager: Handles client connections and disconnections.
   - UserManager: Manages user states and associations with clients.
   - ChannelManager: Manages channels and user memberships.
   - MessageStore: Stores and retrieves messages with timestamps.

2. Network Layer:
   - TCPListener: Listens for incoming TCP connections.
   - ClientSession: Represents an individual client connection.

3. Protocol Layer:
   - ProtocolParser: Parses incoming IRC messages.
   - ProtocolHandler: Processes parsed commands and triggers appropriate actions.

4. State Management:
   - StateManager: Centralizes access to all state-related operations.

5. Interfaces:
   Define interfaces for each component to allow for easy mocking and testing.

Architecture Overview:

1. The Server component initializes and coordinates all other components.

2. TCPListener accepts new connections and creates ClientSession instances.

3. ClientManager maintains active ClientSessions and associates them with Users.

4. UserManager handles user-related operations and maintains user states.

5. ChannelManager manages channels and user memberships within channels.

6. MessageStore handles message persistence and retrieval.

7. ProtocolParser receives raw input from ClientSessions and converts it into structured commands.

8. ProtocolHandler processes parsed commands, interacting with other components as needed.

9. StateManager provides a unified interface for accessing and modifying server state.

Dependency Injection:

- Use a dependency injection container (e.g., dig or wire) to manage dependencies.
- Each component should depend on interfaces rather than concrete implementations.
- The Server component acts as the composition root, initializing and injecting dependencies.
