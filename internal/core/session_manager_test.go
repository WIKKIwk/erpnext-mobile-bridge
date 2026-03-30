package core

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPersistentSessionManagerPersistsAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")

	manager := NewPersistentSessionManager(path, 24*time.Hour)
	token, err := manager.Create(Principal{
		Role:        RoleSupplier,
		DisplayName: "Abdulloh",
		Ref:         "SUP-001",
		Phone:       "+998900001122",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	restarted := NewPersistentSessionManager(path, 24*time.Hour)
	principal, ok := restarted.Get(token)
	if !ok {
		t.Fatal("expected persisted session after restart")
	}
	if principal.Ref != "SUP-001" {
		t.Fatalf("expected supplier ref to persist, got %+v", principal)
	}

	restarted.Delete(token)

	restartedAgain := NewPersistentSessionManager(path, 24*time.Hour)
	if _, ok := restartedAgain.Get(token); ok {
		t.Fatal("expected logout to remove persisted session")
	}
}

func TestPersistentSessionManagerExpiresSessions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")

	manager := NewPersistentSessionManager(path, 10*time.Millisecond)
	token, err := manager.Create(Principal{
		Role:        RoleAdmin,
		DisplayName: "Admin",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	time.Sleep(25 * time.Millisecond)

	restarted := NewPersistentSessionManager(path, 10*time.Millisecond)
	if _, ok := restarted.Get(token); ok {
		t.Fatal("expected expired session to be rejected")
	}
}
