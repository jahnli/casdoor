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

package object

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xorm-io/xorm"
)

func TestRetiredUserMetricsSchemaAndPayload(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	engine.SetMaxOpenConns(1)
	if err := engine.Sync2(new(User), new(Organization)); err != nil {
		t.Fatal(err)
	}
	metas, err := engine.DBMetas()
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range metas {
		for _, column := range table.Columns() {
			if isRemovedUserMetric(column.Name) || column.Name == "init_score" {
				t.Fatalf("new schema contains %s.%s", table.Name, column.Name)
			}
		}
	}
	user := &User{}
	if err := json.Unmarshal([]byte(`{"name":"alice","score":100,"karma":5,"ranking":1}`), user); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []interface{}{user, getUserWithoutThirdIdp(user), &Organization{}} {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]interface{}
		if err := json.Unmarshal(encoded, &fields); err != nil {
			t.Fatal(err)
		}
		for name := range fields {
			if isRemovedUserMetric(name) || name == "initScore" {
				t.Fatalf("retired field returned by API/JWT: %s", name)
			}
		}
	}
}

func TestRetiredUserMetricsConfiguration(t *testing.T) {
	organization := &Organization{
		AccountItems:       []*AccountItem{{Name: "Score"}, {Name: "Karma"}, {Name: "Ranking"}, {Name: "Affiliation", Visible: true}},
		DefaultTokenFields: []string{"Score", "Karma", "Ranking", "Name", "Properties.score"},
	}
	organization.removeUserMetricItems()
	if len(organization.AccountItems) != 1 || organization.AccountItems[0].Name != "Affiliation" || !organization.AccountItems[0].Visible {
		t.Fatalf("unexpected account items: %v", organization.AccountItems)
	}
	if !reflect.DeepEqual(organization.DefaultTokenFields, []string{"Name", "Properties.score"}) {
		t.Fatalf("unexpected token fields: %v", organization.DefaultTokenFields)
	}
	syncer := &Syncer{TableColumns: []*TableColumn{{Name: "points", CasdoorName: "Score"}, {Name: "reputation", CasdoorName: "Karma"}, {Name: "position", CasdoorName: "Ranking"}, {Name: "username", CasdoorName: "Name"}}}
	if fields := syncer.getCasdoorColumns(); !reflect.DeepEqual(fields, []string{"name"}) {
		t.Fatalf("retired fields included in SQL columns: %v", fields)
	}
	if values := syncer.getMapFromOriginalUser(&User{Name: "alice"}); !reflect.DeepEqual(values, map[string]string{"username": "alice"}) {
		t.Fatalf("retired fields would overwrite remote data: %v", values)
	}
	if err := RunSyncer(syncer); err == nil {
		t.Fatal("obsolete syncer was allowed to connect")
	}
	for _, field := range []string{"score", "karma", "ranking"} {
		if _, err := SetUserField(&User{}, field, "100"); err == nil {
			t.Fatalf("obsolete field %s was allowed in a direct update", field)
		}
		if _, err := UpdateUser("org/alice", &User{}, []string{field}, true); err == nil {
			t.Fatalf("obsolete field %s was allowed in an admin update", field)
		}
	}
}

func TestAffiliationSyncWithoutUserMetrics(t *testing.T) {
	syncer := &Syncer{Organization: "org", TableColumns: []*TableColumn{{Name: "department", CasdoorName: "Affiliation", IsHashed: true}}}
	affiliations := map[int]string{42: "Engineering"}
	user := syncer.createUserFromOriginalUser(&OriginalUser{Name: "alice", Affiliation: "42"}, affiliations)
	if user.Affiliation != "Engineering" {
		t.Fatalf("affiliation = %q", user.Affiliation)
	}
	original, err := syncer.createOriginalUserFromUser(user, affiliations)
	if err != nil || original.Affiliation != "42" || user.Affiliation != "Engineering" {
		t.Fatalf("reverse affiliation mapping = %v, %v", original, err)
	}
	if syncer.calculateHash(user) != syncer.calculateHash(syncer.createUserFromOriginalUser(original, affiliations)) {
		t.Fatal("affiliation round-trip changed sync hash")
	}
	if _, err := syncer.createOriginalUserFromUser(&User{Affiliation: "Missing"}, affiliations); err == nil {
		t.Fatal("unknown affiliation would be written to an ID column")
	}
	if _, err := syncer.createOriginalUserFromUser(user, map[int]string{1: "Engineering", 2: "Engineering"}); err == nil {
		t.Fatal("ambiguous affiliation would be written to an ID column")
	}
}
