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

package main

import (
	"strings"
	"testing"

	"github.com/xorm-io/xorm"
)

func testDatabase(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { engine.Close() })
	for _, query := range []string{
		`CREATE TABLE t_user (owner TEXT, name TEXT, affiliation TEXT, score INTEGER, karma INTEGER, ranking INTEGER)`,
		`INSERT INTO t_user VALUES ('org', 'alice', 'Engineering', 42, 3, 1)`,
		`CREATE TABLE t_organization (owner TEXT, name TEXT, init_score INTEGER, account_items TEXT, default_token_fields TEXT)`,
		`INSERT INTO t_organization VALUES ('admin', 'org', 100, '[{"name":"Score"},{"name":"Karma"},{"name":"Ranking"},{"name":"Affiliation","visible":true,"custom":42}]', '["Score","Karma","Ranking","Name"]')`,
		`CREATE TABLE t_application (owner TEXT, name TEXT, token_fields TEXT)`,
		`INSERT INTO t_application VALUES ('admin', 'app', '["score","karma","ranking","Name","Properties.score"]')`,
		`CREATE TABLE t_syncer (owner TEXT, name TEXT, table_columns TEXT, is_enabled INTEGER)`,
		`INSERT INTO t_syncer VALUES ('admin', 'sync', '[{"casdoorName":"Score","name":"remote_score","isKey":true},{"casdoorName":"Name","name":"username"}]', 1)`,
		`INSERT INTO t_syncer VALUES ('admin', 'unchanged', '[{"casdoorName":"Name","name":"username"}]', 1)`,
		`CREATE TABLE user (score INTEGER)`,
		`INSERT INTO user VALUES (77)`,
	} {
		if _, err := engine.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	return engine
}

func TestCleanupPreviewApplyAndRepeat(t *testing.T) {
	engine := testDatabase(t)
	changes, err := planCleanup(engine, "t_")
	if err != nil || len(changes) != 8 {
		t.Fatalf("plan = %d changes, %v", len(changes), err)
	}
	rows, err := engine.QueryString(`SELECT score, affiliation FROM t_user`)
	if err != nil || rows[0]["score"] != "42" {
		t.Fatalf("preview modified user: %v, %v", rows, err)
	}
	if err := applyCleanup(engine, changes); err != nil {
		t.Fatal(err)
	}
	rows, err = engine.QueryString(`SELECT * FROM t_user`)
	if err != nil || rows[0]["name"] != "alice" || rows[0]["affiliation"] != "Engineering" {
		t.Fatalf("unrelated user data changed: %v, %v", rows, err)
	}
	for _, field := range []string{"score", "karma", "ranking"} {
		if _, ok := rows[0][field]; ok {
			t.Fatalf("retired column remains: %s", field)
		}
	}
	rows, _ = engine.QueryString(`SELECT * FROM t_organization`)
	if rows[0]["account_items"] != `[{"name":"Affiliation","visible":true,"custom":42}]` || rows[0]["default_token_fields"] != `["Name"]` {
		t.Fatalf("organization cleanup lost retained configuration: %v", rows)
	}
	rows, _ = engine.QueryString(`SELECT token_fields FROM t_application`)
	if rows[0]["token_fields"] != `["Name","Properties.score"]` {
		t.Fatalf("custom property removed: %v", rows)
	}
	rows, _ = engine.QueryString(`SELECT * FROM t_syncer ORDER BY name`)
	if rows[0]["is_enabled"] != "0" || rows[0]["table_columns"] != `[{"casdoorName":"Name","name":"username"}]` || rows[1]["is_enabled"] != "1" {
		t.Fatalf("incorrect syncer cleanup: %v", rows)
	}
	rows, _ = engine.QueryString(`SELECT score FROM user`)
	if rows[0]["score"] != "77" {
		t.Fatal("unprefixed table modified")
	}
	changes, err = planCleanup(engine, "t_")
	if err != nil || len(changes) != 0 {
		t.Fatalf("repeat plan = %v, %v", changes, err)
	}
}

func TestCleanupRejectsInvalidJSONBeforeWriting(t *testing.T) {
	engine := testDatabase(t)
	if _, err := engine.Exec(`UPDATE t_application SET token_fields = 'invalid'`); err != nil {
		t.Fatal(err)
	}
	changes, err := planCleanup(engine, "t_")
	if err == nil || changes != nil {
		t.Fatalf("invalid JSON accepted: %v, %v", changes, err)
	}
	rows, _ := engine.QueryString(`SELECT account_items FROM t_organization`)
	if !strings.Contains(rows[0]["account_items"], "Score") {
		t.Fatal("planning modified configuration")
	}
}

func TestCleanupRollsBackWhenColumnCannotBeDropped(t *testing.T) {
	engine := testDatabase(t)
	if _, err := engine.Exec(`CREATE INDEX keep_score ON t_user(score)`); err != nil {
		t.Fatal(err)
	}
	changes, err := planCleanup(engine, "t_")
	if err != nil {
		t.Fatal(err)
	}
	if err := applyCleanup(engine, changes); err == nil {
		t.Fatal("expected dependent index to prevent column deletion")
	}
	rows, _ := engine.QueryString(`SELECT account_items FROM t_organization`)
	if !strings.Contains(rows[0]["account_items"], "Score") {
		t.Fatal("configuration update was not rolled back")
	}
	rows, _ = engine.QueryString(`SELECT is_enabled FROM t_syncer WHERE name = 'sync'`)
	if rows[0]["is_enabled"] != "1" {
		t.Fatal("syncer update was not rolled back")
	}
}
