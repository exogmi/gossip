import unittest
import subprocess
import time
import irc.client

class TestNicknameChange(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        # Start the IRC server
        cls.server_process = subprocess.Popen(["go", "run", "cmd/server/main.go"])
        time.sleep(2)  # Give the server some time to start

    @classmethod
    def tearDownClass(cls):
        # Stop the IRC server
        cls.server_process.terminate()
        cls.server_process.wait()

    def setUp(self):
        self.client1 = irc.client.Reactor()
        self.client2 = irc.client.Reactor()
        self.received_messages = []
        self.server1 = self.client1.server()
        self.server2 = self.client2.server()
        self.connection1 = self.server1.connect("localhost", 6667, "user1")
        self.connection2 = self.server2.connect("localhost", 6667, "user2")
        self.connection1.add_global_handler("all_events", self.on_event)
        self.connection2.add_global_handler("all_events", self.on_event)

    def tearDown(self):
        self.connection1.disconnect()
        self.connection2.disconnect()

    def on_event(self, connection, event):
        self.received_messages.append((event.type, event.source, event.target, event.arguments))

    def test_nickname_change(self):
        channel = "#testchannel"
        new_nickname = "user1_new"

        # Join the channel
        self.connection1.join(channel)
        self.connection2.join(channel)

        # Give some time for the joins to complete
        time.sleep(1)

        # Change nickname for user1
        self.connection1.nick(new_nickname)

        # Wait for the nickname change to be processed
        start_time = time.time()
        nick_change_received = False
        self_nick_change_received = False
        
        while time.time() - start_time < 10:  # Increased timeout to 10 seconds
            self.client1.process_once(0.1)
            self.client2.process_once(0.1)
            
            for msg in self.received_messages:
                if len(msg) >= 4 and msg[0] == "nick":
                    if msg[1].split("!")[0] == "user1" and msg[3][0] == new_nickname:
                        nick_change_received = True
                    if msg[1].split("!")[0] == "user1" and msg[3][0] == new_nickname and msg[2] == "user1":
                        self_nick_change_received = True
                
            if nick_change_received and self_nick_change_received:
                break

        self.assertTrue(nick_change_received, "User2 did not receive the nickname change notification")
        self.assertTrue(self_nick_change_received, "User1 did not receive its own nickname change notification")

if __name__ == '__main__':
    unittest.main()
