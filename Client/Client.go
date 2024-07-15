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
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ExtendedClient struct {
	*whatsmeow.Client
}

func (ec *ExtendedClient) Send(to string, message string) error {
	return messages.SendTextMessage(ec.Client, to, message)
}

func (ec *ExtendedClient) Reply(to string, quotedMessageID string, message string) error {
	return messages.ReplyToMessage(ec.Client, to, quotedMessageID, message)
}

func (ec *ExtendedClient) SendImage(to string, path string, caption string) error {
	return messages.SendImageMessage(ec.Client, to, path, caption)
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
