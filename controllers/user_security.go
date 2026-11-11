package controllers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Explicit database column names preserve the legacy camel-case blob columns.
func getUserSecurityUpdateColumns(body []byte, columnsStr string) ([]string, error) {
	fields := map[string]string{
		"faceIds":             "face_ids",
		"mfaAccounts":         "mfaAccounts",
		"webauthnCredentials": "webauthnCredentials",
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("auth:Unauthorized operation")
	}

	columns := make([]string, 0, len(payload))
	for field := range payload {
		column, ok := fields[field]
		if !ok {
			return nil, fmt.Errorf("auth:Unauthorized operation")
		}
		columns = append(columns, column)
	}
	if columnsStr != "" {
		for _, field := range strings.Split(columnsStr, ",") {
			column := fields[field]
			if column == "" {
				for _, allowedColumn := range fields {
					if field == allowedColumn {
						column = allowedColumn
						break
					}
				}
			}
			if !slices.Contains(columns, column) {
				return nil, fmt.Errorf("auth:Unauthorized operation")
			}
		}
	}
	return columns, nil
}
