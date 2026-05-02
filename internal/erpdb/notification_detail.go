package erpdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"

	"mobile_server/internal/core"
)

var notificationCommentTagPattern = regexp.MustCompile(`<[^>]+>`)

func (r *Reader) NotificationDetailByReceiptID(ctx context.Context, receiptID string) (core.NotificationDetail, error) {
	targetName, targetType, eventType, err := resolveNotificationTargetReceiptID(receiptID)
	if err != nil {
		return core.NotificationDetail{}, err
	}

	switch targetType {
	case "delivery_note":
		detail, err := r.deliveryNoteNotificationDetail(ctx, targetName, receiptID)
		if err != nil {
			return core.NotificationDetail{}, err
		}
		return detail, nil
	default:
		detail, err := r.purchaseReceiptNotificationDetail(ctx, targetName, receiptID, eventType)
		if err != nil {
			return core.NotificationDetail{}, err
		}
		return detail, nil
	}
}

func resolveNotificationTargetReceiptID(receiptID string) (targetName, targetType, eventType string, err error) {
	trimmedReceiptID := strings.TrimSpace(receiptID)
	if strings.HasPrefix(trimmedReceiptID, supplierAckEventPrefixDB) {
		eventType = "supplier_ack"
		parts := strings.SplitN(strings.TrimPrefix(trimmedReceiptID, supplierAckEventPrefixDB), ":", 2)
		if len(parts) > 0 {
			trimmedReceiptID = strings.TrimSpace(parts[0])
		}
	}
	if strings.HasPrefix(trimmedReceiptID, "customer_delivery_result:") {
		parts := strings.SplitN(strings.TrimPrefix(trimmedReceiptID, "customer_delivery_result:"), ":", 2)
		if len(parts) > 0 {
			targetName = strings.TrimSpace(parts[0])
		}
		if targetName == "" {
			return "", "", "", fmt.Errorf("delivery note id is required")
		}
		return targetName, "delivery_note", eventType, nil
	}
	if trimmedReceiptID == "" {
		return "", "", "", fmt.Errorf("receipt id is required")
	}
	return trimmedReceiptID, "purchase_receipt", eventType, nil
}

func (r *Reader) purchaseReceiptNotificationDetail(ctx context.Context, name, receiptID, eventType string) (core.NotificationDetail, error) {
	row, err := r.purchaseReceiptRowByName(ctx, name)
	if err != nil {
		return core.NotificationDetail{}, err
	}

	record := purchaseReceiptRowToDispatchRecord(row)
	if eventType == "supplier_ack" {
		record.ID = strings.TrimSpace(receiptID)
		record.EventType = eventType
		record.Highlight = "Supplier mahsulotni qaytarganingizni tasdiqladi"
	}

	comments, err := r.notificationComments(ctx, "Purchase Receipt", name, 100)
	if err != nil {
		return core.NotificationDetail{}, err
	}

	return core.NotificationDetail{
		Record:   record,
		Comments: comments,
	}, nil
}

func (r *Reader) deliveryNoteNotificationDetail(ctx context.Context, name, receiptID string) (core.NotificationDetail, error) {
	row, err := r.deliveryNoteRowByName(ctx, name)
	if err != nil {
		return core.NotificationDetail{}, err
	}

	record, ok := buildCustomerResultDispatch(row)
	if !ok {
		record = deliveryNoteRowToDispatchRecord(row)
		record.ID = strings.TrimSpace(receiptID)
	} else if strings.TrimSpace(receiptID) != "" {
		record.ID = strings.TrimSpace(receiptID)
	}

	comments, err := r.notificationComments(ctx, "Delivery Note", name, 100)
	if err != nil {
		return core.NotificationDetail{}, err
	}

	return core.NotificationDetail{
		Record:   record,
		Comments: comments,
	}, nil
}

func (r *Reader) purchaseReceiptRowByName(ctx context.Context, name string) (purchaseReceiptSummaryRow, error) {
	var row purchaseReceiptSummaryRow
	err := r.db.QueryRowContext(ctx, `
		SELECT
			pr.name,
			pr.supplier,
			COALESCE(pr.supplier_name, ''),
			pr.docstatus,
			COALESCE(pr.status, ''),
			COALESCE(pr.total_qty, 0),
			COALESCE(CAST(pr.posting_date AS CHAR), ''),
			COALESCE(pr.supplier_delivery_note, ''),
			COALESCE(pr.remarks, ''),
			COALESCE(pr.currency, ''),
			COALESCE(pri.item_code, ''),
			COALESCE(pri.item_name, ''),
			COALESCE(pri.uom, ''),
			COALESCE(pri.amount, 0)
		FROM `+"`tabPurchase Receipt`"+` pr
		LEFT JOIN `+"`tabPurchase Receipt Item`"+` pri ON pri.parent = pr.name AND pri.idx = 1
		WHERE pr.name = ?
		LIMIT 1`,
		strings.TrimSpace(name),
	).Scan(
		&row.Name,
		&row.Supplier,
		&row.SupplierName,
		&row.DocStatus,
		&row.Status,
		&row.TotalQty,
		&row.PostingDate,
		&row.SupplierDeliveryNote,
		&row.Remarks,
		&row.Currency,
		&row.ItemCode,
		&row.ItemName,
		&row.UOM,
		&row.Amount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return purchaseReceiptSummaryRow{}, fmt.Errorf("purchase receipt not found")
		}
		return purchaseReceiptSummaryRow{}, err
	}
	return row, nil
}

func (r *Reader) deliveryNoteRowByName(ctx context.Context, name string) (deliveryNoteSummaryRow, error) {
	var row deliveryNoteSummaryRow
	err := r.db.QueryRowContext(ctx, `
		SELECT
			dn.name,
			dn.customer,
			COALESCE(dn.customer_name, ''),
			dn.docstatus,
			COALESCE(CAST(dn.modified AS CHAR), ''),
			COALESCE(dn.total_qty, 0),
			COALESCE(dni.returned_qty, 0),
			COALESCE(dn.accord_customer_reason, ''),
			COALESCE(dni.item_code, ''),
			COALESCE(dni.item_name, ''),
			COALESCE(dni.uom, ''),
			COALESCE(dn.accord_flow_state, 0),
			COALESCE(dn.accord_customer_state, 0)
		FROM `+"`tabDelivery Note`"+` dn
		LEFT JOIN `+"`tabDelivery Note Item`"+` dni ON dni.parent = dn.name AND dni.idx = 1
		WHERE dn.name = ?
		LIMIT 1`,
		strings.TrimSpace(name),
	).Scan(
		&row.Name,
		&row.Customer,
		&row.CustomerName,
		&row.DocStatus,
		&row.Modified,
		&row.Qty,
		&row.ReturnedQty,
		&row.CustomerReason,
		&row.ItemCode,
		&row.ItemName,
		&row.UOM,
		&row.AccordFlowState,
		&row.AccordCustomerState,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return deliveryNoteSummaryRow{}, fmt.Errorf("delivery note not found")
		}
		return deliveryNoteSummaryRow{}, err
	}
	return row, nil
}

func (r *Reader) notificationComments(ctx context.Context, doctype, name string, limit int) ([]core.NotificationComment, error) {
	limit = clampLimit(limit, 0, 200)
	query := `
		SELECT
			c.name,
			COALESCE(CAST(c.creation AS CHAR), ''),
			COALESCE(c.content, '')
		FROM ` + "`tabComment`" + ` c
		WHERE c.reference_doctype = ?
		  AND c.reference_name = ?
		ORDER BY c.creation ASC, c.name ASC`
	if limit > 0 {
		query += "\n\t\tLIMIT ?"
	}
	args := []interface{}{doctype, strings.TrimSpace(name)}
	if limit > 0 {
		args = append(args, limit)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]core.NotificationComment, 0, limit)
	for rows.Next() {
		var (
			id           string
			createdLabel string
			content      string
		)
		if err := rows.Scan(&id, &createdLabel, &content); err != nil {
			return nil, err
		}
		authorLabel, body := parseNotificationCommentContent(content)
		if body == "" {
			continue
		}
		result = append(result, core.NotificationComment{
			ID:           strings.TrimSpace(id),
			AuthorLabel:  authorLabel,
			Body:         body,
			CreatedLabel: strings.TrimSpace(createdLabel),
		})
	}
	return result, rows.Err()
}

func parseNotificationCommentContent(content string) (string, string) {
	trimmed := sanitizeNotificationCommentContent(content)
	if trimmed == "" {
		return "", ""
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) >= 2 {
		head := strings.TrimSpace(lines[0])
		body := strings.TrimSpace(strings.Join(lines[1:], "\n"))
		if body != "" &&
			(strings.HasPrefix(head, "Supplier") ||
				strings.HasPrefix(head, "Werka") ||
				strings.HasPrefix(head, "Customer") ||
				strings.HasPrefix(head, "Admin")) {
			return head, body
		}
	}
	return "Tizim", trimmed
}

func sanitizeNotificationCommentContent(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}
	replaced := strings.ReplaceAll(trimmed, "<br>", "\n")
	replaced = strings.ReplaceAll(replaced, "<br/>", "\n")
	replaced = strings.ReplaceAll(replaced, "<br />", "\n")
	replaced = notificationCommentTagPattern.ReplaceAllString(replaced, "")
	replaced = html.UnescapeString(replaced)
	lines := strings.Split(strings.ReplaceAll(replaced, "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		if cleaned == "" {
			continue
		}
		filtered = append(filtered, cleaned)
	}
	return strings.Join(filtered, "\n")
}

