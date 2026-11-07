package object

import "github.com/xorm-io/builder"

// GetAuditUsers selects only profile fields, without expanding roles, permissions or MFA.
func GetAuditUsers(names map[string][]string) ([]*User, error) {
	users := []*User{}
	if len(names) == 0 {
		return users, nil
	}
	conditions := []builder.Cond{}
	for owner, usernames := range names {
		conditions = append(conditions, builder.Eq{"owner": owner}.And(builder.In("name", usernames)))
	}
	err := ormer.Engine.Cols("owner", "name", "display_name", "avatar", "permanent_avatar", "properties", "gender", "is_admin", "bio", "homepage", "affiliation", "email", "phone", "birthday", "location").
		Where(builder.Or(conditions...)).Find(&users)
	return users, err
}
