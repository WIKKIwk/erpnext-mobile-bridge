package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"mobile_server/internal/config"
	"mobile_server/internal/erpnext"
	"mobile_server/internal/importitems"
)

func main() {
	csvPath := flag.String("csv", "", "path to CSV file")
	customer := flag.String("customer", "", "existing customer name or id")
	column := flag.String("column", "", "csv column name or zero-based index (optional)")
	uom := flag.String("uom", "Kg", "stock uom for created items")
	itemGroup := flag.String("item-group", "Tayyor mahsulot", "ERPNext item group for created items")
	dryRun := flag.Bool("dry-run", false, "show planned changes without writing to ERPNext")
	erpURL := flag.String("erp-url", "", "override ERP URL")
	apiKey := flag.String("erp-api-key", "", "override ERP API key")
	apiSecret := flag.String("erp-api-secret", "", "override ERP API secret")
	flag.Parse()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	opts := importitems.Options{
		CSVPath:   *csvPath,
		Customer:  *customer,
		Column:    *column,
		UOM:       *uom,
		ItemGroup: *itemGroup,
		DryRun:    *dryRun,
		BaseURL:   firstNonEmpty(*erpURL, cfg.DefaultERPURL),
		APIKey:    firstNonEmpty(*apiKey, cfg.DefaultERPAPIKey),
		APISecret: firstNonEmpty(*apiSecret, cfg.DefaultERPAPISecret),
	}

	client := erpnext.NewClient(&http.Client{Timeout: cfg.RequestTimeout})
	result, err := importitems.Run(context.Background(), client, os.Stdout, opts)
	if err != nil {
		log.Fatalf("import failed: %v", err)
	}

	fmt.Printf("done: customer=%s (%s), created=%d, existing=%d, assigned=%d\n",
		result.CustomerName,
		result.CustomerID,
		len(result.Created),
		len(result.Existing),
		len(result.Assigned),
	)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
