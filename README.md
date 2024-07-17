<h1 align="center">
  <br>
  <img src="https://media.giphy.com/media/vFKqnCdLPNOKc/giphy.gif" alt="Easy-Meow" width="200">
  <br>
  🐱 Easy-Meow 🐱
  <br>
</h1>

<h4 align="center">A purr-fect wrapper for WhatsApp automation built on the Whatsmeow backbone</h4>

<p align="center">
  <img src="https://img.shields.io/badge/status-stable-brightgreen.svg" alt="Status: Stable">
  <img src="https://img.shields.io/badge/version-1.0.0-blue.svg" alt="Version: 1.0.11">
  <img src="https://img.shields.io/badge/made%20with-Go-00ADD8.svg" alt="Made with: Go">
</p>

<p align="center">
  <img src="https://media.giphy.com/media/JIX9t2j0ZTN9S/giphy.gif" alt="Coding Cat" width="600">
</p>

## 🚀 About Easy-Meow

Easy-Meow is a Go package that simplifies WhatsApp automation. Built on the powerful Whatsmeow backbone, it reduces boilerplate and makes WhatsApp bot development a breeze!

### 🌟 Key Features

- 📉 Less code, more functionality
- 🚀 Built on the robust Whatsmeow framework
- 🛠 Simplified API for common WhatsApp operations
- 🐱 Purr-fectly easy to integrate into your projects

<p align="center">
  <img src="https://media.giphy.com/media/3oKIPnAiaMCws8nOsE/giphy.gif" alt="Excited Cat" width="300">
</p>

## 🎉 What's New in v1.0.1

- **Diverse Message Types:**
  - **Text:** 📝 Send plain text messages effortlessly.
  - **Images:** 📸 Share images with optional captions.
  - **Videos:** 🎥 Send videos with captions.
  - **Audio:** 🎵 Share audio files (including PTT – push-to-talk).
  - **Documents:** 📄 Send documents with filenames and captions.
  - **Stickers:** 🎉 Share your favorite stickers.
  - **GIFs:** 🕺 Send animated GIFs with captions.
  - **Mentions:** 👥 Tag users in messages.
  - **Phone Numbers:** 📞 Send messages with clickable phone numbers.
  - **Reactions:** 😍 Add emojis to react to messages.

- **Advanced Functionality:**
  - **Replies:** 🔁 Reply to specific messages.
  - **Edits:** ✏️ Modify previously sent messages.
  - **Deletion:** 🗑️ Delete messages you've sent.

## 🔮 Future Plans

- 📚 Comprehensive documentation
- 🧪 Example projects
- 🎨 Customizable themes for bot responses
- 🌐 Multi-language support

## 💌 Stay Updated

Want to be notified about updates? Star this repository and watch for the latest news!

<p align="center">
  <img src="https://media.giphy.com/media/LmNwrBhejkK9EFP504/giphy.gif" alt="Cat Heart" width="60">
</p>

## 📦 Usage

To get started with Easy-Meow, follow these steps:

1. Initialize a new Go module:

    ```sh
    go mod init <your-module-name>
    ```

2. Get the Easy-Meow package:

    ```sh
    go get github.com/hacxk/easy-meow
    ```

3. Tidy up the module dependencies:

    ```sh
    go mod tidy
    ```

4. Import and use Easy-Meow in your Go code:

Here's a beautifully formatted `README.md` usage example for your Go program:

```markdown
# WhatsApp Client Example

This example demonstrates how to use the custom WhatsApp client to handle events and send messages using the `easy-meow` package.

## Usage

### Import the Necessary Packages

```go
package main

import (
	"context"
	"fmt"
	"log"

	whatsappclient "github.com/hacxk/easy-meow/Client" // Importing your custom WhatsApp client package

	"go.mau.fi/whatsmeow/types/events" // Importing events from the WhatsMeow library
)

// ### Define the Event Handler Function

// Define the event handler function
func myEventHandler(client *whatsappclient.WhatsAppClient, evt interface{}) {
	// Type switch to handle different event types
	switch v := evt.(type) {
	case *events.Message:
		// Extract message text from the message event (handle both conversation and extended messages)
		var messageText string
		if v.Message.Conversation != nil {
			messageText = *v.Message.Conversation // Extract text from Conversation message
		} else if v.Message.ExtendedTextMessage != nil {
			messageText = *v.Message.ExtendedTextMessage.Text // Extract text from ExtendedTextMessage
		}

		// Print image message info (if available) and sender details
		fmt.Print(v.Message.ImageMessage)
		fmt.Printf("Received a message from %s (%s): %s\n", v.Info.Sender.String(), v.Info.Chat.String(), messageText)

		// Check if the received message text is "hi"
		if messageText == "hi" {

			sock := client.GetClient() // Get the underlying client

			// Send a reply message
			_, err := sock.Reply(v, "Hello! 👋")
			if err != nil {
				log.Printf("Error sending reply: %v", err) // Log any error encountered while sending reply
			}
		}

	case *events.Receipt:
		// Handle receipt events
		fmt.Printf("Received a receipt: %+v\n", v)
	}
}

// ### Main Function

func main() {
	ctx := context.Background() // Create a context for managing the connection

	// Initialize the WhatsApp client with a store (database) file
	client, err := whatsappclient.NewWhatsAppClient("examplestore.db")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err) // Log and exit if client creation fails
	}

	// Connect the client
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err) // Log and exit if connection fails
	}

	// Add the event handler to the client
	client.AddEventHandler(func(evt interface{}) {
		myEventHandler(client, evt) // Call the event handler with each received event
	})

	// Check for any errors in sending the message
	if err != nil {
		log.Printf("Failed to send message: %v", err) // Log any error encountered while sending a message
	} else {
		fmt.Println("Message sent successfully") // Confirm successful message sending
	}

	// Keep the program running
	fmt.Println("Client is running. Press Ctrl+C to exit.") // Inform the user that the client is running
	select {}                                               // Block indefinitely to keep the program running
}
```


## 🙌 Credits

This project stands on the shoulders of giants:

- [Whatsmeow](https://github.com/tulir/whatsmeow) - The powerful backbone of Easy-Meow.

## 📄 License

MIT

<p align="center">
  Made with ❤️ by HACXK (Me Solo!)
</p>

<p align="center">
  <img src="https://media.giphy.com/media/WYEWpk4lRPDq0/giphy.gif" alt="Sleeping Cat" width="100">
</p>
