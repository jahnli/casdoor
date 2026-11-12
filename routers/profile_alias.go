package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// Resolve aliases before both authorization and controller execution, so they
// cannot authorize one tenant while executing an operation against another.
func resolveProfileAliases(ctx *context.Context, username string) error {
	return resolveProfileAliasesWithApplication(ctx, username, func(username string) (*object.Application, error) {
		user, err := object.GetUser(username)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("auth:Unauthorized operation")
		}
		return object.GetApplicationByUser(user)
	})
}

func resolveProfileAliasesWithApplication(ctx *context.Context, username string, getApplication func(string) (*object.Application, error)) error {
	if username == "" || object.IsAppUser(username) {
		return nil
	}
	if !strings.Contains(ctx.Request.URL.RawQuery, object.CurrentOrganizationAlias) &&
		!strings.Contains(ctx.Request.URL.RawQuery, url.QueryEscape(object.CurrentOrganizationAlias)) &&
		!strings.Contains(ctx.Request.Form.Encode(), url.QueryEscape(object.CurrentOrganizationAlias)) &&
		!bytes.Contains(ctx.Input.RequestBody, []byte(url.QueryEscape(object.CurrentOrganizationAlias))) &&
		!bytes.Contains(ctx.Input.RequestBody, []byte(object.CurrentOrganizationAlias)) {
		return nil
	}
	owner, _, err := util.GetOwnerAndNameFromIdWithError(username)
	if err != nil {
		return err
	}
	applicationName := ""
	resolve := func(key, value string) (string, error) {
		if !strings.Contains(value, object.CurrentOrganizationAlias) {
			return value, nil
		}
		applicationValue, applicationReference := profileApplicationReference(key, value, ctx.Request.URL.Path)
		if applicationReference {
			if applicationName == "" {
				app, err := getApplication(username)
				if err != nil {
					return "", err
				}
				if app == nil {
					return "", fmt.Errorf("application not found")
				}
				applicationName = app.Name
			}
			ctx.Input.SetData("profileApplicationName", applicationName)
			resolvedName := applicationName
			if applicationValue == object.CurrentApplicationAlias+"-org-"+object.CurrentOrganizationAlias {
				resolvedName += "-org-" + owner
			}
			if strings.HasPrefix(value, "admin/") {
				return "admin/" + resolvedName, nil
			}
			return resolvedName, nil
		}
		return resolveOrganizationAlias(key, value, owner, strings.Contains(ctx.Request.URL.Path, "organization")), nil
	}
	resolveValues := func(values url.Values) error {
		for key, items := range values {
			for i, value := range items {
				resolved, err := resolve(key, value)
				if err != nil {
					return err
				}
				items[i] = resolved
			}
		}
		return nil
	}
	query := ctx.Request.URL.Query()
	if err := resolveValues(query); err != nil {
		return err
	}
	ctx.Request.URL.RawQuery = query.Encode()

	contentType := ctx.Request.Header.Get("Content-Type")
	if json.Valid(ctx.Input.RequestBody) {
		if ctx.Request.Form == nil {
			ctx.Request.Form = query
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := ctx.Request.ParseMultipartForm(web.BConfig.MaxMemory); err != nil {
			return err
		}
	} else if err := ctx.Request.ParseForm(); err != nil {
		return err
	}
	for _, values := range []url.Values{ctx.Request.Form, ctx.Request.PostForm} {
		if err := resolveValues(values); err != nil {
			return err
		}
	}
	if ctx.Request.MultipartForm != nil {
		if err := resolveValues(ctx.Request.MultipartForm.Value); err != nil {
			return err
		}
	}

	// Only routing fields are normalized; passwords, properties and other user
	// content must retain their exact bytes and values.
	var body map[string]json.RawMessage
	if !json.Valid(ctx.Input.RequestBody) || json.Unmarshal(ctx.Input.RequestBody, &body) != nil || body == nil {
		return nil
	}
	changed, err := resolveProfileBody(body, resolve)
	if err != nil || !changed {
		return err
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx.Input.RequestBody = data
	ctx.Request.Body = io.NopCloser(bytes.NewReader(data))
	ctx.Request.ContentLength = int64(len(data))
	ctx.Request.Header.Set("Content-Length", fmt.Sprint(len(data)))
	return nil
}

func profileApplicationReference(key, value, path string) (string, bool) {
	name := strings.TrimPrefix(value, "admin/")
	if name != object.CurrentApplicationAlias && name != object.CurrentApplicationAlias+"-org-"+object.CurrentOrganizationAlias {
		return "", false
	}
	return name, key == "application" || key == "applicationId" || key == "id" && strings.Contains(path, "application") && strings.HasPrefix(value, "admin/")
}

func resolveOrganizationAlias(key, value, owner string, organizationObject bool) string {
	switch key {
	case "owner", "userOwner", "organization", "organizationName":
		if value == object.CurrentOrganizationAlias {
			return owner
		}
	case "name":
		if organizationObject && value == object.CurrentOrganizationAlias {
			return owner
		}
	case "id", "userId", "fullFilePath":
		if strings.HasPrefix(value, object.CurrentOrganizationAlias+"/") {
			return owner + strings.TrimPrefix(value, object.CurrentOrganizationAlias)
		}
		if organizationObject && value == "admin/"+object.CurrentOrganizationAlias {
			return "admin/" + owner
		}
		if key == "fullFilePath" {
			return strings.ReplaceAll(value, "/"+object.CurrentOrganizationAlias+"/", "/"+owner+"/")
		}
	}
	return value
}

func resolveProfileBody(body map[string]json.RawMessage, resolve func(string, string) (string, error)) (bool, error) {
	changed := false
	for _, key := range []string{"owner", "userOwner", "organization", "name", "id", "userId", "fullFilePath", "applicationId", "application"} {
		var value string
		if json.Unmarshal(body[key], &value) != nil {
			continue
		}
		resolved, err := resolve(key, value)
		if err != nil {
			return false, err
		}
		if resolved != value {
			body[key], _ = json.Marshal(resolved)
			changed = true
		}
	}
	var user map[string]json.RawMessage
	if json.Unmarshal(body["user"], &user) == nil && user != nil {
		userChanged, err := resolveProfileBody(user, resolve)
		if err != nil {
			return false, err
		}
		if userChanged {
			body["user"], _ = json.Marshal(user)
			changed = true
		}
	}
	return changed, nil
}
