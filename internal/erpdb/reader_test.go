package erpdb

import (
	"os"
	"path/filepath"
	"testing"

	"mobile_server/internal/core"
)

func TestClampLimit(t *testing.T) {
	if got := clampLimit(0, 50, 500); got != 50 {
		t.Fatalf("expected fallback 50, got %d", got)
	}
	if got := clampLimit(900, 50, 500); got != 500 {
		t.Fatalf("expected max 500, got %d", got)
	}
	if got := clampLimit(120, 50, 500); got != 120 {
		t.Fatalf("expected explicit 120, got %d", got)
	}
}

func TestLikePatternEscapesWildcards(t *testing.T) {
	got := likePattern(`a%b_c\z`)
	want := `%a\%b\_c\\z%`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildSearchPatternsAddsLatinFallbackForCyrillic(t *testing.T) {
	items := []core.SupplierItem{
		{Code: "A001", Name: "Central Item"},
		{Code: "A002", Name: "A’lo Ta’m Kanada"},
		{Code: "A003", Name: "Another Item"},
	}
	got := rankSupplierItemsByQuery(items, "ало")
	if len(got) != 1 {
		t.Fatalf("expected one normalized match, got %d", len(got))
	}
	if got[0].Name != "A’lo Ta’m Kanada" {
		t.Fatalf("unexpected top ranked item: %+v", got[0])
	}
}

func TestRankCustomerItemOptionsByQueryPrefersItemMatches(t *testing.T) {
	options := []core.CustomerItemOption{
		{
			CustomerRef:  "cust-1",
			CustomerName: "Afsona Market",
			ItemCode:     "X001",
			ItemName:     "Random Product",
		},
		{
			CustomerRef:  "cust-2",
			CustomerName: "Other Customer",
			ItemCode:     "X002",
			ItemName:     "Afsona",
		},
	}

	got := rankCustomerItemOptionsByQuery(options, "афсона")
	if len(got) != 2 {
		t.Fatalf("expected two matches, got %d", len(got))
	}
	if got[0].ItemName != "Afsona" {
		t.Fatalf("expected direct item match first, got %+v", got[0])
	}
}

func TestRankSupplierItemSearchEntriesByQueryMatchesLinkedCustomerNames(t *testing.T) {
	items := []supplierItemSearchEntry{
		{
			item: core.SupplierItem{Code: "ITEM-001", Name: "Universal Item"},
			searchTerms: []string{
				"ITEM-001",
				"Universal Item",
				"Saidamin",
				"HotLunch",
			},
		},
		{
			item: core.SupplierItem{Code: "ITEM-002", Name: "Other Item"},
			searchTerms: []string{
				"ITEM-002",
				"Other Item",
				"Saidamin",
			},
		},
	}

	got := rankSupplierItemSearchEntriesByQuery(items, "hotlunch")
	if len(got) != 1 {
		t.Fatalf("expected one customer-linked match, got %d", len(got))
	}
	if got[0].Code != "ITEM-001" {
		t.Fatalf("unexpected matched item: %+v", got[0])
	}
}

func TestConfigFromSiteConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "site_config.json")
	if err := os.WriteFile(path, []byte(`{"db_name":"site_db","db_password":"secret","db_type":"mariadb"}`), 0o644); err != nil {
		t.Fatalf("write site_config: %v", err)
	}

	cfg, err := ConfigFromSiteConfig(path, "Stores - A")
	if err != nil {
		t.Fatalf("ConfigFromSiteConfig() error = %v", err)
	}
	if cfg.Name != "site_db" || cfg.User != "site_db" {
		t.Fatalf("unexpected db identity: %+v", cfg)
	}
	if cfg.Password != "secret" {
		t.Fatalf("expected password to be loaded")
	}
	if cfg.DefaultWarehouse != "Stores - A" {
		t.Fatalf("expected default warehouse to be preserved")
	}
}

func TestClassifyWerkaReceiptSkipsUnannouncedPending(t *testing.T) {
	status, include := classifyWerkaReceipt(purchaseReceiptSummaryRow{
		DocStatus: 0,
		Status:    "Draft",
		TotalQty:  1,
		Remarks:   "Accord Werka Aytilmagan: pending",
	})
	if status != "draft" {
		t.Fatalf("expected draft status, got %q", status)
	}
	if include {
		t.Fatalf("expected pending unannounced receipt to be skipped")
	}
}

func TestDeliveryStatusUsesAccordFields(t *testing.T) {
	status := deliveryStatus(deliveryNoteSummaryRow{
		DocStatus:           1,
		AccordFlowState:     deliveryFlowStateSubmittedDB,
		AccordCustomerState: customerStateConfirmedDB,
	})
	if status != "accepted" {
		t.Fatalf("expected accepted status, got %q", status)
	}
}
