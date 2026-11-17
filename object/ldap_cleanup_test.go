package object

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/xorm-io/xorm"
)

type ldapCleanupSearchFunc func(*goldap.SearchRequest) (*goldap.SearchResult, error)

func (f ldapCleanupSearchFunc) Search(request *goldap.SearchRequest) (*goldap.SearchResult, error) {
	return f(request)
}

func setupLdapCleanupTest(t *testing.T, softDelete bool) *Ldap {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	engine.SetMaxOpenConns(1)
	previousOrmer, previousEnforcer := ormer, userEnforcer
	t.Cleanup(func() {
		ormer, userEnforcer = previousOrmer, previousEnforcer
		engine.Close()
	})
	ormer = &Ormer{Engine: engine, driverName: "sqlite"}
	if err := engine.Sync2(new(User), new(Ldap), new(Organization), new(Token), new(Session), new(ThirdPartyLink), new(PasswordHistory), new(Syncer)); err != nil {
		t.Fatal(err)
	}
	rbacModel, err := model.NewModelFromString(`[request_definition]
r = sub, obj, act
[policy_definition]
p = sub, obj, act
[role_definition]
g = _, _
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`)
	if err != nil {
		t.Fatal(err)
	}
	enforcer, err := casbin.NewSyncedEnforcer(rbacModel)
	if err != nil {
		t.Fatal(err)
	}
	userEnforcer = NewUserGroupEnforcer(enforcer)
	ldap := &Ldap{Id: "directory", Owner: "org", BaseDn: "dc=example,dc=com"}
	if _, err := engine.Insert(ldap, &Organization{Owner: "admin", Name: "org", EnableSoftDeletion: softDelete}); err != nil {
		t.Fatal(err)
	}
	return ldap
}

func TestLdapCleanupDeletesOnlyMissingLinkedUsers(t *testing.T) {
	for _, softDelete := range []bool{false, true} {
		t.Run(fmt.Sprintf("softDeletion=%t", softDelete), func(t *testing.T) {
			ldap := setupLdapCleanupTest(t, softDelete)
			users := []*User{
				{Owner: "org", Name: "missing", Ldap: "gone"},
				{Owner: "org", Name: "present", Ldap: "uuid-present"},
				{Owner: "org", Name: "uid-fallback", Ldap: "uid-present"},
				{Owner: "org", Name: "cn-fallback", Ldap: "cn-present"},
				{Owner: "org", Name: "local"},
				{Owner: "other", Name: "other-org", Ldap: "gone"},
				{Owner: "org", Name: "already-deleted", Ldap: "old", IsDeleted: true},
				{Owner: "org", Name: "filtered", Ldap: "still-exists"},
			}
			if _, err := ormer.Engine.Insert(users); err != nil {
				t.Fatal(err)
			}
			if _, err := userEnforcer.AddGroupForUser("org/missing", "org/team"); err != nil {
				t.Fatal(err)
			}
			if _, err := ormer.Engine.Insert(&ThirdPartyLink{Owner: "org", UserName: "missing", ProviderName: "provider", ProviderId: "id"}, &PasswordHistory{Owner: "org", Name: "missing"}); err != nil {
				t.Fatal(err)
			}
			snapshot := []LdapUser{{Uuid: "uuid-present"}, {Uid: "uid-present"}, {Cn: "cn-present"}}
			search := ldapCleanupSearchFunc(func(request *goldap.SearchRequest) (*goldap.SearchResult, error) {
				if strings.Contains(request.Filter, "still-exists") {
					return &goldap.SearchResult{Entries: []*goldap.Entry{{DN: "uid=filtered," + ldap.BaseDn}}}, nil
				}
				if !strings.Contains(request.Filter, "(uid=gone)") {
					t.Fatalf("unexpected deletion candidate: %s", request.Filter)
				}
				return &goldap.SearchResult{}, nil
			})
			deleted, err := cleanupLdapUsers(ldap, snapshot, search)
			if err != nil || deleted != 1 {
				t.Fatalf("cleanup = %d, %v; want one deletion", deleted, err)
			}
			missing, err := getUser("org", "missing")
			if err != nil {
				t.Fatal(err)
			}
			if softDelete {
				if missing == nil || !missing.IsDeleted || missing.DeletedTime == "" {
					t.Fatal("missing user was not soft-deleted")
				}
			} else if missing != nil {
				t.Fatal("missing user was not removed")
			}
			for _, expected := range users[1:] {
				user, err := getUser(expected.Owner, expected.Name)
				if err != nil || user == nil || user.IsDeleted != expected.IsDeleted {
					t.Fatalf("unrelated user %s was changed: %v", expected.GetId(), err)
				}
			}
			links, err := GetThirdPartyLinksByUser("org", "missing")
			if err != nil || len(links) != 0 {
				t.Fatalf("third-party links were not cleaned: %v", err)
			}
			if _, exists, err := getPasswordHistory("org", "missing"); err != nil || exists {
				t.Fatalf("password history was not cleaned: %v", err)
			}
			if roles, err := userEnforcer.enforcer.GetRolesForUser("org/missing"); err != nil || len(roles) != 0 {
				t.Fatalf("group memberships were not cleaned: %v", err)
			}
			if deleted, err := cleanupLdapUsers(ldap, snapshot, search); err != nil || deleted != 0 {
				t.Fatalf("repeat cleanup = %d, %v; want no further deletions", deleted, err)
			}
			if softDelete {
				if user, err := getUser("org", "missing"); err != nil || user == nil || !user.IsDeleted {
					t.Fatal("repeat cleanup permanently deleted a soft-deleted user")
				}
			}
		})
	}
}

func TestLdapCleanupSkipsUnsafeSnapshots(t *testing.T) {
	for _, name := range []string{"empty", "missing-identifier", "multiple-servers", "changed-scope"} {
		t.Run(name, func(t *testing.T) {
			ldap := setupLdapCleanupTest(t, false)
			if _, err := ormer.Engine.Insert(&User{Owner: "org", Name: "missing", Ldap: "gone"}); err != nil {
				t.Fatal(err)
			}
			snapshot := []LdapUser{{Uuid: "someone-else"}}
			switch name {
			case "empty":
				snapshot = nil
			case "missing-identifier":
				snapshot = append(snapshot, LdapUser{})
			case "multiple-servers":
				if _, err := ormer.Engine.Insert(&Ldap{Id: "second", Owner: "org"}); err != nil {
					t.Fatal(err)
				}
			case "changed-scope":
				ldap.BaseDn = "dc=changed"
			}
			search := ldapCleanupSearchFunc(func(*goldap.SearchRequest) (*goldap.SearchResult, error) {
				t.Fatal("unsafe snapshot reached deletion verification")
				return nil, nil
			})
			deleted, err := cleanupLdapUsers(ldap, snapshot, search)
			if deleted != 0 || (name != "empty" && err == nil) {
				t.Fatalf("unsafe cleanup = %d, %v", deleted, err)
			}
			if user, err := getUser("org", "missing"); err != nil || user == nil || user.IsDeleted {
				t.Fatal("unsafe snapshot deleted a user")
			}
		})
	}
}

func TestLdapCleanupVerificationErrorsKeepUser(t *testing.T) {
	for _, name := range []string{"search-error", "referral", "nil-result"} {
		t.Run(name, func(t *testing.T) {
			ldap := setupLdapCleanupTest(t, false)
			if _, err := ormer.Engine.Insert(&User{Owner: "org", Name: "missing", Ldap: "gone"}); err != nil {
				t.Fatal(err)
			}
			search := ldapCleanupSearchFunc(func(*goldap.SearchRequest) (*goldap.SearchResult, error) {
				switch name {
				case "search-error":
					return &goldap.SearchResult{}, errors.New("connection lost")
				case "referral":
					return &goldap.SearchResult{Referrals: []string{"ldap://other-server"}}, nil
				default:
					return nil, nil
				}
			})
			if deleted, err := cleanupLdapUsers(ldap, []LdapUser{{Uuid: "present"}}, search); err == nil || deleted != 0 {
				t.Fatalf("failed verification = %d, %v", deleted, err)
			}
			if user, err := getUser("org", "missing"); err != nil || user == nil || user.IsDeleted {
				t.Fatal("failed LDAP verification deleted a user")
			}
		})
	}
}

func TestLdapCleanupEscapesIdentity(t *testing.T) {
	id := `alice*)(uid=*)`
	search := ldapCleanupSearchFunc(func(request *goldap.SearchRequest) (*goldap.SearchResult, error) {
		if request.BaseDN != "dc=example" || request.Scope != goldap.ScopeWholeSubtree || !strings.Contains(request.Filter, `(uid=alice\2a\29\28uid=\2a\29)`) {
			t.Fatalf("unsafe lookup request: %+v", request)
		}
		if _, err := goldap.CompileFilter(request.Filter); err != nil {
			t.Fatalf("invalid lookup filter: %v", err)
		}
		return &goldap.SearchResult{}, nil
	})
	if exists, err := ldapCleanupUserExists(search, "dc=example", id); err != nil || exists {
		t.Fatalf("existence check = %t, %v", exists, err)
	}
}
