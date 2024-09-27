import unittest
import subprocess
import time
import irc.client

class TestChannelCommunication(unittest.TestCase):
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
        self.connection1.add_global_handler("pubmsg", self.on_pubmsg)
        self.connection2.add_global_handler("pubmsg", self.on_pubmsg)

    def tearDown(self):
        self.connection1.disconnect()
        self.connection2.disconnect()

    def on_pubmsg(self, connection, event):
        self.received_messages.append(event.arguments[0])

    def test_channel_communication(self):
        channel = "#testchannel"
        message1 = "Hello, World!"
        message2 = "This is another message."

        # Join the channel
        self.connection1.join(channel)
        self.connection2.join(channel)

        # Give some time for the joins to complete
        time.sleep(1)

        # Send messages from both clients
        self.connection1.privmsg(channel, message1)
        self.connection2.privmsg(channel, message2)

        # Wait for the messages to be received
        start_time = time.time()
        while len(self.received_messages) < 2 and time.time() - start_time < 5:
            self.client1.process_once(0.1)
            self.client2.process_once(0.1)

        # Check if both messages were received
        self.assertEqual(len(self.received_messages), 2)
        self.assertIn(message1, self.received_messages)
        self.assertIn(message2, self.received_messages)

if __name__ == '__main__':
    unittest.main()
