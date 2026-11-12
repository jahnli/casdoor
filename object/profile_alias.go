package object

// These fixed aliases are resolved from the authenticated session before API
// authorization. They contain no tenant name or application identifier.
const (
	CurrentOrganizationAlias = "@current"
	CurrentApplicationAlias  = "@current"
)

func GetMaskedProfileOrganization(organization *Organization, isGlobalAdmin bool) *Organization {
	if organization == nil || isGlobalAdmin {
		return organization
	}
	profile := *organization
	profile.Name = CurrentOrganizationAlias
	profile.DisplayName = ""
	profile.DefaultApplication = ""
	return &profile
}

func GetMaskedProfileApplication(application *Application, isGlobalAdmin bool) *Application {
	if application == nil || isGlobalAdmin {
		return application
	}
	profile := *application
	profile.Name = CurrentApplicationAlias
	profile.DisplayName = ""
	profile.Organization = CurrentOrganizationAlias
	profile.OrganizationObj = GetMaskedProfileOrganization(application.OrganizationObj, false)
	return &profile
}
