package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
)

// DBTrip will perform the request to a database and compose the http response
func DBTrip(request *http.Request, rule *Rule) (*http.Response, error) {
	// the request body contains the query
	queryData, _ := io.ReadAll(request.Body)
	query := string(queryData)
	// performing the query
	rows, err := rule.db.Query(query)
	if err != nil {
		return nil, err
	}
	// turning the rows into JSON
	body, err := toJSON(rows)
	if err != nil {
		return nil, err
	}
	response := http.Response{StatusCode: 200, Request: request}
	response.Header = http.Header{}

	response.Header.Set("content-type", "application/json")
	response.Body = io.NopCloser(bytes.NewReader(body))
	return &response, nil
}

// toJSON will convert database rows into JSON
func toJSON(rows *sql.Rows) ([]byte, error) {
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	count := len(columnTypes)
	finalRows := make([]any, 0)
	for rows.Next() {
		scanArgs := make([]any, count)
		for i, v := range columnTypes {
			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
				scanArgs[i] = new(sql.NullString)
				break
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
				break
			case "INT4":
				scanArgs[i] = new(sql.NullInt64)
				break
			case "INT":
				scanArgs[i] = new(sql.NullInt64)
				break
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		masterData := map[string]any{}
		for i, v := range columnTypes {
			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				masterData[v.Name()] = z.Bool
				continue
			}
			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				masterData[v.Name()] = z.String
				continue
			}
			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				masterData[v.Name()] = z.Int64
				continue
			}
			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}
			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				masterData[v.Name()] = z.Int32
				continue
			}
			masterData[v.Name()] = scanArgs[i]
		}
		finalRows = append(finalRows, masterData)
	}
	return json.Marshal(finalRows)
}
