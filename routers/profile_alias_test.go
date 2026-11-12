package routers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
)

func profileAliasContext(method, target, contentType, body string) *context.Context {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	ctx := context.NewContext()
	ctx.Reset(httptest.NewRecorder(), req)
	ctx.Input.RequestBody = []byte(body)
	return ctx
}

func TestResolveProfileApplicationCallback(t *testing.T) {
	for _, shared := range []bool{false, true} {
		application := object.CurrentApplicationAlias
		resolvedApplication := "private-app"
		if shared {
			application += "-org-" + object.CurrentOrganizationAlias
			resolvedApplication += "-org-tenant"
		}
		body, err := json.Marshal(map[string]string{
			"type": "link", "application": application, "state": application, "callbackState": "provider-original-state",
		})
		if err != nil {
			t.Fatal(err)
		}
		ctx := profileAliasContext(http.MethodPost, "/api/login", "application/json", string(body))
		lookup := func(username string) (*object.Application, error) {
			if username != "tenant/alice" {
				t.Fatalf("application lookup used another principal: %q", username)
			}
			return &object.Application{Name: "private-app"}, nil
		}
		if err := resolveProfileAliasesWithApplication(ctx, "tenant/alice", lookup); err != nil {
			t.Fatal(err)
		}
		var result map[string]string
		if err := json.Unmarshal(ctx.Input.RequestBody, &result); err != nil {
			t.Fatal(err)
		}
		if result["application"] != resolvedApplication || result["state"] != application || result["callbackState"] != "provider-original-state" || ctx.Input.GetData("profileApplicationName") != "private-app" {
			t.Fatalf("callback must resolve only its routing target: %v", result)
		}

		ctx = profileAliasContext(http.MethodGet, "/api/get-application?id="+url.QueryEscape("admin/"+application), "", "")
		if err := resolveProfileAliasesWithApplication(ctx, "tenant/alice", lookup); err != nil {
			t.Fatal(err)
		}
		if ctx.Input.Query("id") != "admin/"+resolvedApplication || ctx.Input.GetData("profileApplicationName") != "private-app" {
			t.Fatal("application queries must retain the masking context after alias resolution")
		}
	}
}

func TestResolveProfileAliasBeforeAuthorization(t *testing.T) {
	ctx := profileAliasContext(http.MethodGet, "/api/get-user?id=%40current/bob&owner=another-tenant", "", "")
	if err := resolveProfileAliases(ctx, "tenant/alice"); err != nil {
		t.Fatal(err)
	}
	if ctx.Input.Query("id") != "tenant/bob" || ctx.Input.Query("owner") != "another-tenant" {
		t.Fatal("aliases must use the session tenant, explicit foreign targets must stay unchanged")
	}
	objects, err := getObjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 2 || objects[0].Owner != "tenant" || objects[1].Owner != "another-tenant" {
		t.Fatalf("authorization must check both resolved and explicit targets: %+v", objects)
	}
}

func TestResolveProfileAliasJSON(t *testing.T) {
	body := `{"user":{"owner":"@current","name":"alice"},"password":"@current/keep-this","properties":{"owner":"@current"},"amount":9007199254740993}`
	ctx := profileAliasContext(http.MethodPost, "/api/unlink", "application/json", body)
	if err := resolveProfileAliases(ctx, "tenant/alice"); err != nil {
		t.Fatal(err)
	}
	var result struct {
		User       struct{ Owner, Name string }
		Password   string
		Properties map[string]string
		Amount     int64
	}
	if err := json.Unmarshal(ctx.Input.RequestBody, &result); err != nil {
		t.Fatal(err)
	}
	if result.User.Owner != "tenant" || result.User.Name != "alice" || result.Password != "@current/keep-this" || result.Properties["owner"] != "@current" || result.Amount != 9007199254740993 {
		t.Fatalf("only routing fields may change: %+v", result)
	}
}

func TestResolveProfileAliasForms(t *testing.T) {
	form := url.Values{"owner": {"@current"}, "name": {"alice"}, "secret": {"@current/unchanged"}}
	ctx := profileAliasContext(http.MethodPost, "/api/mfa/setup/enable", "application/x-www-form-urlencoded", form.Encode())
	if err := resolveProfileAliases(ctx, "tenant/alice"); err != nil {
		t.Fatal(err)
	}
	if ctx.Request.Form.Get("owner") != "tenant" || ctx.Request.Form.Get("secret") != "@current/unchanged" {
		t.Fatal("form routing must resolve without changing MFA secrets")
	}

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	if err := writer.WriteField("owner", "@current"); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("name", "alice"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	ctx = profileAliasContext(http.MethodPost, "/api/mfa/setup/initiate", writer.FormDataContentType(), buffer.String())
	if err := resolveProfileAliases(ctx, "tenant/alice"); err != nil {
		t.Fatal(err)
	}
	if ctx.Request.Form.Get("owner") != "tenant" || url.Values(ctx.Request.MultipartForm.Value).Get("owner") != "tenant" {
		t.Fatal("multipart values must use the same tenant for authorization and execution")
	}
}

func TestResolveProfileAliasRequiresSession(t *testing.T) {
	ctx := profileAliasContext(http.MethodGet, "/api/get-user?id=@current/alice", "", "")
	if err := resolveProfileAliases(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if ctx.Input.Query("id") != "@current/alice" {
		t.Fatal("an unauthenticated request must not resolve an alias")
	}
}

func TestResolveOrganizationAliasTargets(t *testing.T) {
	for _, tt := range []struct {
		key, value, want string
		organization     bool
	}{
		{"id", "admin/@current", "admin/tenant", true},
		{"id", "other/alice", "other/alice", false},
		{"name", "@current", "@current", false},
		{"name", "@current", "tenant", true},
		{"fullFilePath", "resource/@current/alice/face.png", "resource/tenant/alice/face.png", false},
		{"password", "@current", "@current", true},
	} {
		if got := resolveOrganizationAlias(tt.key, tt.value, "tenant", tt.organization); got != tt.want {
			t.Errorf("resolve %s=%q: got %q, want %q", tt.key, tt.value, got, tt.want)
		}
	}
}
