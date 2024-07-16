package messages

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"strings" // Add this import statement

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func SendTextMessage(client *whatsmeow.Client, evt *events.Message, message string) error {
	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}
	var err error
	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func ReactionMessage(client *whatsmeow.Client, evt *events.Message, reaction string) error {
	// Extract values and create variables to hold them:
	chat := evt.Info.Chat.String()
	isFromMe := evt.Info.IsFromMe
	id := evt.Info.ID

	msg := &waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{
			Text: proto.String(reaction),
			Key: &waE2E.MessageKey{
				ID:        &id,
				FromMe:    &isFromMe,
				RemoteJID: &chat,
			},
		},
	}

	_, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func ReplyToMessage(client *whatsmeow.Client, evt *events.Message, message string) error {
	recipientJID := evt.Info.Sender.String()

	// Split the JID to remove any device part
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]

	// Reconstruct the JID
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	quotedMessageID := evt.Info.ID

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(message),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      &quotedMessageID,
				Participant:   proto.String(jid.String()),
				QuotedMessage: evt.Message,
			},
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendImageMessage(client *whatsmeow.Client, evt *events.Message, imageFile string, caption string) error {
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
	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Mimetype:      proto.String(http.DetectContentType(data)),
			StaticURL:     proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
		},
	}

	// Add caption if provided
	if caption != "" {
		msg.ImageMessage.Caption = proto.String(caption)
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendImageMessageReply(client *whatsmeow.Client, evt *events.Message, imageFile string, caption string) error {
	recipientJID := evt.Info.Sender.String()

	// Split the JID to remove any device part
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]

	// Reconstruct the JID
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
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
	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Mimetype:      proto.String(http.DetectContentType(data)),
			StaticURL:     proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			ContextInfo: &waE2E.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	// Add caption if provided
	if caption != "" {
		msg.ImageMessage.Caption = proto.String(caption)
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendVideoMessage(client *whatsmeow.Client, evt *events.Message, videoFile string, caption string) error {
	// Read the video file
	data, err := os.ReadFile(videoFile)
	if err != nil {
		return fmt.Errorf("failed to read video file: %v", err)
	}

	// Upload the video to WhatsApp
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload video: %v", err)
	}

	// Create the video message
	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       proto.String(caption),
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendVideoMessageReply(client *whatsmeow.Client, evt *events.Message, videoFile string, caption string) error {
	recipientJID := evt.Info.Sender.String()
	// Split the JID to remove any device part
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	// Reconstruct the JID
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Read the video file
	data, err := os.ReadFile(videoFile)
	if err != nil {
		return fmt.Errorf("failed to read video file: %v", err)
	}

	// Upload the video to WhatsApp
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload video: %v", err)
	}

	// Create the video message
	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       proto.String(caption),
			ContextInfo: &waE2E.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendAudioMessage(client *whatsmeow.Client, evt *events.Message, audioFile string, ptt bool) error {
	data, err := os.ReadFile(audioFile)
	if err != nil {
		return fmt.Errorf("failed to read audio file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		return fmt.Errorf("failed to upload audio: %v", err)
	}

	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			PTT:           proto.Bool(ptt),
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendAudioMessageReply(client *whatsmeow.Client, evt *events.Message, audioFile string, ptt bool) error {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	data, err := os.ReadFile(audioFile)
	if err != nil {
		return fmt.Errorf("failed to read audio file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		return fmt.Errorf("failed to upload audio: %v", err)
	}

	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			PTT:           proto.Bool(ptt),
			ContextInfo: &waE2E.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendDocumentMessage(client *whatsmeow.Client, evt *events.Message, documentFile string, fileName string, caption string) error {
	data, err := os.ReadFile(documentFile)
	if err != nil {
		return fmt.Errorf("failed to read document file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload document: %v", err)
	}

	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileName:      proto.String(fileName),
			Caption:       proto.String(caption),
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendDocumentMessageReply(client *whatsmeow.Client, evt *events.Message, documentFile string, fileName string, caption string) error {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}

	data, err := os.ReadFile(documentFile)
	if err != nil {
		return fmt.Errorf("failed to read document file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload document: %v", err)
	}

	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileName:      proto.String(fileName),
			Caption:       proto.String(caption),
			ContextInfo: &waE2E.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendStickerMessage(client *whatsmeow.Client, evt *events.Message, stickerFile string) error {
	data, err := os.ReadFile(stickerFile)
	if err != nil {
		return fmt.Errorf("failed to read sticker file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload sticker: %v", err)
	}

	msg := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String("image/webp"),
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendGifMessage(client *whatsmeow.Client, evt *events.Message, gifFile string, caption string) error {
	data, err := os.ReadFile(gifFile)
	if err != nil {
		return fmt.Errorf("failed to read GIF file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload GIF: %v", err)
	}

	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String("video/mp4"),
			Caption:       proto.String(caption),
			GifPlayback:   proto.Bool(true),
		},
	}

	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendStickerMessageReply(client *whatsmeow.Client, evt *events.Message, stickerFile string) error {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}
	data, err := os.ReadFile(stickerFile)
	if err != nil {
		return fmt.Errorf("failed to read sticker file: %v", err)
	}
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload sticker: %v", err)
	}
	msg := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String("image/webp"),
			ContextInfo: &waE2E.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}
	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendGifMessageReply(client *whatsmeow.Client, evt *events.Message, gifFile string, caption string) error {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return fmt.Errorf("invalid recipient JID: %v", err)
	}
	data, err := os.ReadFile(gifFile)
	if err != nil {
		return fmt.Errorf("failed to read GIF file: %v", err)
	}
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return fmt.Errorf("failed to upload GIF: %v", err)
	}
	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String("video/mp4"),
			Caption:       proto.String(caption),
			GifPlayback:   proto.Bool(true),
			ContextInfo: &waE2E.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}
	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendMentionMessage(client *whatsmeow.Client, evt *events.Message, message string, mentions []string) error {
	mentionedJIDs := make([]string, len(mentions))
	for i, mention := range mentions {
		mentionedJID, err := types.ParseJID(mention)
		if err != nil {
			return fmt.Errorf("invalid mentioned JID: %v", err)
		}
		mentionedJIDs[i] = mentionedJID.String()
	}

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(message),
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: mentionedJIDs,
			},
		},
	}

	var err error
	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}

func SendPhoneNumberMessage(client *whatsmeow.Client, evt *events.Message, phoneNumber string, message string) error {
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(message),
			ContextInfo: &waE2E.ContextInfo{
				Participant: proto.String(phoneNumber + "@s.whatsapp.net"),
			},
		},
	}

	var err error
	_, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	return err
}
