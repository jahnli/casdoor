package controllers

import (
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) maskUserProfile(user *object.User) *object.User {
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(c.GetSessionUsername())
	return maskUserProfileForOwner(user, c.IsGlobalAdmin(), owner)
}

func maskUserProfileForOwner(user *object.User, isGlobalAdmin bool, owner string) *object.User {
	profile := object.GetMaskedUserProfile(user, isGlobalAdmin)
	if user != nil && profile != user {
		if owner == user.Owner {
			// Requests resolve this fixed alias against the authenticated session.
			profile.Owner = object.CurrentOrganizationAlias
		}
	}
	return profile
}

func (c *ApiController) maskUserProfiles(users []*object.User) []*object.User {
	isGlobalAdmin := c.IsGlobalAdmin()
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(c.GetSessionUsername())
	profiles := make([]*object.User, len(users))
	for i, user := range users {
		profiles[i] = maskUserProfileForOwner(user, isGlobalAdmin, owner)
	}
	return profiles
}

func (c *ApiController) maskLoginApplicationOrganization(application *object.Application) *object.Application {
	if application == nil || c.IsGlobalAdmin() {
		return application
	}
	if c.Ctx.Input.GetData("profileApplicationName") == application.Name {
		return object.GetMaskedProfileApplication(application, false)
	}
	actor := c.getCurrentUser()
	if actor == nil || actor.Owner != application.Organization {
		return application
	}
	masked := *application
	masked.Organization = object.CurrentOrganizationAlias
	masked.OrganizationObj = object.GetMaskedProfileOrganization(application.OrganizationObj, false)
	return &masked
}
