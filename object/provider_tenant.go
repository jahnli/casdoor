// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"net/http"
	"time"

	"github.com/casdoor/casdoor/email"
	"github.com/casdoor/casdoor/notification"
	"github.com/casdoor/casdoor/util"
	notify "github.com/casdoor/notify2"
)

// IsTenantOwner reports an organization whose providers are managed by organization admins.
// Providers owned by "admin" or "built-in" are managed only by global admins.
func IsTenantOwner(owner string) bool {
	return owner != "admin" && owner != "built-in"
}

func isTenantProvider(provider *Provider) bool {
	return IsTenantOwner(provider.Owner)
}

func checkTenantProviderHost(provider *Provider, host string) error {
	if !isTenantProvider(provider) {
		return nil
	}
	return util.CheckInternetHost(host)
}

func getTenantHttpClient(provider *Provider) *http.Client {
	if !isTenantProvider(provider) {
		return nil
	}
	return util.NewInternetOnlyHttpClient(30 * time.Second)
}

func getProviderHttpClient(provider *Provider) *http.Client {
	if client := getTenantHttpClient(provider); client != nil {
		return client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func restrictTenantEmailProvider(provider *Provider, emailProvider email.EmailProvider) {
	switch p := emailProvider.(type) {
	case *email.HttpEmailProvider:
		p.SetHttpClient(getTenantHttpClient(provider))
	case *email.Microsoft365EmailProvider:
		p.SetHttpClient(getTenantHttpClient(provider))
	case *email.MicrosoftGraphEmailProvider:
		p.SetHttpClient(getTenantHttpClient(provider))
	case *email.GmailEmailProvider:
		p.SetHttpClient(getTenantHttpClient(provider))
	}
}

func RestrictNotificationClient(client notify.Notifier, httpClient *http.Client) {
	if httpNotificationClient, ok := client.(*notification.HttpNotificationClient); ok {
		httpNotificationClient.SetHttpClient(httpClient)
	}
}
