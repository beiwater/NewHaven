package memory_test

import (
	"context"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	"github.com/beiwater/NewHaven/backend/internal/domain/social"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func TestSnapshotRestoreAdvancesPersistedRecordIDs(t *testing.T) {
	ctx := context.Background()
	original := memory.New()
	ledger := &finance.LedgerEntry{CompanyID: 1, Kind: "test", Amount: 10}
	message := &social.Message{CompanyID: 1, Channel: "global", Content: "before restart"}
	notification := &social.Notification{CompanyID: 1, Kind: "test", Message: "before restart"}
	if err := original.AppendLedgerEntry(ctx, ledger); err != nil {
		t.Fatal(err)
	}
	if err := original.SaveMessage(ctx, message); err != nil {
		t.Fatal(err)
	}
	if err := original.CreateNotification(ctx, notification); err != nil {
		t.Fatal(err)
	}

	restored := memory.New()
	restored.LoadFromSnapshot(original.GetSnapshotData())
	newLedger := &finance.LedgerEntry{CompanyID: 1, Kind: "test", Amount: 20}
	newMessage := &social.Message{CompanyID: 1, Channel: "global", Content: "after restart"}
	newNotification := &social.Notification{CompanyID: 1, Kind: "test", Message: "after restart"}
	if err := restored.AppendLedgerEntry(ctx, newLedger); err != nil {
		t.Fatal(err)
	}
	if err := restored.SaveMessage(ctx, newMessage); err != nil {
		t.Fatal(err)
	}
	if err := restored.CreateNotification(ctx, newNotification); err != nil {
		t.Fatal(err)
	}

	if newLedger.ID != ledger.ID+1 {
		t.Fatalf("ledger ID after restore = %d, want %d", newLedger.ID, ledger.ID+1)
	}
	if newMessage.ID != message.ID+1 {
		t.Fatalf("message ID after restore = %d, want %d", newMessage.ID, message.ID+1)
	}
	if newNotification.ID != notification.ID+1 {
		t.Fatalf("notification ID after restore = %d, want %d", newNotification.ID, notification.ID+1)
	}
}
