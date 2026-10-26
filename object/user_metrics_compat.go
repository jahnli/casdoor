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

import "strings"

// Keep configurations saved by older versions from reintroducing removed fields.
func isRemovedUserMetric(name string) bool {
	switch strings.ToLower(name) {
	case "score", "karma", "ranking":
		return true
	}
	return false
}

func (organization *Organization) removeUserMetricItems() {
	items := make([]*AccountItem, 0, len(organization.AccountItems))
	for _, item := range organization.AccountItems {
		if item != nil && !isRemovedUserMetric(item.Name) {
			items = append(items, item)
		}
	}
	organization.AccountItems = items
	fields := make([]string, 0, len(organization.DefaultTokenFields))
	for _, field := range organization.DefaultTokenFields {
		if !isRemovedUserMetric(field) {
			fields = append(fields, field)
		}
	}
	organization.DefaultTokenFields = fields
}
