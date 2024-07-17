package messages

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	utils "easy-meow/Utils"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// Helper function to convert bool to int (0 or 1)
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func SendTextMessage(client *whatsmeow.Client, evt *events.Message, message string) (*whatsmeow.SendResponse, error) {
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}
	// Handle the error during SendMessage
	sendResp, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err) // Format error message
	}

	return &sendResp, nil // Return the response and nil error if successful
}

func ReactionMessage(client *whatsmeow.Client, evt *events.Message, reaction string) (*whatsmeow.SendResponse, error) {
	chat := evt.Info.Chat.String()
	isFromMe := evt.Info.IsFromMe
	id := evt.Info.ID

	msg := &waProto.Message{
		ReactionMessage: &waProto.ReactionMessage{
			Key: &waProto.MessageKey{
				RemoteJID: proto.String(chat),
				FromMe:    proto.Bool(isFromMe),
				ID:        proto.String(id),
			},
			Text:              proto.String(reaction),
			SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
	}

	// Handle the error during SendMessage
	sendResp, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err) // Format error message
	}

	return &sendResp, nil // Return the response and nil error if successful
}

func EditMessage(client *whatsmeow.Client, evt *events.Message, sentMessageID string, newMessageText string) (*whatsmeow.SendResponse, error) {
	// Create the EditedMessage content
	msg := &waProto.Message{
		ProtocolMessage: &waProto.ProtocolMessage{
			Key: &waProto.MessageKey{
				FromMe:      proto.Bool(true),
				RemoteJID:   proto.String(evt.Info.Chat.ToNonAD().String()),
				ID:          &sentMessageID,
				Participant: proto.String(client.Store.ID.ToNonAD().String()),
			},
			Type:                      waProto.ProtocolMessage_MESSAGE_EDIT.Enum(),
			EphemeralExpiration:       proto.Uint32(0),
			EphemeralSettingTimestamp: proto.Int64(0),
			EditedMessage:             &waProto.Message{}, // Initialize EditedMessage to an empty struct
			TimestampMS:               proto.Int64(0),
		},
	}

	// Safely determine the message type and set the edited text
	if evt.Message.Conversation != nil {
		msg.ProtocolMessage.EditedMessage.Conversation = &newMessageText
	} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
		msg.ProtocolMessage.EditedMessage.ExtendedTextMessage = &waProto.ExtendedTextMessage{
			Text: &newMessageText,
		}
	} else if evt.Message.ImageMessage != nil && evt.Message.ImageMessage.Caption != nil {
		msg.ProtocolMessage.EditedMessage.ImageMessage = &waProto.ImageMessage{
			Caption: &newMessageText,
		}
	} else if evt.Message.VideoMessage != nil && evt.Message.VideoMessage.Caption != nil {
		msg.ProtocolMessage.EditedMessage.VideoMessage = &waProto.VideoMessage{
			Caption: &newMessageText,
		}
	} else {
		return nil, fmt.Errorf("unsupported message type for editing")
	}

	// Send the edit request
	sendResp, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message: %v", err)
	}

	return &sendResp, nil
}

func ReplyToMessage(client *whatsmeow.Client, evt *events.Message, message string) (*whatsmeow.SendResponse, error) {
	recipientJID := evt.Info.Sender.String()

	// Split the JID to remove any device part
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]

	// Reconstruct the JID
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %v", err)
	}

	quotedMessageID := evt.Info.ID

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      &quotedMessageID,
				Participant:   proto.String(jid.String()),
				QuotedMessage: evt.Message,
			},
		},
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendImageMessage(client *whatsmeow.Client, evt *events.Message, imageFile string, caption string) (*whatsmeow.SendResponse, error) {
	// Upload Image
	data, err := os.ReadFile(imageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %v", err)
	}

	// Determine MIME type
	mimeType := http.DetectContentType(data)
	if mimeType == "" {
		mimeType = "image/jpeg" // Default to JPEG if detection fails
	}

	// Create Image Message
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Mimetype:      proto.String(mimeType),
			StaticURL:     proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
		},
	}

	if caption != "" {
		msg.ImageMessage.Caption = proto.String(caption)
	}

	// Get and Set Thumbnail (With Error Handling)
	thumbnailBytes, err := utils.GetThumbnail(imageFile)
	if err == nil { // Only set thumbnail if no errors occurred
		msg.ImageMessage.JPEGThumbnail = thumbnailBytes
	}

	// Send Message
	sendResp, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send image message: %v", err)
	}
	return &sendResp, nil
}

func SendImageMessageReply(client *whatsmeow.Client, evt *events.Message, imageFile string, caption string) (*whatsmeow.SendResponse, error) {
	recipientJID := evt.Info.Sender.String()

	// Split the JID to remove any device part
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]

	// Reconstruct the JID
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Read the image file
	data, err := os.ReadFile(imageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %v", err)
	}

	// Upload the image to WhatsApp
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %v", err)
	}

	// Create the image message
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Mimetype:      proto.String(http.DetectContentType(data)),
			StaticURL:     proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			ContextInfo: &waProto.ContextInfo{
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

	// Get and Set Thumbnail (With Error Handling)
	thumbnailBytes, err := utils.GetThumbnail(imageFile)
	if err == nil { // Only set thumbnail if no errors occurred
		msg.ImageMessage.JPEGThumbnail = thumbnailBytes
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendVideoMessage(client *whatsmeow.Client, evt *events.Message, videoFile string, caption string) (*whatsmeow.SendResponse, error) {
	// Read the video file
	data, err := os.ReadFile(videoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read video file: %v", err)
	}

	// Upload the video to WhatsApp
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %v", err)
	}

	// Create the video message
	msg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
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

	// Get and Set Thumbnail (With Error Handling)
	thumbnailBytes, err := utils.GetThumbnail(videoFile)
	if err == nil { // Only set thumbnail if no errors occurred
		msg.VideoMessage.JPEGThumbnail = thumbnailBytes
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendVideoMessageReply(client *whatsmeow.Client, evt *events.Message, videoFile string, caption string) (*whatsmeow.SendResponse, error) {
	recipientJID := evt.Info.Sender.String()
	// Split the JID to remove any device part
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	// Reconstruct the JID
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %v", err)
	}

	// Read the video file
	data, err := os.ReadFile(videoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read video file: %v", err)
	}

	// Upload the video to WhatsApp
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %v", err)
	}

	// Create the video message
	msg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       proto.String(caption),
			ContextInfo: &waProto.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	// Get and Set Thumbnail (With Error Handling)
	thumbnailBytes, err := utils.GetThumbnail(videoFile)
	if err == nil { // Only set thumbnail if no errors occurred
		msg.VideoMessage.JPEGThumbnail = thumbnailBytes
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendAudioMessage(client *whatsmeow.Client, evt *events.Message, audioFile string, ptt bool) (*whatsmeow.SendResponse, error) {
	data, err := os.ReadFile(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio: %v", err)
	}

	msg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
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

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendAudioMessageReply(client *whatsmeow.Client, evt *events.Message, audioFile string, ptt bool) (*whatsmeow.SendResponse, error) {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %v", err)
	}

	data, err := os.ReadFile(audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio: %v", err)
	}

	msg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			PTT:           proto.Bool(ptt),
			ContextInfo: &waProto.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendDocumentMessage(client *whatsmeow.Client, evt *events.Message, documentFile string, fileName string, caption string) (*whatsmeow.SendResponse, error) {
	data, err := os.ReadFile(documentFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read document file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %v", err)
	}

	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
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

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendDocumentMessageReply(client *whatsmeow.Client, evt *events.Message, documentFile string, fileName string, caption string) (*whatsmeow.SendResponse, error) {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %v", err)
	}

	data, err := os.ReadFile(documentFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read document file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %v", err)
	}

	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileName:      proto.String(fileName),
			Caption:       proto.String(caption),
			ContextInfo: &waProto.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendStickerMessage(client *whatsmeow.Client, evt *events.Message, stickerFile string) (*whatsmeow.SendResponse, error) {
	data, err := os.ReadFile(stickerFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read sticker file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload sticker: %v", err)
	}

	msg := &waProto.Message{
		StickerMessage: &waProto.StickerMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String("image/webp"),
		},
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendGifMessage(client *whatsmeow.Client, evt *events.Message, gifFile string, caption string) (*whatsmeow.SendResponse, error) {
	data, err := os.ReadFile(gifFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read GIF file: %v", err)
	}

	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to upload GIF: %v", err)
	}

	msg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
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

	// Get and Set Thumbnail (With Error Handling)
	thumbnailBytes, err := utils.GetThumbnail(gifFile)
	if err == nil { // Only set thumbnail if no errors occurred
		msg.VideoMessage.JPEGThumbnail = thumbnailBytes
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendStickerMessageReply(client *whatsmeow.Client, evt *events.Message, stickerFile string) (*whatsmeow.SendResponse, error) {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %v", err)
	}
	data, err := os.ReadFile(stickerFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read sticker file: %v", err)
	}
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload sticker: %v", err)
	}
	msg := &waProto.Message{
		StickerMessage: &waProto.StickerMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String("image/webp"),
			ContextInfo: &waProto.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendGifMessageReply(client *whatsmeow.Client, evt *events.Message, gifFile string, caption string) (*whatsmeow.SendResponse, error) {
	recipientJID := evt.Info.Sender.String()
	parts := strings.SplitN(recipientJID, ":", 2)
	recipient := parts[0]
	jid, err := types.ParseJID(recipient + "@" + evt.Info.Sender.Server)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %v", err)
	}
	data, err := os.ReadFile(gifFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read GIF file: %v", err)
	}
	uploaded, err := client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to upload GIF: %v", err)
	}
	msg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			Mimetype:      proto.String("video/mp4"),
			Caption:       proto.String(caption),
			GifPlayback:   proto.Bool(true),
			ContextInfo: &waProto.ContextInfo{
				QuotedMessage: evt.Message,
				StanzaID:      proto.String(evt.Info.ID),
				Participant:   proto.String(jid.String()),
			},
		},
	}

	// Get and Set Thumbnail (With Error Handling)
	thumbnailBytes, err := utils.GetThumbnail(gifFile)
	if err == nil { // Only set thumbnail if no errors occurred
		msg.VideoMessage.JPEGThumbnail = thumbnailBytes
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err = client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendMentionMessage(client *whatsmeow.Client, evt *events.Message, message string, mentions []string) (*whatsmeow.SendResponse, error) {
	mentionedJIDs := make([]string, len(mentions))
	for i, mention := range mentions {
		mentionedJID, err := types.ParseJID(mention)
		if err != nil {
			return nil, fmt.Errorf("invalid mentioned JID: %v", err)
		}
		mentionedJIDs[i] = mentionedJID.String()
	}

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID: mentionedJIDs,
			},
		},
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendPhoneNumberMessage(client *whatsmeow.Client, evt *events.Message, phoneNumber string, message string) (*whatsmeow.SendResponse, error) {
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
			ContextInfo: &waProto.ContextInfo{
				Participant: proto.String(phoneNumber + "@s.whatsapp.net"),
			},
		},
	}

	var sendResp whatsmeow.SendResponse
	sendResp, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func SendPolls(client *whatsmeow.Client, evt *events.Message, question string, pollOptions []string, onlyOnce bool) (*whatsmeow.SendResponse, error) {
	// Create options for the PollCreationMessage
	var options []*waProto.PollCreationMessage_Option
	for _, optionText := range pollOptions {
		options = append(options, &waProto.PollCreationMessage_Option{
			OptionName: proto.String(optionText),
		})
	}

	msg := &waProto.Message{
		PollCreationMessage: &waProto.PollCreationMessage{
			Name:                   proto.String(question),
			Options:                options,
			SelectableOptionsCount: proto.Uint32(uint32(btoi(onlyOnce))), // Fix
		},
	}

	sendResp, err := client.SendMessage(context.Background(), evt.Info.Chat, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	return &sendResp, nil
}

func DeleteMessage(client *whatsmeow.Client, evt *events.Message, messageID string) (*whatsmeow.SendResponse, error) {
	// Construct ProtocolMessage to revoke the message
	revokeMsg := &waProto.Message{
		ProtocolMessage: &waProto.ProtocolMessage{
			Key: &waProto.MessageKey{
				RemoteJID: nil, // Initialize with nil
				ID:        proto.String(messageID),
				FromMe:    proto.Bool(evt.Info.IsFromMe),
			},
			Type: waProto.ProtocolMessage_REVOKE.Enum(),
		},
	}

	// Check if evt.Info.Chat is not empty (zero value check)
	if evt.Info.Chat.String() != "" {
		revokeMsg.ProtocolMessage.Key.RemoteJID = proto.String(evt.Info.Chat.String())
	}

	// Send the revoke message using the JID
	getres, err := client.SendMessage(context.Background(), evt.Info.Chat, revokeMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to delete message: %v", err)
	}

	return &getres, nil
}
