// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// remove-user-metrics previews or applies the cleanup of retired user metrics.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/xorm-io/xorm"
	"github.com/xorm-io/xorm/schemas"
)

type change struct {
	label string
	sql   string
	args  []interface{}
}

func removed(name string) bool {
	switch strings.ToLower(name) {
	case "score", "karma", "ranking":
		return true
	}
	return false
}

// Preserve arbitrary properties of retained items, including future fields.
func cleanList(raw, key string) (string, bool, error) {
	if strings.TrimSpace(raw) == "" {
		return raw, false, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return "", false, err
	}
	kept := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		var name string
		if key == "" {
			if err := json.Unmarshal(item, &name); err != nil {
				return "", false, err
			}
		} else {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(item, &fields); err != nil {
				return "", false, err
			}
			if value, ok := fields[key]; ok {
				if err := json.Unmarshal(value, &name); err != nil {
					return "", false, err
				}
			}
		}
		if !removed(name) {
			kept = append(kept, item)
		}
	}
	if len(items) == len(kept) {
		return raw, false, nil
	}
	encoded, err := json.Marshal(kept)
	return string(encoded), true, err
}

func qualifiedTable(engine *xorm.Engine, table string) string {
	schema := engine.Dialect().URI().Schema
	if schema != "" && engine.Dialect().URI().DBType == schemas.POSTGRES {
		return engine.Quote(schema) + "." + engine.Quote(table)
	}
	return engine.Quote(table)
}

func planCleanup(engine *xorm.Engine, prefix string) ([]change, error) {
	metas, err := engine.DBMetas()
	if err != nil {
		return nil, err
	}
	tables := map[string]*schemas.Table{}
	for _, table := range metas {
		tables[table.Name] = table
	}
	changes := []change{}
	for _, spec := range []struct{ table, column, key string }{
		{"organization", "account_items", "name"},
		{"organization", "default_token_fields", ""},
		{"application", "token_fields", ""},
		{"syncer", "table_columns", "casdoorName"},
	} {
		name := prefix + spec.table
		table := tables[name]
		if table == nil || table.GetColumn(spec.column) == nil {
			continue
		}
		qt, qc := qualifiedTable(engine, name), engine.Quote(spec.column)
		rows, err := engine.QueryString("SELECT " + engine.Quote("owner") + ", " + engine.Quote("name") + ", " + qc + " FROM " + qt)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			cleaned, changed, err := cleanList(row[spec.column], spec.key)
			if err != nil {
				return nil, fmt.Errorf("invalid JSON in %s.%s (owner=%q, name=%q): %w", name, spec.column, row["owner"], row["name"], err)
			}
			if !changed {
				continue
			}
			set := qc + " = ?"
			// Old mappings may have used a metric as the identity key or affiliation ID.
			// Disable those syncers until an administrator reviews their new mappings.
			if spec.table == "syncer" {
				if table.GetColumn("is_enabled") == nil {
					return nil, fmt.Errorf("%s has no is_enabled column; cannot safely retire its mappings", name)
				}
				set += ", " + engine.Quote("is_enabled") + " = ?"
			}
			args := []interface{}{cleaned}
			if spec.table == "syncer" {
				args = append(args, false)
			}
			args = append(args, row["owner"], row["name"])
			label := fmt.Sprintf("clean %s.%s (owner=%q, name=%q)", name, spec.column, row["owner"], row["name"])
			if spec.table == "syncer" {
				label += " and disable syncer for mapping review"
			}
			changes = append(changes, change{label, "UPDATE " + qt + " SET " + set + " WHERE " + engine.Quote("owner") + " = ? AND " + engine.Quote("name") + " = ?", args})
		}
	}
	for _, spec := range []struct{ table, column string }{
		{"user", "score"}, {"user", "karma"}, {"user", "ranking"}, {"organization", "init_score"},
	} {
		name := prefix + spec.table
		if table := tables[name]; table != nil && table.GetColumn(spec.column) != nil {
			query := "ALTER TABLE " + qualifiedTable(engine, name) + " DROP COLUMN " + engine.Quote(spec.column)
			changes = append(changes, change{"drop " + name + "." + spec.column, query, nil})
		}
	}
	return changes, nil
}

func applyCleanup(engine *xorm.Engine, changes []change) error {
	session := engine.NewSession()
	defer session.Close()
	// MySQL DDL implicitly commits. Other supported drivers can roll back the full plan.
	transactional := engine.Dialect().URI().DBType != schemas.MYSQL
	if transactional {
		if err := session.Begin(); err != nil {
			return err
		}
		defer session.Rollback()
	}
	for _, change := range changes {
		args := append([]interface{}{change.sql}, change.args...)
		if _, err := session.Exec(args...); err != nil {
			return fmt.Errorf("%s: %w", change.label, err)
		}
	}
	if transactional {
		return session.Commit()
	}
	return nil
}

func run() error {
	config := flag.String("config", "conf/app.conf", "Casdoor configuration file")
	apply := flag.Bool("apply", false, "apply the displayed cleanup (default: read-only preview)")
	flag.Parse()
	if err := web.LoadAppConfig("ini", *config); err != nil {
		return fmt.Errorf("cannot load configuration: %w", err)
	}
	driver := conf.GetConfigString("driverName")
	switch driver {
	case "mysql", "postgres", "sqlite", "sqlite3":
	default:
		return fmt.Errorf("unsupported cleanup driver %q; supported: mysql, postgres, sqlite/sqlite3", driver)
	}
	adapter, err := object.NewAdapter(driver, conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
	if err != nil {
		return fmt.Errorf("cannot open database; verify the selected configuration and connection settings")
	}
	defer adapter.Engine.Close()
	changes, err := planCleanup(adapter.Engine, conf.GetConfigString("tableNamePrefix"))
	if err != nil {
		return err
	}
	for _, change := range changes {
		fmt.Println(change.label)
	}
	fmt.Printf("Planned changes: %d\n", len(changes))
	if !*apply {
		fmt.Println("Preview only. Stop Casdoor and back up the database before running with -apply.")
		return nil
	}
	if err := applyCleanup(adapter.Engine, changes); err != nil {
		return err
	}
	fmt.Println("Cleanup complete.")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
