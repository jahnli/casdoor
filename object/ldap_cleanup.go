package object

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	goldap "github.com/go-ldap/ldap/v3"
)

type ldapCleanupSearcher interface {
	Search(*goldap.SearchRequest) (*goldap.SearchResult, error)
}

// CleanupLdapUsers fetches its own complete snapshot: the manual sync request
// contains only selected users and must never be used to infer deletions.
func CleanupLdapUsers(owner string, ldapId string) (int, error) {
	ldap, err := GetLdapOfOwner(owner, ldapId)
	if err != nil {
		return 0, err
	}
	if ldap == nil {
		return 0, fmt.Errorf("LDAP %s does not exist in organization %s", ldapId, owner)
	}
	conn, err := ldap.GetLdapConn()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	users, err := conn.GetLdapUsers(ldap)
	if err != nil {
		return 0, err
	}
	return cleanupLdapUsers(ldap, users, conn.Conn)
}

// cleanupLdapUsers is called only after a successful full LDAP search.
// Existing import/update operations and scheduling remain unchanged.
func cleanupLdapUsers(ldap *Ldap, users []LdapUser, searcher ldapCleanupSearcher) (int, error) {
	if len(users) == 0 {
		logs.Warning("LDAP cleanup skipped for %s: the complete user list is empty", ldap.Id)
		return 0, nil
	}

	// The user record has no LDAP server ID. Deletion is safe only under the
	// single-LDAP-per-organization contract, including on future configuration changes.
	ldaps, err := GetLdaps(ldap.Owner)
	if err != nil {
		return 0, err
	}
	if len(ldaps) != 1 || ldaps[0].Id != ldap.Id {
		return 0, fmt.Errorf("LDAP cleanup requires exactly one LDAP configuration in organization %s", ldap.Owner)
	}
	if ldaps[0].BaseDn != ldap.BaseDn || ldaps[0].Filter != ldap.Filter || ldaps[0].Host != ldap.Host || ldaps[0].Port != ldap.Port {
		return 0, fmt.Errorf("LDAP configuration changed during sync; cleanup skipped for %s", ldap.Id)
	}

	present := make(map[string]bool, len(users))
	for _, user := range users {
		id := user.GetLdapUuid()
		if strings.TrimSpace(id) == "" {
			return 0, fmt.Errorf("LDAP cleanup skipped for %s: a directory entry has no user identifier", ldap.Id)
		}
		present[id] = true
	}

	var candidates []*User
	err = ormer.Engine.Where("owner = ? AND ldap <> '' AND (is_deleted = ? OR is_deleted IS NULL)", ldap.Owner, false).Find(&candidates)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, user := range candidates {
		if present[user.Ldap] {
			continue
		}
		exists, err := ldapCleanupUserExists(searcher, ldap.BaseDn, user.Ldap)
		if err != nil {
			return deleted, fmt.Errorf("LDAP cleanup could not verify user %s: %w", user.GetId(), err)
		}
		if exists {
			continue
		}

		// Re-read before deletion so an already soft-deleted or unlinked account
		// is not deleted again on the next cycle (DeleteUser would hard-delete it).
		current, err := getUser(user.Owner, user.Name)
		if err != nil {
			return deleted, err
		}
		if current == nil || current.IsDeleted || current.Ldap != user.Ldap {
			continue
		}
		affected, err := DeleteUser(current)
		if err != nil {
			return deleted, fmt.Errorf("LDAP cleanup failed to delete user %s: %w", user.GetId(), err)
		}
		if affected {
			deleted++
			logs.Info("LDAP cleanup for %s deleted user %s", ldap.Id, user.GetId())
		}
	}
	return deleted, nil
}

func ldapCleanupUserExists(searcher ldapCleanupSearcher, baseDn string, id string) (bool, error) {
	// GetLdapUuid falls back from the directory UUID to uid (sAMAccountName on
	// AD), then cn. Check all forms, without the configured listing filter, so
	// a narrower filter alone does not delete an account still in the directory.
	value := goldap.EscapeFilter(id)
	filter := fmt.Sprintf("(|(entryUUID=%s)(objectGUID=%s)(uid=%s)(sAMAccountName=%s)(cn=%s))", value, value, value, value, value)
	request := goldap.NewSearchRequest(baseDn, goldap.ScopeWholeSubtree, goldap.NeverDerefAliases,
		1, 0, false, filter, []string{"1.1"}, nil)
	result, err := searcher.Search(request)
	if err != nil {
		return false, err
	}
	if result == nil || len(result.Referrals) != 0 {
		return false, fmt.Errorf("LDAP existence check did not return a complete result")
	}
	return len(result.Entries) != 0, nil
}
