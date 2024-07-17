package whatsappclient

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	messages "easy-meow/Message"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ExtendedClient struct {
	*whatsmeow.Client
}

func (ec *ExtendedClient) Send(evt *events.Message, message string) (*whatsmeow.SendResponse, error) {
	return messages.SendTextMessage(ec.Client, evt, message)
}

func (ec *ExtendedClient) React(evt *events.Message, emoji string) (*whatsmeow.SendResponse, error) {
	return messages.ReactionMessage(ec.Client, evt, emoji)
}

func (ec *ExtendedClient) Edit(evt *events.Message, sendMessageID string, newMessage string) (*whatsmeow.SendResponse, error) {
	return messages.EditMessage(ec.Client, evt, sendMessageID, newMessage)
}

func (ec *ExtendedClient) Reply(evt *events.Message, message string) (*whatsmeow.SendResponse, error) {
	return messages.ReplyToMessage(ec.Client, evt, message)
}

func (ec *ExtendedClient) SendImage(evt *events.Message, path string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendImageMessage(ec.Client, evt, path, finalCaption)
}

func (ec *ExtendedClient) SendImageReply(evt *events.Message, path string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendImageMessageReply(ec.Client, evt, path, finalCaption)
}

func (ec *ExtendedClient) SendVideo(evt *events.Message, path string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendVideoMessage(ec.Client, evt, path, finalCaption)
}

func (ec *ExtendedClient) SendVideoReply(evt *events.Message, path string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendVideoMessageReply(ec.Client, evt, path, finalCaption)
}

func (ec *ExtendedClient) SendAudio(evt *events.Message, path string, ptt bool) (*whatsmeow.SendResponse, error) {
	return messages.SendAudioMessage(ec.Client, evt, path, ptt)
}

func (ec *ExtendedClient) SendAudioReply(evt *events.Message, path string, ptt bool) (*whatsmeow.SendResponse, error) {
	return messages.SendAudioMessageReply(ec.Client, evt, path, ptt)
}

func (ec *ExtendedClient) SendDocument(evt *events.Message, path string, filename string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendDocumentMessage(ec.Client, evt, path, filename, finalCaption)
}

func (ec *ExtendedClient) SendDocumentReply(evt *events.Message, path string, filename string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendDocumentMessageReply(ec.Client, evt, path, filename, finalCaption)
}

func (ec *ExtendedClient) SendSticker(evt *events.Message, path string) (*whatsmeow.SendResponse, error) {
	return messages.SendStickerMessage(ec.Client, evt, path)
}

func (ec *ExtendedClient) SendStickerReply(evt *events.Message, path string) (*whatsmeow.SendResponse, error) {
	return messages.SendStickerMessageReply(ec.Client, evt, path)
}

func (ec *ExtendedClient) SendGif(evt *events.Message, path string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendGifMessage(ec.Client, evt, path, finalCaption)
}

func (ec *ExtendedClient) SendGifReply(evt *events.Message, path string, caption ...string) (*whatsmeow.SendResponse, error) {
	var finalCaption string
	if len(caption) > 0 {
		finalCaption = caption[0]
	}
	return messages.SendGifMessageReply(ec.Client, evt, path, finalCaption)
}

func (ec *ExtendedClient) SendMention(evt *events.Message, message string, mentions []string) (*whatsmeow.SendResponse, error) {
	return messages.SendMentionMessage(ec.Client, evt, message, mentions)
}

func (ec *ExtendedClient) SendPhone(evt *events.Message, phonenumber string, message string) (*whatsmeow.SendResponse, error) {
	return messages.SendPhoneNumberMessage(ec.Client, evt, phonenumber, message)
}

func (ec *ExtendedClient) CreatePoll(evt *events.Message, question string, option []string, onlyonce bool) (*whatsmeow.SendResponse, error) {
	return messages.SendPolls(ec.Client, evt, question, option, onlyonce)
}

func (ec *ExtendedClient) Delete(evt *events.Message, messageID string) (*whatsmeow.SendResponse, error) {
	return messages.DeleteMessage(ec.Client, evt, messageID)
}

type WhatsAppClient struct {
	client *ExtendedClient
	dbPath string
}

func NewWhatsAppClient(dbPath string) (*WhatsAppClient, error) {
	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath), dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQL store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get device from store: %w", err)
	}

	clientLog := waLog.Stdout("WhatsApp", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	extendedClient := &ExtendedClient{Client: client}

	return &WhatsAppClient{
		client: extendedClient,
		dbPath: dbPath,
	}, nil
}

func (wac *WhatsAppClient) Connect(ctx context.Context) error {
	if wac.client.Store.ID == nil {
		qrChan, _ := wac.client.GetQRChannel(ctx)
		err := wac.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("Scan the QR code above to log in")
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err := wac.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
	}

	wac.setupSignalHandler()
	return nil
}

func (wac *WhatsAppClient) GetClient() *ExtendedClient {
	return wac.client
}

func (wac *WhatsAppClient) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nDisconnecting...")
		wac.client.Disconnect()
		os.Exit(0)
	}()
}

func (wac *WhatsAppClient) Disconnect() {
	wac.client.Disconnect()
}

func (wac *WhatsAppClient) IsConnected() bool {
	return wac.client.IsConnected()
}

func (wac *WhatsAppClient) AddEventHandler(handler func(interface{})) {
	wac.client.AddEventHandler(handler)
}
