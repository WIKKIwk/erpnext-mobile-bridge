package erpnext

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func (c *Client) SearchCustomers(ctx context.Context, baseURL, apiKey, apiSecret, query string, limit int) ([]Customer, error) {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 500 {
		limit = 500
	}

	filtersJSON, _ := json.Marshal([][]interface{}{
		{"disabled", "=", 0},
	})

	params := url.Values{}
	params.Set("fields", `["name","customer_name","mobile_no"]`)
	params.Set("filters", string(filtersJSON))
	params.Set("limit_page_length", strconv.Itoa(limit))
	params.Set("order_by", "modified desc")

	if trimmed := strings.TrimSpace(query); trimmed != "" {
		like := "%" + strings.ReplaceAll(trimmed, "\"", "") + "%"
		orFiltersJSON, _ := json.Marshal([][]interface{}{
			{"name", "like", like},
			{"customer_name", "like", like},
			{"mobile_no", "like", like},
		})
		params.Set("or_filters", string(orFiltersJSON))
	}

	var payload struct {
		Data []struct {
			Name         string `json:"name"`
			CustomerName string `json:"customer_name"`
			MobileNo     string `json:"mobile_no"`
		} `json:"data"`
	}

	endpoint := normalized + "/api/resource/Customer?" + params.Encode()
	if err := c.doJSON(ctx, endpoint, apiKey, apiSecret, &payload); err != nil {
		return nil, err
	}

	items := make([]Customer, 0, len(payload.Data))
	for _, row := range payload.Data {
		name := strings.TrimSpace(row.CustomerName)
		if name == "" {
			name = strings.TrimSpace(row.Name)
		}
		items = append(items, Customer{
			ID:    strings.TrimSpace(row.Name),
			Name:  name,
			Phone: strings.TrimSpace(row.MobileNo),
		})
	}
	return items, nil
}

func (c *Client) GetCustomer(ctx context.Context, baseURL, apiKey, apiSecret, id string) (Customer, error) {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return Customer{}, err
	}

	endpoint := normalized + "/api/resource/Customer/" + url.PathEscape(strings.TrimSpace(id))
	var payload struct {
		Data struct {
			Name         string `json:"name"`
			CustomerName string `json:"customer_name"`
			MobileNo     string `json:"mobile_no"`
		} `json:"data"`
	}
	if err := c.doJSON(ctx, endpoint, apiKey, apiSecret, &payload); err != nil {
		return Customer{}, err
	}
	name := strings.TrimSpace(payload.Data.CustomerName)
	if name == "" {
		name = strings.TrimSpace(payload.Data.Name)
	}
	return Customer{
		ID:    strings.TrimSpace(payload.Data.Name),
		Name:  name,
		Phone: strings.TrimSpace(payload.Data.MobileNo),
	}, nil
}

func (c *Client) ListCustomerItems(ctx context.Context, baseURL, apiKey, apiSecret, customerRef, query string, limit int) ([]Item, error) {
	normalized, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	filtersJSON, _ := json.Marshal([][]interface{}{
		{"party_type", "=", "Customer"},
		{"party", "=", strings.TrimSpace(customerRef)},
		{"restrict_based_on", "=", "Item"},
	})

	params := url.Values{}
	params.Set("fields", `["name","party","party_type","restrict_based_on","based_on_value"]`)
	params.Set("filters", string(filtersJSON))
	params.Set("limit_page_length", strconv.Itoa(limit))

	var payload struct {
		Data []struct {
			BasedOnValue string `json:"based_on_value"`
		} `json:"data"`
	}
	endpoint := normalized + "/api/resource/Party%20Specific%20Item?" + params.Encode()
	if err := c.doJSON(ctx, endpoint, apiKey, apiSecret, &payload); err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(payload.Data))
	seen := make(map[string]struct{}, len(payload.Data))
	for _, row := range payload.Data {
		code := strings.TrimSpace(row.BasedOnValue)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return []Item{}, nil
	}

	items, err := c.GetItemsByCodes(ctx, normalized, apiKey, apiSecret, codes)
	if err != nil {
		return nil, err
	}

	trimmedQuery := strings.ToLower(strings.TrimSpace(query))
	if trimmedQuery != "" {
		filtered := make([]Item, 0, len(items))
		for _, item := range items {
			if strings.Contains(strings.ToLower(strings.TrimSpace(item.Code)), trimmedQuery) ||
				strings.Contains(strings.ToLower(strings.TrimSpace(item.Name)), trimmedQuery) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}
