package google

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

type Config struct {
	BaseURL      string
	KeyName      string
	SignedKeyB64 string
}

type Signer struct {
	baseURL   string
	keyName   string
	signedKey []byte
}

func NewSigner(config Config) (*Signer, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("google cdn base_url is required")
	}
	keyName := strings.TrimSpace(config.KeyName)
	if keyName == "" {
		return nil, fmt.Errorf("google cdn key_name is required")
	}
	rawKey := strings.TrimSpace(config.SignedKeyB64)
	if rawKey == "" {
		return nil, fmt.Errorf("google cdn signed key is required")
	}
	decodedKey, err := decodeSigningKey(rawKey)
	if err != nil {
		return nil, err
	}
	return &Signer{baseURL: baseURL, keyName: keyName, signedKey: decodedKey}, nil
}

func (signer *Signer) SignedObjectURL(objectPath string, expiresAt time.Time) (string, error) {
	if signer == nil {
		return "", fmt.Errorf("google cdn signer is not initialized")
	}
	cleanObjectPath := strings.Trim(path.Clean(strings.TrimSpace(objectPath)), "/")
	if cleanObjectPath == "" {
		return "", fmt.Errorf("object_path is required")
	}
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(15 * time.Minute)
	}
	expires := expiresAt.UTC().Unix()
	escapedObjectPath := strings.ReplaceAll(url.PathEscape(cleanObjectPath), "%2F", "/")
	unsignedURL := fmt.Sprintf("%s/%s?Expires=%d&KeyName=%s", signer.baseURL, escapedObjectPath, expires, url.QueryEscape(signer.keyName))
	mac := hmac.New(sha1.New, signer.signedKey)
	if _, err := mac.Write([]byte(unsignedURL)); err != nil {
		return "", fmt.Errorf("sign google cdn url: %w", err)
	}
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return unsignedURL + "&Signature=" + url.QueryEscape(signature), nil
}

func decodeSigningKey(value string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	decoded, err = base64.URLEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	decoded, err = base64.RawStdEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	decoded, err = base64.StdEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	return nil, fmt.Errorf("decode google cdn signed key: invalid base64")
}

var _ applicationcontrolplane.ProjectCDNSigner = (*Signer)(nil)
