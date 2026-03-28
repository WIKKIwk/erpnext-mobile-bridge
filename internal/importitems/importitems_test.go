package importitems

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mobile_server/internal/erpnext"
)

type stubERP struct {
	customers     []erpnext.Customer
	itemsByCode   map[string]erpnext.Item
	createdInputs []erpnext.CreateItemInput
	assignments   [][2]string
}

func (s *stubERP) SearchCustomers(ctx context.Context, baseURL, apiKey, apiSecret, query string, limit int) ([]erpnext.Customer, error) {
	return s.customers, nil
}

func (s *stubERP) GetItemsByCodes(ctx context.Context, baseURL, apiKey, apiSecret string, itemCodes []string) ([]erpnext.Item, error) {
	result := make([]erpnext.Item, 0, len(itemCodes))
	for _, code := range itemCodes {
		if item, ok := s.itemsByCode[strings.TrimSpace(code)]; ok {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *stubERP) CreateItem(ctx context.Context, baseURL, apiKey, apiSecret string, input erpnext.CreateItemInput) (erpnext.Item, error) {
	s.createdInputs = append(s.createdInputs, input)
	item := erpnext.Item{
		Code: input.Code,
		Name: input.Name,
		UOM:  input.UOM,
	}
	if s.itemsByCode == nil {
		s.itemsByCode = map[string]erpnext.Item{}
	}
	s.itemsByCode[input.Code] = item
	return item, nil
}

func (s *stubERP) AssignCustomerToItem(ctx context.Context, baseURL, apiKey, apiSecret, itemCode, customerRef string) error {
	s.assignments = append(s.assignments, [2]string{itemCode, customerRef})
	return nil
}

func TestRunImportsCSVItemsAndAssignsCustomer(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "items.csv")
	if err := os.WriteFile(csvPath, []byte("name\nPista\nMakiz Pasta\nPista\n"), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	stub := &stubERP{
		customers: []erpnext.Customer{{ID: "CUST-001", Name: "Makiz"}},
		itemsByCode: map[string]erpnext.Item{
			"Pista": {Code: "Pista", Name: "Pista", UOM: "Kg"},
		},
	}

	result, err := Run(context.Background(), stub, nil, Options{
		CSVPath:   csvPath,
		Customer:  "Makiz",
		UOM:       "Kg",
		ItemGroup: "Tayyor mahsulot",
		BaseURL:   "http://erp.test",
		APIKey:    "key",
		APISecret: "secret",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ItemsFound != 2 {
		t.Fatalf("expected 2 deduped items, got %+v", result)
	}
	if len(result.Existing) != 1 || result.Existing[0] != "Pista" {
		t.Fatalf("expected existing Pista, got %+v", result.Existing)
	}
	if len(stub.createdInputs) != 1 {
		t.Fatalf("expected 1 created item, got %+v", stub.createdInputs)
	}
	if stub.createdInputs[0].ItemGroup != "Tayyor mahsulot" {
		t.Fatalf("expected item group Tayyor mahsulot, got %+v", stub.createdInputs[0])
	}
	if len(stub.assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %+v", stub.assignments)
	}
}

func TestLoadNamesFromCSVDetectsSemicolonHeader(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "items.csv")
	if err := os.WriteFile(csvPath, []byte("mahsulot;izoh\nIsko 16sm;test\nCheers nachos;test\n"), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	names, column, rowsRead, err := loadNamesFromCSV(csvPath, "")
	if err != nil {
		t.Fatalf("loadNamesFromCSV() error = %v", err)
	}
	if column != 0 || rowsRead != 2 {
		t.Fatalf("unexpected column/rows: column=%d rows=%d", column, rowsRead)
	}
	if len(names) != 2 || names[0] != "Isko 16sm" || names[1] != "Cheers nachos" {
		t.Fatalf("unexpected names: %+v", names)
	}
}
