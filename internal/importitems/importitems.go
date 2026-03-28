package importitems

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"mobile_server/internal/erpnext"
)

type ERP interface {
	SearchCustomers(ctx context.Context, baseURL, apiKey, apiSecret, query string, limit int) ([]erpnext.Customer, error)
	GetItemsByCodes(ctx context.Context, baseURL, apiKey, apiSecret string, itemCodes []string) ([]erpnext.Item, error)
	CreateItem(ctx context.Context, baseURL, apiKey, apiSecret string, input erpnext.CreateItemInput) (erpnext.Item, error)
	AssignCustomerToItem(ctx context.Context, baseURL, apiKey, apiSecret, itemCode, customerRef string) error
}

type Options struct {
	CSVPath   string
	Customer  string
	Column    string
	UOM       string
	ItemGroup string
	DryRun    bool
	BaseURL   string
	APIKey    string
	APISecret string
}

type Result struct {
	CustomerID   string
	CustomerName string
	DetectedCol  int
	RowsRead     int
	ItemsFound   int
	Created      []string
	Existing     []string
	Assigned     []string
}

func Run(ctx context.Context, erp ERP, out io.Writer, opts Options) (Result, error) {
	if strings.TrimSpace(opts.CSVPath) == "" {
		return Result{}, fmt.Errorf("csv path is required")
	}
	if strings.TrimSpace(opts.Customer) == "" {
		return Result{}, fmt.Errorf("customer is required")
	}
	if strings.TrimSpace(opts.BaseURL) == "" || strings.TrimSpace(opts.APIKey) == "" || strings.TrimSpace(opts.APISecret) == "" {
		return Result{}, fmt.Errorf("erp credentials are required")
	}
	if strings.TrimSpace(opts.UOM) == "" {
		opts.UOM = "Kg"
	}
	if strings.TrimSpace(opts.ItemGroup) == "" {
		opts.ItemGroup = "Tayyor mahsulot"
	}

	customer, err := resolveCustomer(ctx, erp, opts)
	if err != nil {
		return Result{}, err
	}

	names, detectedCol, rowsRead, err := loadNamesFromCSV(opts.CSVPath, opts.Column)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		DetectedCol:  detectedCol,
		RowsRead:     rowsRead,
		ItemsFound:   len(names),
	}

	for _, name := range names {
		itemCode := strings.TrimSpace(name)
		existing, err := erp.GetItemsByCodes(ctx, opts.BaseURL, opts.APIKey, opts.APISecret, []string{itemCode})
		if err != nil {
			return result, err
		}
		item := erpnext.Item{
			Code: itemCode,
			Name: name,
			UOM:  opts.UOM,
		}
		if len(existing) > 0 && strings.EqualFold(strings.TrimSpace(existing[0].Code), itemCode) {
			item = existing[0]
			result.Existing = append(result.Existing, item.Code)
		} else {
			if !opts.DryRun {
				item, err = erp.CreateItem(ctx, opts.BaseURL, opts.APIKey, opts.APISecret, erpnext.CreateItemInput{
					Code:      itemCode,
					Name:      name,
					UOM:       opts.UOM,
					ItemGroup: opts.ItemGroup,
				})
				if err != nil {
					return result, fmt.Errorf("create item %q: %w", name, err)
				}
			}
			result.Created = append(result.Created, itemCode)
		}

		if !opts.DryRun {
			if err := erp.AssignCustomerToItem(ctx, opts.BaseURL, opts.APIKey, opts.APISecret, item.Code, customer.ID); err != nil {
				return result, fmt.Errorf("assign customer %q to item %q: %w", customer.ID, item.Code, err)
			}
		}
		result.Assigned = append(result.Assigned, item.Code)
	}

	if out != nil {
		fmt.Fprintf(out, "customer: %s (%s)\n", result.CustomerName, result.CustomerID)
		fmt.Fprintf(out, "rows read: %d\n", result.RowsRead)
		fmt.Fprintf(out, "items found: %d\n", result.ItemsFound)
		fmt.Fprintf(out, "created: %d\n", len(result.Created))
		fmt.Fprintf(out, "existing: %d\n", len(result.Existing))
		fmt.Fprintf(out, "assigned: %d\n", len(result.Assigned))
	}
	return result, nil
}

func resolveCustomer(ctx context.Context, erp ERP, opts Options) (erpnext.Customer, error) {
	query := strings.TrimSpace(opts.Customer)
	customers, err := erp.SearchCustomers(ctx, opts.BaseURL, opts.APIKey, opts.APISecret, query, 100)
	if err != nil {
		return erpnext.Customer{}, err
	}

	var exact []erpnext.Customer
	for _, customer := range customers {
		if strings.EqualFold(strings.TrimSpace(customer.ID), query) ||
			strings.EqualFold(strings.TrimSpace(customer.Name), query) {
			exact = append(exact, customer)
		}
	}
	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) > 1 {
		return erpnext.Customer{}, fmt.Errorf("multiple customers matched %q", query)
	}
	return erpnext.Customer{}, fmt.Errorf("customer %q not found", query)
}

func loadNamesFromCSV(path, column string) ([]string, int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, -1, 0, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, -1, 0, err
	}
	comma := ','
	firstLine := string(data)
	if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
		firstLine = firstLine[:idx]
	}
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		comma = ';'
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = comma
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, -1, 0, err
	}
	if len(rows) == 0 {
		return nil, -1, 0, fmt.Errorf("csv is empty")
	}

	colIndex, skipFirst, err := detectColumn(rows, column)
	if err != nil {
		return nil, -1, 0, err
	}
	seen := map[string]struct{}{}
	names := make([]string, 0, len(rows))
	rowsRead := 0
	for index, row := range rows {
		if skipFirst && index == 0 {
			continue
		}
		if colIndex >= len(row) {
			continue
		}
		value := strings.TrimSpace(row[colIndex])
		if value == "" {
			continue
		}
		rowsRead++
		if _, ok := seen[strings.ToLower(value)]; ok {
			continue
		}
		seen[strings.ToLower(value)] = struct{}{}
		names = append(names, value)
	}
	if len(names) == 0 {
		return nil, colIndex, rowsRead, fmt.Errorf("csv produced no item names")
	}
	return names, colIndex, rowsRead, nil
}

func detectColumn(rows [][]string, requested string) (index int, skipFirst bool, err error) {
	header := rows[0]
	if trimmed := strings.TrimSpace(requested); trimmed != "" {
		if idx, convErr := strconv.Atoi(trimmed); convErr == nil {
			if idx < 0 || idx >= len(header) {
				return -1, false, fmt.Errorf("column index %d out of range", idx)
			}
			return idx, false, nil
		}
		for i, value := range header {
			if strings.EqualFold(strings.TrimSpace(value), trimmed) {
				return i, true, nil
			}
		}
		return -1, false, fmt.Errorf("column %q not found", trimmed)
	}

	headerNames := []string{"name", "item_name", "mahsulot", "product", "item"}
	for i, value := range header {
		normalized := strings.ToLower(strings.TrimSpace(value))
		for _, candidate := range headerNames {
			if normalized == candidate {
				return i, true, nil
			}
		}
	}
	return 0, false, nil
}
