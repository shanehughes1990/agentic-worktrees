package bootstrap

import "testing"

func TestValidateDatabaseDSNAcceptsPostgresSchemes(t *testing.T) {
	valid := []string{
		"postgres://user:pass@localhost:5432/app?sslmode=disable",
		"postgresql://user:pass@localhost:5432/app?sslmode=disable",
	}
	for _, dsn := range valid {
		if err := validateDatabaseDSN(dsn); err != nil {
			t.Fatalf("expected dsn %q to be valid: %v", dsn, err)
		}
	}
}

func TestValidateDatabaseDSNRejectsInvalidValue(t *testing.T) {
	if err := validateDatabaseDSN("not-a-dsn"); err == nil {
		t.Fatalf("expected invalid dsn to be rejected")
	}
}

func TestValidateDatabaseDSNRejectsNonPostgresScheme(t *testing.T) {
	if err := validateDatabaseDSN("mysql://user:pass@localhost:3306/app"); err == nil {
		t.Fatalf("expected non-postgres dsn to be rejected")
	}
}
