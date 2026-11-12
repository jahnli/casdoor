package object

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestGetMaskedUserProfile(t *testing.T) {
	user := &User{
		Owner: "semi", Name: "alice", Id: "sso-subject", Type: "internal-user-type",
		SignupApplication: "private-signup-app", AccessToken: "private-jwt", OriginalToken: "private-upstream-token", OriginalRefreshToken: "private-refresh-token",
		Groups: []string{"semi/private-group"},
		Properties: map[string]string{
			"open_id": "private-open-id", "department_value": "private-department-json",
			"department": "研发部", "job_number": "0410",
		},
		Roles:       []*Role{{Name: "role", Groups: []string{"semi/private-group"}, SourceGroups: []string{"semi/private-group"}}, nil},
		Permissions: []*Permission{{Name: "permission", Groups: []string{"semi/private-group"}, SourceGroups: []string{"semi/private-group"}, SourceRoles: []string{"semi/role"}}, nil},
	}
	before, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	profile := GetMaskedUserProfile(user, false)
	if !reflect.DeepEqual(profile.MaskedFields, []string{"owner", "id", "signupApplication", "type", "groups", "properties.open_id", "properties.department_value"}) {
		t.Errorf("masked fields = %v, want presence of the seven restricted fields", profile.MaskedFields)
	}
	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"private-open-id", "private-department-json", "semi/private-group", "internal-user-type", "sso-subject", "private-signup-app", "private-jwt", "private-upstream-token", "private-refresh-token", "semi/"} {
		if strings.Contains(string(data), value) {
			t.Errorf("restricted value %q leaked in profile", value)
		}
	}
	if profile.Id != "" || profile.Owner != "" || profile.SignupApplication != "" || profile.Name != user.Name {
		t.Error("restricted identifiers must be cleared while preserving the account name")
	}
	if profile.Properties["department"] != "研发部" || profile.Properties["job_number"] != "0410" {
		t.Error("readable employment fields must be preserved")
	}
	if profile.Roles[0].Name != "role" || !reflect.DeepEqual(profile.Permissions[0].SourceRoles, []string{CurrentOrganizationAlias + "/role"}) {
		t.Error("role and permission information must be preserved")
	}
	profile.Properties["department"] = "changed"

	after, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("masking must not mutate the original user, properties, roles or permissions")
	}
	if profile := GetMaskedUserProfile(user, true); profile != user {
		t.Error("global administrators must receive the original profile")
	}
}

func TestGetMaskedUserProfileEmpty(t *testing.T) {
	if GetMaskedUserProfile(nil, false) != nil || GetMaskedUserProfile(nil, true) != nil {
		t.Error("a missing user must stay nil")
	}
	profile := GetMaskedUserProfile(&User{}, false)
	if profile.Properties != nil || len(profile.Groups) != 0 || profile.Type != "" || len(profile.MaskedFields) != 0 {
		t.Error("an empty profile must stay empty")
	}
}

func TestGetMaskedUserProfileFieldPresence(t *testing.T) {
	user := &User{
		Type:       "normal-user",
		Groups:     []string{},
		Properties: map[string]string{"open_id": "", "department_value": "internal-department-json"},
	}
	profile := GetMaskedUserProfile(user, false)
	if !reflect.DeepEqual(profile.MaskedFields, []string{"type", "properties.department_value"}) {
		t.Errorf("masked fields = %v, empty fields must not be marked", profile.MaskedFields)
	}
	adminProfile := GetMaskedUserProfile(user, true)
	if len(adminProfile.MaskedFields) != 0 || adminProfile.Type != user.Type || adminProfile.Properties["department_value"] != user.Properties["department_value"] {
		t.Error("global administrators must receive the original data without masking metadata")
	}
}
