package controllers

import (
	"testing"

	"github.com/casdoor/casdoor/object"
)

func TestProfileRoutingNeverReturnsInternalIdentifiers(t *testing.T) {
	user := &object.User{Owner: "private-tenant", Name: "alice", Id: "private-id", SignupApplication: "private-app"}
	for _, tt := range []struct {
		name, owner, wantOwner string
		global                 bool
	}{
		{"self", "private-tenant", object.CurrentOrganizationAlias, false},
		{"organization administrator", "private-tenant", object.CurrentOrganizationAlias, false},
		{"foreign public profile", "another-tenant", "", false},
		{"anonymous", "", "", false},
		{"global administrator", "built-in", "private-tenant", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			profile := maskUserProfileForOwner(user, tt.global, tt.owner)
			if profile.Owner != tt.wantOwner {
				t.Fatalf("owner = %q, want %q", profile.Owner, tt.wantOwner)
			}
			if !tt.global && (profile.Id != "" || profile.SignupApplication != "") {
				t.Error("restricted identifiers must be removed")
			}
			if tt.global && profile != user {
				t.Error("global administrators must retain the original profile")
			}
		})
	}
	if user.Owner != "private-tenant" || user.Id != "private-id" || user.SignupApplication != "private-app" {
		t.Error("response masking must not modify the database object")
	}
}
