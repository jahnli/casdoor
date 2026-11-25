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

package action

import (
	"encoding/json"
	"strings"
)

// Module is an installed action: its name in the organization, the WebAssembly module and the hosts it may call
type Module struct {
	Name   string
	Module []byte
	Hosts  []string
}

func IsNotImplemented(err error) bool {
	return err != nil && strings.Contains(err.Error(), "the action does not implement")
}

// ImportFile is a file the admin uploaded for an import, such as the realm export of another identity system
type ImportFile struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

// ImportListRequest asks an action of kind "import" for a page of one kind of objects of the other system
type ImportListRequest struct {
	// "certs", "groups", "applications", "users", "roles" or "permissions"
	Kind         string        `json:"kind"`
	Cursor       string        `json:"cursor"`
	Organization string        `json:"organization"`
	Files        []*ImportFile `json:"files"`
}

// ImportListResult holds the objects in Casdoor's own JSON, and the cursor of the next page, empty after the last one
type ImportListResult struct {
	Items    []json.RawMessage `json:"items"`
	Cursor   string            `json:"cursor"`
	Warnings []string          `json:"warnings"`
}

// ListImport runs "import.list" of an action of kind "import"
func ListImport(m *Module, config *Config, req *ImportListRequest) (*ImportListResult, error) {
	res := &ImportListResult{}
	err := Call(m.Name, m.Module, m.Hosts, "import.list", config, req, res)
	return res, err
}

// TestImport runs the optional "import.test" of an action of kind "import"
func TestImport(m *Module, config *Config, files []*ImportFile) error {
	return Call(m.Name, m.Module, m.Hosts, "import.test", config, map[string]interface{}{"files": files}, nil)
}

// VerifyImportedPassword runs "import.verifyPassword" of an action of kind "import" on the sign-in of a user whose
// password hash came from the other system in a format Casdoor doesn't know, with the configuration of the migration,
// for an action that checks the password with the other system
func VerifyImportedPassword(m *Module, config *Config, password string, hash string, salt string) (bool, error) {
	resp := struct {
		Valid bool `json:"valid"`
	}{}
	err := Call(m.Name, m.Module, m.Hosts, "import.verifyPassword", config, map[string]string{
		"password": password,
		"hash":     hash,
		"salt":     salt,
	}, &resp)
	return resp.Valid, err
}
