package teller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func GenerateWebhookSignatures(timestamp time.Time, body []byte, secrets []string) []string {
	signatures := make([]string, len(secrets))
	for i := range secrets {
		secret := secrets[i]
		data := fmt.Sprintf("%d.%s", timestamp.Unix(), string(body))
		sig := hmac.New(sha256.New, []byte(secret))
		sig.Write([]byte(data))
		dataHmac := sig.Sum(nil)
		hmacHex := hex.EncodeToString(dataHmac)
		signatures[i] = hmacHex
	}
	return signatures
}
