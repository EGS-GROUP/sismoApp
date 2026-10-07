package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/SherClockHolmes/webpush-go"
)

// The VAPID keys we generated earlier
const (
	vapidPublicKey  = "BAbxrgETNt_uTCMa4VaQbdl23plVghM7KrhccCessv8KCGvc3G33nMNTgpWRQwjv1nnKGp1HVuRDyjcqhNxzs6Y"
	vapidPrivateKey = "uluwcxkMCgatrP8WKGniWkmyZreN5rnq812fhlRKiNs"
)

type PushSubscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func saveSubscription(sub PushSubscription) error {
	insertSQL := `INSERT INTO push_subscriptions (endpoint, keys_p256dh, keys_auth, created_at) VALUES (?, ?, ?, ?) ON CONFLICT(endpoint) DO UPDATE SET keys_p256dh=excluded.keys_p256dh, keys_auth=excluded.keys_auth`
	_, err := db.Exec(insertSQL, sub.Endpoint, sub.Keys.P256dh, sub.Keys.Auth, time.Now().Unix())
	return err
}

func getSubscriptions() ([]webpush.Subscription, error) {
	rows, err := db.Query(`SELECT endpoint, keys_p256dh, keys_auth FROM push_subscriptions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []webpush.Subscription
	for rows.Next() {
		var endpoint, p256dh, auth string
		if err := rows.Scan(&endpoint, &p256dh, &auth); err != nil {
			return nil, err
		}
		subs = append(subs, webpush.Subscription{
			Endpoint: endpoint,
			Keys: webpush.Keys{
				P256dh: p256dh,
				Auth:   auth,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

func sendPushNotificationToAll(title, body string) {
	subs, err := getSubscriptions()
	if err != nil {
		log.Printf("Error getting subscriptions: %v", err)
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"title": title,
		"body":  body,
		"url":   "/",
	})

	for _, sub := range subs {
		// Send Notification
		resp, err := webpush.SendNotification(payload, &sub, &webpush.Options{
			Subscriber:      "mailto:admin@jagmedia.com.ve", // Required email
			VAPIDPublicKey:  vapidPublicKey,
			VAPIDPrivateKey: vapidPrivateKey,
			TTL:             30,
		})
		if err != nil {
			log.Printf("Error sending push: %v", err)
		} else {
			resp.Body.Close()
		}
	}
}
