package postgres

import (
	"context"
	"strings"
	"testing"
)

func TestSCMTokenCryptoEncryptDecryptAndInitialKeyWrite(t *testing.T) {
	db := newTestDB(t)
	crypto, err := NewSCMTokenCrypto(db)
	if err != nil {
		t.Fatalf("new scm token crypto: %v", err)
	}

	var keyCount int64
	if err := db.Model(&scmTokenKeyRecord{}).Count(&keyCount).Error; err != nil {
		t.Fatalf("count key records: %v", err)
	}
	if keyCount != 1 {
		t.Fatalf("expected one generated active key, got %d", keyCount)
	}

	envelope, err := crypto.Encrypt(context.Background(), "ghp_secret_token")
	if err != nil {
		t.Fatalf("encrypt token: %v", err)
	}
	if !isEncryptedSCMTokenEnvelope(envelope) {
		t.Fatalf("expected encrypted token envelope, got %q", envelope)
	}

	plaintext, err := crypto.Decrypt(context.Background(), envelope)
	if err != nil {
		t.Fatalf("decrypt token: %v", err)
	}
	if plaintext != "ghp_secret_token" {
		t.Fatalf("expected decrypted plaintext to match input, got %q", plaintext)
	}
}

func TestSCMTokenCryptoMigratesLegacyPlaintextAndRotatesKeys(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&projectSCMRecord{}); err != nil {
		t.Fatalf("migrate project scm table: %v", err)
	}

	crypto, err := NewSCMTokenCrypto(db)
	if err != nil {
		t.Fatalf("new scm token crypto: %v", err)
	}

	legacy := projectSCMRecord{ProjectID: "project-1", SCMID: "scm-1", SCMProvider: "github", SCMToken: "legacy_plaintext"}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatalf("seed legacy scm token: %v", err)
	}

	if err := crypto.MigrateLegacyPlaintextSCMTokens(context.Background()); err != nil {
		t.Fatalf("migrate legacy tokens: %v", err)
	}

	var migrated projectSCMRecord
	if err := db.WithContext(context.Background()).Where("project_id = ? AND scm_id = ?", "project-1", "scm-1").Take(&migrated).Error; err != nil {
		t.Fatalf("load migrated scm row: %v", err)
	}
	if !isEncryptedSCMTokenEnvelope(migrated.SCMToken) {
		t.Fatalf("expected migrated token envelope, got %q", migrated.SCMToken)
	}

	beforeRotate := migrated.SCMToken
	if err := crypto.RotateAndReencryptSCMTokens(context.Background()); err != nil {
		t.Fatalf("rotate and re-encrypt tokens: %v", err)
	}

	if err := db.WithContext(context.Background()).Where("project_id = ? AND scm_id = ?", "project-1", "scm-1").Take(&migrated).Error; err != nil {
		t.Fatalf("load rotated scm row: %v", err)
	}
	if !isEncryptedSCMTokenEnvelope(migrated.SCMToken) {
		t.Fatalf("expected rotated token envelope, got %q", migrated.SCMToken)
	}
	if strings.TrimSpace(migrated.SCMToken) == strings.TrimSpace(beforeRotate) {
		t.Fatal("expected rotated token envelope to differ after key rotation")
	}

	decrypted, err := crypto.Decrypt(context.Background(), migrated.SCMToken)
	if err != nil {
		t.Fatalf("decrypt rotated token: %v", err)
	}
	if decrypted != "legacy_plaintext" {
		t.Fatalf("expected rotated token to preserve plaintext value, got %q", decrypted)
	}

	var activeCount int64
	if err := db.Model(&scmTokenKeyRecord{}).Where("status = ?", keyStatusActive).Count(&activeCount).Error; err != nil {
		t.Fatalf("count active keys: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly one active key after rotation, got %d", activeCount)
	}
	var retiredCount int64
	if err := db.Model(&scmTokenKeyRecord{}).Where("status = ?", keyStatusRetired).Count(&retiredCount).Error; err != nil {
		t.Fatalf("count retired keys: %v", err)
	}
	if retiredCount < 1 {
		t.Fatalf("expected at least one retired key after rotation, got %d", retiredCount)
	}
}

func TestSCMTokenCryptoDecryptLeavesLegacyPlaintextUntouched(t *testing.T) {
	crypto, err := NewSCMTokenCrypto(newTestDB(t))
	if err != nil {
		t.Fatalf("new scm token crypto: %v", err)
	}
	plaintext, err := crypto.Decrypt(context.Background(), "legacy_token")
	if err != nil {
		t.Fatalf("decrypt legacy plaintext: %v", err)
	}
	if plaintext != "legacy_token" {
		t.Fatalf("expected plaintext passthrough for legacy token, got %q", plaintext)
	}
}
