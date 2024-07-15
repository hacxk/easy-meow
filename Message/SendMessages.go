package messages

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// SendTextMessage sends a text message
func SendTextMessage(client *whatsmeow.Client, to string, message string) error {
	recipient, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}
	_, err = client.SendMessage(context.Background(), recipient, msg)
	return err
}

// ReplyToMessage replies to a specific message
func ReplyToMessage(client *whatsmeow.Client, to string, quotedMessageID string, message string) error {
	recipient, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      &quotedMessageID,
				Participant:   proto.String(recipient.String()),
				QuotedMessage: &waProto.Message{Conversation: proto.String("Quoted message")},
			},
		},
	}

	_, err = client.SendMessage(context.Background(), recipient, msg)
	return err
}

func SendImageMessage(client *whatsmeow.Client, to string, imageFile string, caption string) error {
	recipient, err := types.ParseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Read the image file
	data, err := os.ReadFile(imageFile)
	if err != nil {
		return fmt.Errorf("failed to read image file: %v", err)
	}

	// Upload the image to WhatsApp
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %v", err)
	}

	// Create the image message
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       proto.String(caption),
			StaticURL:     proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
		},
	}

	_, err = client.SendMessage(context.Background(), recipient, msg)
	return err
}
