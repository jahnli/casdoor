package object

import "testing"

func TestMaskedProfileConfiguration(t *testing.T) {
	organization := &Organization{Name: "private-tenant", DisplayName: "Private tenant", DefaultApplication: "private-app", CountryCodes: []string{"CN"}, MfaRememberInHours: 24}
	application := &Application{Name: "private-app", Organization: "private-tenant", OrganizationObj: organization}
	profile := GetMaskedProfileApplication(application, false)
	if profile.Name != CurrentApplicationAlias || profile.Organization != CurrentOrganizationAlias || profile.OrganizationObj.Name != CurrentOrganizationAlias || profile.OrganizationObj.DisplayName != "" || profile.OrganizationObj.DefaultApplication != "" {
		t.Error("account application configuration must not expose the internal tenant or application")
	}
	if profile.OrganizationObj.MfaRememberInHours != 24 || len(profile.OrganizationObj.CountryCodes) != 1 {
		t.Error("MFA policy must be preserved")
	}
	if application.Name != "private-app" || organization.Name != "private-tenant" || organization.DefaultApplication != "private-app" {
		t.Error("response masking must not mutate the source configuration")
	}
	if GetMaskedProfileApplication(application, true) != application || GetMaskedProfileOrganization(organization, true) != organization {
		t.Error("global administrators must retain the original configuration")
	}
	if GetMaskedProfileApplication(nil, false) != nil || GetMaskedProfileOrganization(nil, false) != nil {
		t.Error("nil configurations must stay nil")
	}
}
