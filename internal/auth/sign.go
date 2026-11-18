package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func Sign(token, msg string) string {
	mac := hmac.New(sha256.New, []byte(token))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

func Verify(token, msg, sig string) bool {
	return hmac.Equal([]byte(Sign(token, msg)), []byte(sig))
}

func Message(sessionID string, requestID int64, action, letter string) string {
	return fmt.Sprintf("%s|%d|%s|%s", sessionID, requestID, action, letter)
}
