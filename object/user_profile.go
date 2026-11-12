package object

import (
	"maps"
	"strings"
)

// GetMaskedUserProfile restricts internal profile data in account and user query
// responses, independently of the organization's account-item configuration.
// Protocol token generation uses the original user, never this response copy.
func GetMaskedUserProfile(user *User, isGlobalAdmin bool) *User {
	if user == nil || isGlobalAdmin {
		return user
	}

	profile := *user
	profile.MaskedFields = nil
	for _, field := range []struct{ name, value string }{
		{"owner", user.Owner}, {"id", user.Id}, {"signupApplication", user.SignupApplication},
	} {
		if field.value != "" {
			profile.MaskedFields = append(profile.MaskedFields, field.name)
		}
	}
	if user.Type != "" {
		profile.MaskedFields = append(profile.MaskedFields, "type")
	}
	if len(user.Groups) > 0 {
		profile.MaskedFields = append(profile.MaskedFields, "groups")
	}
	for _, key := range []string{"user_id", "open_id", "department_value"} {
		if user.Properties[key] != "" {
			profile.MaskedFields = append(profile.MaskedFields, "properties."+key)
		}
	}
	profile.Type = ""
	profile.Owner = ""
	profile.Id = ""
	profile.SignupApplication = ""
	// A JWT or upstream OAuth token can contain the same restricted fields.
	profile.AccessToken = ""
	profile.OriginalToken = ""
	profile.OriginalRefreshToken = ""
	profile.Groups = nil
	profile.Properties = maps.Clone(user.Properties)
	delete(profile.Properties, "user_id")
	delete(profile.Properties, "open_id")
	delete(profile.Properties, "department_value")
	for key := range profile.Properties {
		if strings.HasPrefix(key, "oauth_") && (strings.HasSuffix(key, "_accessToken") || strings.HasSuffix(key, "_refreshToken")) {
			delete(profile.Properties, key)
		}
	}

	// Resolved role and permission sources can also expose group membership.
	profile.Roles = make([]*Role, len(user.Roles))
	for i, role := range user.Roles {
		if role != nil {
			masked := *role
			masked.Owner = ""
			if role.Owner == user.Owner {
				masked.Owner = CurrentOrganizationAlias
			}
			masked.Users = nil
			masked.Groups = nil
			masked.SourceGroups = nil
			masked.Roles = maskProfileIds(role.Roles, user.Owner)
			profile.Roles[i] = &masked
		}
	}
	profile.Permissions = make([]*Permission, len(user.Permissions))
	for i, permission := range user.Permissions {
		if permission != nil {
			masked := *permission
			masked.Owner = ""
			if permission.Owner == user.Owner {
				masked.Owner = CurrentOrganizationAlias
			}
			masked.Users = nil
			masked.Groups = nil
			masked.SourceGroups = nil
			masked.SourceRoles = maskProfileIds(permission.SourceRoles, user.Owner)
			masked.Roles = maskProfileIds(permission.Roles, user.Owner)
			masked.Model = maskProfileId(permission.Model, user.Owner)
			masked.Adapter = maskProfileId(permission.Adapter, user.Owner)
			profile.Permissions[i] = &masked
		}
	}

	return &profile
}

func maskProfileId(id, owner string) string {
	if strings.HasPrefix(id, owner+"/") {
		return CurrentOrganizationAlias + strings.TrimPrefix(id, owner)
	}
	return id
}

func maskProfileIds(ids []string, owner string) []string {
	if ids == nil {
		return nil
	}
	masked := make([]string, len(ids))
	for i, id := range ids {
		masked[i] = maskProfileId(id, owner)
	}
	return masked
}
