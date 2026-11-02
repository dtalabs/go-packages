package d1

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type metadata struct {
	Changes   int64 `json:"changes"`
	LastRowID int64 `json:"last_row_id"`
}

func (m metadata) LastInsertId() (int64, error) { return m.LastRowID, nil }
func (m metadata) RowsAffected() (int64, error) { return m.Changes, nil }

type queryResult struct {
	Success bool     `json:"success"`
	Error   string   `json:"error"`
	Meta    metadata `json:"meta"`
	Results struct {
		Columns []string `json:"columns"`
		Rows    [][]any  `json:"rows"`
	} `json:"results"`
}

func (c *connection) execute(ctx context.Context, query string, args []driver.NamedValue) (*queryResult, error) {
	params := make([]any, len(args))
	for i, arg := range args {
		if arg.Name != "" {
			return nil, fmt.Errorf("d1: named parameters are unsupported; use positional ? placeholders")
		}
		switch v := arg.Value.(type) {
		case nil, string, int64, float64:
			params[i] = v
		case bool:
			if v {
				params[i] = 1
			} else {
				params[i] = 0
			}
		case time.Time:
			params[i] = v.Format("2006-01-02 15:04:05.999999999-07:00")
		case []byte:
			values := make([]int, len(v))
			for j, b := range v {
				values[j] = int(b)
			}
			params[i] = values
		default:
			return nil, fmt.Errorf("d1: unsupported parameter type %T", v)
		}
	}
	body, err := json.Marshal(struct {
		SQL    string `json:"sql"`
		Params []any  `json:"params"`
	}{query, params})
	if err != nil {
		return nil, fmt.Errorf("d1: encode query: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("d1: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("d1: HTTP %d", resp.StatusCode)
	}
	var envelope struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result []queryResult `json:"result"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err = dec.Decode(&envelope); err != nil {
		return nil, fmt.Errorf("d1: decode response: %w", err)
	}
	if !envelope.Success || len(envelope.Errors) > 0 {
		if len(envelope.Errors) > 0 {
			return nil, fmt.Errorf("d1: API error %d: %s", envelope.Errors[0].Code, envelope.Errors[0].Message)
		}
		return nil, fmt.Errorf("d1: query failed")
	}
	if len(envelope.Result) != 1 {
		return nil, fmt.Errorf("d1: expected one statement result, received %d; submit one SQL statement per call", len(envelope.Result))
	}
	r := &envelope.Result[0]
	if !r.Success {
		return nil, fmt.Errorf("d1: statement failed: %s", r.Error)
	}
	return r, nil
}

type rows struct {
	columns []string
	values  [][]any
	index   int
}

func (r *rows) Columns() []string { return r.columns }
func (r *rows) Close() error      { r.values = nil; return nil }
func (r *rows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	row := r.values[r.index]
	r.index++
	if len(row) != len(r.columns) {
		return fmt.Errorf("d1: row has %d values for %d columns", len(row), len(r.columns))
	}
	for i, v := range row {
		switch value := v.(type) {
		case nil, string:
			dest[i] = value
		case json.Number:
			if n, err := value.Int64(); err == nil {
				dest[i] = n
			} else {
				f, err := value.Float64()
				if err != nil {
					return err
				}
				dest[i] = f
			}
		case []any:
			blob := make([]byte, len(value))
			for j, item := range value {
				n, ok := item.(json.Number)
				if !ok {
					return fmt.Errorf("d1: invalid blob")
				}
				b, err := n.Int64()
				if err != nil || b < 0 || b > 255 {
					return fmt.Errorf("d1: invalid blob byte")
				}
				blob[j] = byte(b)
			}
			dest[i] = blob
		default:
			return fmt.Errorf("d1: unsupported result type %T", v)
		}
	}
	return nil
}
