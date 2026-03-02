package postgres

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	scmTokenEnvelopeVersion = "v1"
	keyStatusActive         = "active"
	keyStatusRetired        = "retired"
)

type scmTokenKeyRecord struct {
	gorm.Model
	KeyID       string `gorm:"column:key_id;size:64;not null;uniqueIndex"`
	Status      string `gorm:"column:status;size:32;not null;index"`
	KeyMaterial string `gorm:"column:key_material;type:text;not null"`
}

func (scmTokenKeyRecord) TableName() string {
	return "scm_token_keys"
}

type SCMTokenCrypto struct {
	db *gorm.DB
}

func NewSCMTokenCrypto(db *gorm.DB) (*SCMTokenCrypto, error) {
	if db == nil {
		return nil, fmt.Errorf("scm token crypto db is required")
	}
	if err := db.AutoMigrate(&scmTokenKeyRecord{}); err != nil {
		return nil, fmt.Errorf("migrate scm token keys: %w", err)
	}
	crypto := &SCMTokenCrypto{db: db}
	if _, err := crypto.ensureActiveKey(context.Background()); err != nil {
		return nil, err
	}
	return crypto, nil
}

func (crypto *SCMTokenCrypto) Encrypt(ctx context.Context, plaintext string) (string, error) {
	if crypto == nil || crypto.db == nil {
		return "", fmt.Errorf("scm token crypto is not initialized")
	}
	activeKey, err := crypto.ensureActiveKey(ctx)
	if err != nil {
		return "", err
	}
	dataKey, err := crypto.decodeDataKey(activeKey.KeyMaterial)
	if err != nil {
		return "", err
	}
	sealedValue, err := sealWithKey(dataKey, plaintext)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s|%s|%s", scmTokenEnvelopeVersion, activeKey.KeyID, sealedValue), nil
}

func (crypto *SCMTokenCrypto) Decrypt(ctx context.Context, envelope string) (string, error) {
	if crypto == nil || crypto.db == nil {
		return "", fmt.Errorf("scm token crypto is not initialized")
	}
	trimmedEnvelope := strings.TrimSpace(envelope)
	parts := strings.SplitN(trimmedEnvelope, "|", 3)
	if len(parts) != 3 || parts[0] != scmTokenEnvelopeVersion {
		return trimmedEnvelope, nil
	}
	keyID := strings.TrimSpace(parts[1])
	if keyID == "" {
		return "", fmt.Errorf("scm token envelope key id is required")
	}
	var keyRecord scmTokenKeyRecord
	if err := crypto.db.WithContext(ctx).Model(&scmTokenKeyRecord{}).Where("key_id = ?", keyID).Take(&keyRecord).Error; err != nil {
		return "", fmt.Errorf("load scm token key %q: %w", keyID, err)
	}
	dataKey, err := crypto.decodeDataKey(keyRecord.KeyMaterial)
	if err != nil {
		return "", err
	}
	return openWithKey(dataKey, parts[2])
}

func (crypto *SCMTokenCrypto) RotateAndReencryptSCMTokens(ctx context.Context) error {
	if crypto == nil || crypto.db == nil {
		return fmt.Errorf("scm token crypto is not initialized")
	}
	newKey, err := crypto.rotateActiveKey(ctx)
	if err != nil {
		return err
	}
	var records []projectSCMRecord
	if err := crypto.db.WithContext(ctx).Model(&projectSCMRecord{}).Find(&records).Error; err != nil {
		return fmt.Errorf("list scm token records for re-encryption: %w", err)
	}
	for _, record := range records {
		plaintext, err := crypto.Decrypt(ctx, record.SCMToken)
		if err != nil {
			return fmt.Errorf("decrypt scm token for scm_id %q: %w", record.SCMID, err)
		}
		encryptedValue, err := crypto.encryptWithKey(ctx, newKey, plaintext)
		if err != nil {
			return fmt.Errorf("encrypt scm token for scm_id %q: %w", record.SCMID, err)
		}
		if err := crypto.db.WithContext(ctx).
			Model(&projectSCMRecord{}).
			Where("id = ?", record.ID).
			Update("scm_token", encryptedValue).Error; err != nil {
			return fmt.Errorf("persist re-encrypted scm token for scm_id %q: %w", record.SCMID, err)
		}
	}
	return nil
}

func (crypto *SCMTokenCrypto) encryptWithKey(ctx context.Context, keyRecord scmTokenKeyRecord, plaintext string) (string, error) {
	dataKey, err := crypto.decodeDataKey(keyRecord.KeyMaterial)
	if err != nil {
		return "", err
	}
	sealedValue, err := sealWithKey(dataKey, plaintext)
	if err != nil {
		return "", err
	}
	_ = ctx
	return fmt.Sprintf("%s|%s|%s", scmTokenEnvelopeVersion, keyRecord.KeyID, sealedValue), nil
}

func (crypto *SCMTokenCrypto) ensureActiveKey(ctx context.Context) (scmTokenKeyRecord, error) {
	if crypto == nil || crypto.db == nil {
		return scmTokenKeyRecord{}, fmt.Errorf("scm token crypto is not initialized")
	}
	var activeKey scmTokenKeyRecord
	err := crypto.db.WithContext(ctx).Model(&scmTokenKeyRecord{}).Where("status = ?", keyStatusActive).Order("updated_at DESC").Take(&activeKey).Error
	if err == nil {
		return activeKey, nil
	}
	if err != gorm.ErrRecordNotFound {
		return scmTokenKeyRecord{}, fmt.Errorf("load active scm token key: %w", err)
	}
	return crypto.rotateActiveKey(ctx)
}

func (crypto *SCMTokenCrypto) rotateActiveKey(ctx context.Context) (scmTokenKeyRecord, error) {
	if crypto == nil || crypto.db == nil {
		return scmTokenKeyRecord{}, fmt.Errorf("scm token crypto is not initialized")
	}
	newKeyMaterial := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, newKeyMaterial); err != nil {
		return scmTokenKeyRecord{}, fmt.Errorf("generate scm token data key: %w", err)
	}
	newKey := scmTokenKeyRecord{KeyID: fmt.Sprintf("key_%d", time.Now().UTC().UnixNano()), Status: keyStatusActive, KeyMaterial: base64.StdEncoding.EncodeToString(newKeyMaterial)}
	txErr := crypto.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&scmTokenKeyRecord{}).Where("status = ?", keyStatusActive).Update("status", keyStatusRetired).Error; err != nil {
			return fmt.Errorf("retire active scm token key: %w", err)
		}
		if err := tx.Create(&newKey).Error; err != nil {
			return fmt.Errorf("persist new scm token key: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return scmTokenKeyRecord{}, txErr
	}
	return newKey, nil
}

func (crypto *SCMTokenCrypto) decodeDataKey(encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("decode scm token data key: %w", err)
	}
	if len(decoded) != 32 {
		return nil, fmt.Errorf("invalid scm token data key length")
	}
	return decoded, nil
}

func sealWithKey(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("init cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("init gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("%s|%s", base64.StdEncoding.EncodeToString(nonce), base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func openWithKey(key []byte, payload string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(payload), "|", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted payload format")
	}
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("init cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("init gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt payload: %w", err)
	}
	return string(plaintext), nil
}

func isEncryptedSCMTokenEnvelope(value string) bool {
	parts := strings.SplitN(strings.TrimSpace(value), "|", 3)
	return len(parts) == 3 && parts[0] == scmTokenEnvelopeVersion
}

func (crypto *SCMTokenCrypto) MigrateLegacyPlaintextSCMTokens(ctx context.Context) error {
	if crypto == nil || crypto.db == nil {
		return fmt.Errorf("scm token crypto is not initialized")
	}
	activeKey, err := crypto.ensureActiveKey(ctx)
	if err != nil {
		return err
	}
	var records []projectSCMRecord
	if err := crypto.db.WithContext(ctx).Model(&projectSCMRecord{}).Find(&records).Error; err != nil {
		return fmt.Errorf("list scm token records for migration: %w", err)
	}
	for _, record := range records {
		trimmedToken := strings.TrimSpace(record.SCMToken)
		if isEncryptedSCMTokenEnvelope(trimmedToken) {
			continue
		}
		encryptedValue, encryptErr := crypto.encryptWithKey(ctx, activeKey, trimmedToken)
		if encryptErr != nil {
			return fmt.Errorf("encrypt legacy scm token for scm_id %q: %w", record.SCMID, encryptErr)
		}
		if err := crypto.db.WithContext(ctx).
			Model(&projectSCMRecord{}).
			Where("id = ?", record.ID).
			Update("scm_token", encryptedValue).Error; err != nil {
			return fmt.Errorf("persist migrated scm token for scm_id %q: %w", record.SCMID, err)
		}
	}
	return nil
}
