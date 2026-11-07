package controllers

import "github.com/casdoor/casdoor/object"

func (c *ApiController) responseAuditList(rows interface{}, count int64) {
	if c.Ctx.Input.Query("includeUsers") != "true" {
		c.ResponseOk(rows, count)
		return
	}
	globalAdmin, actor := c.isGlobalAdmin()
	names := map[string][]string{}
	seen := map[string]bool{}
	add := func(owner, name string) {
		if owner == "" || name == "" || (!globalAdmin && (actor == nil || !actor.IsAdmin || actor.Owner != owner)) {
			return
		}
		key := owner + "/" + name
		if !seen[key] {
			seen[key] = true
			names[owner] = append(names[owner], name)
		}
	}
	switch records := rows.(type) {
	case []*object.Record:
		for _, row := range records {
			add(row.Organization, row.User)
		}
	case []*object.Token:
		for _, row := range records {
			add(row.Organization, row.User)
		}
	case []*object.Session:
		for _, row := range records {
			add(row.Owner, row.Name)
		}
	case []*object.VerificationRecord:
		for _, row := range records {
			add(row.Owner, row.User)
		}
	}
	users, err := object.GetAuditUsers(names)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	profiles := map[string]interface{}{}
	for key := range seen {
		profiles[key] = nil
	}
	organizations := map[string]*object.Organization{}
	for _, user := range users {
		organization, found := organizations[user.Owner]
		if !found {
			organization, err = object.GetOrganizationByUser(user)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
			organizations[user.Owner] = organization
		}
		if organization == nil {
			continue
		}
		id := user.GetId()
		user, err = object.GetFilteredUser(user, true, true, organization.AccountItems)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		properties := map[string]string{}
		for _, key := range []string{"department", "job_level", "job_date", "job_number", "job_type", "job_description", "ethnicity", "coverUrl"} {
			if value, ok := user.Properties[key]; ok {
				properties[key] = value
			}
		}
		profiles[id] = map[string]interface{}{
			"owner": user.Owner, "name": user.Name, "displayName": user.DisplayName,
			"avatar": user.Avatar, "permanentAvatar": user.PermanentAvatar, "properties": properties,
			"gender": user.Gender, "isAdmin": user.IsAdmin, "bio": user.Bio,
			"homepage": user.Homepage, "affiliation": user.Affiliation, "email": user.Email,
			"phone": user.Phone, "birthday": user.Birthday, "location": user.Location,
		}
	}
	c.ResponseJsonData(&Response{Status: "ok", Data3: profiles}, rows, count)
}
