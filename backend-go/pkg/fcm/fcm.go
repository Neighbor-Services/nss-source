package fcm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client interface {
	SendToToken(ctx context.Context, token, title, body string, data map[string]string) error
	SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) ([]string, error)
}

type fcmClient struct {
	msgClient *messaging.Client
}

// NewClient initializes the Firebase Messaging Client using a credentials file path or credentials JSON string.
func NewClient(credentialsPath, credentialsJSON string) (Client, error) {
	ctx := context.Background()
	var opts []option.ClientOption

	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	} else if credentialsPath != "" {
		if _, err := os.Stat(credentialsPath); err == nil {
			opts = append(opts, option.WithCredentialsFile(credentialsPath))
		} else {
			return nil, fmt.Errorf("firebase credentials file not found at %s: %w", credentialsPath, err)
		}
	} else {
		// Fallback to standard environment variable or default credentials
		if defaultCred := os.Getenv("FIREBASE_CREDENTIALS_FILE"); defaultCred != "" {
			opts = append(opts, option.WithCredentialsFile(defaultCred))
		} else if googleAppCred := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); googleAppCred != "" {
			opts = append(opts, option.WithCredentialsFile(googleAppCred))
		} else {
			return nil, errors.New("no firebase credentials provided")
		}
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase messaging client: %w", err)
	}

	log.Println("[FCM] Firebase Cloud Messaging client initialized successfully")
	return &fcmClient{msgClient: msgClient}, nil
}

func (c *fcmClient) buildMessagePayload(title, body string, data map[string]string) (*messaging.Notification, *messaging.AndroidConfig, *messaging.APNSConfig, map[string]string) {
	if data == nil {
		data = make(map[string]string)
	}
	// Always ensure title and message are in data payload for Flutter background handlers
	if _, ok := data["title"]; !ok && title != "" {
		data["title"] = title
	}
	if _, ok := data["message"]; !ok && body != "" {
		data["message"] = body
	}

	notif := &messaging.Notification{
		Title: title,
		Body:  body,
	}

	androidConfig := &messaging.AndroidConfig{
		Priority: "high",
		Notification: &messaging.AndroidNotification{
			ChannelID: "high_importance_channel",
			Sound:     "default",
			Priority:  messaging.PriorityHigh,
		},
	}

	apnsConfig := &messaging.APNSConfig{
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Sound:            "default",
				ContentAvailable: true,
			},
		},
	}

	return notif, androidConfig, apnsConfig, data
}

func (c *fcmClient) SendToToken(ctx context.Context, token, title, body string, data map[string]string) error {
	if token == "" {
		return errors.New("empty device token")
	}

	notif, androidConfig, apnsConfig, payloadData := c.buildMessagePayload(title, body, data)

	msg := &messaging.Message{
		Token:        token,
		Notification: notif,
		Android:      androidConfig,
		APNS:         apnsConfig,
		Data:         payloadData,
	}

	_, err := c.msgClient.Send(ctx, msg)
	if err != nil {
		log.Printf("[FCM] Error sending message to token %s...: %v", token[:min(10, len(token))], err)
		return err
	}

	return nil
}

// SendMulticast sends a push notification to up to 500 tokens in a batch.
// It returns a list of failed/invalid tokens that should be cleaned up.
func (c *fcmClient) SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) ([]string, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	notif, androidConfig, apnsConfig, payloadData := c.buildMessagePayload(title, body, data)

	var invalidTokens []string

	// Firebase allows max 500 tokens per batch
	const batchSize = 500
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]

		msg := &messaging.MulticastMessage{
			Tokens:       batch,
			Notification: notif,
			Android:      androidConfig,
			APNS:         apnsConfig,
			Data:         payloadData,
		}

		response, err := c.msgClient.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Printf("[FCM] Error sending multicast batch: %v", err)
			continue
		}

		for idx, res := range response.Responses {
			if !res.Success {
				t := batch[idx]
				if messaging.IsRegistrationTokenNotRegistered(res.Error) ||
					messaging.IsInvalidArgument(res.Error) {
					invalidTokens = append(invalidTokens, t)
				}
				log.Printf("[FCM] Multicast item %d failed: %v", idx, res.Error)
			}
		}
	}

	return invalidTokens, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
