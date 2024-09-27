import unittest
import subprocess
import time
import irc.client
import threading

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
        self.client1 = irc.client.IRC()
        self.client2 = irc.client.IRC()

    def tearDown(self):
        if hasattr(self, 'connection1'):
            self.connection1.disconnect()
        if hasattr(self, 'connection2'):
            self.connection2.disconnect()

    def test_channel_communication(self):
        channel = "#testchannel"
        message = "Hello, World!"
        received_messages = []

        def on_pubmsg(connection, event):
            received_messages.append(event.arguments[0])

        # Connect client1
        self.connection1 = self.client1.server().connect("localhost", 6667, "user1")
        self.connection1.add_global_handler("pubmsg", on_pubmsg)

        # Connect client2
        self.connection2 = self.client2.server().connect("localhost", 6667, "user2")
        self.connection2.add_global_handler("pubmsg", on_pubmsg)

        # Join the channel
        self.connection1.join(channel)
        self.connection2.join(channel)

        # Give some time for the joins to complete
        time.sleep(1)

        # Send a message from client1
        self.connection1.privmsg(channel, message)

        # Wait for the message to be received
        start_time = time.time()
        while len(received_messages) < 1 and time.time() - start_time < 5:
            self.client1.process_once(0.1)
            self.client2.process_once(0.1)

        # Check if client2 received the message
        self.assertEqual(len(received_messages), 1)
        self.assertEqual(received_messages[0], message)

if __name__ == '__main__':
    unittest.main()
