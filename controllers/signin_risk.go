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

package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	signinAssessmentKey  = "signinAssessment"
	signinRiskSessionKey = "signinFlow"
	riskMfaSessionKey    = "flowMfa"
	riskDeniedCode       = "flow-denied"
)

func (c *ApiController) assessSigninRisk(user *object.User) (*object.SigninRisk, bool) {
	if assessment, ok := c.Ctx.Input.GetData(signinAssessmentKey).(*object.SigninRisk); ok {
		return assessment, assessment.Decision == "Deny"
	}
	if user == nil {
		return nil, false
	}
	application, _ := c.Ctx.Input.GetData("signinApplication").(string)
	device := object.CheckSigninDevice(user, application, c.Ctx.Request)
	c.Ctx.Input.SetData("deviceSatisfiesMfa", device.SatisfiesMfa)
	assessment := &object.SigninRisk{Decision: "Allow", Signals: []string{}}
	c.Ctx.Input.SetData(signinAssessmentKey, assessment)
	if device.Denied != "" {
		assessment.Decision = "Deny"
		c.ResponseErrorWithCode(riskDeniedCode, c.T(device.Denied))
		return assessment, true
	}
	result := object.RunSigninFlows(user, application, util.GetClientIpFromRequest(c.Ctx.Request), c.Ctx.Request, object.FlowTriggerPreSignin, nil)
	if result != nil {
		if result.Deny != "" {
			assessment.Decision = "Deny"
			c.ResponseErrorWithCode(riskDeniedCode, result.Deny)
			return assessment, true
		}
		if result.RequireMfa {
			assessment.Decision = "Mfa"
		}
	}
	return assessment, false
}

func (c *ApiController) rememberSigninFlow(user *object.User) {
	application, _ := c.Ctx.Input.GetData("signinApplication").(string)
	c.SetSession(signinRiskSessionKey, fmt.Sprintf("%s|%s|%d", user.GetId(), application, time.Now().Unix()))
}

func (c *ApiController) checkSigninRiskOnLoggedIn(application *object.Application, user *object.User, method string) bool {
	if c.Ctx.Input.Param("sessionSignin") == "true" {
		return false
	}
	verified, _ := c.Ctx.Input.GetData("signinMfaVerified").(bool)
	if verified {
		pending, _ := c.GetSession(signinRiskSessionKey).(string)
		c.SetSession(signinRiskSessionKey, "")
		c.SetSession(riskMfaSessionKey, "")
		parts := strings.Split(pending, "|")
		if len(parts) == 3 && parts[0] == user.GetId() && parts[1] == application.Name {
			timestamp, err := strconv.ParseInt(parts[2], 10, 64)
			age := time.Since(time.Unix(timestamp, 0))
			if err == nil && age >= 0 && age < 15*time.Minute {
				return false
			}
		}
	}
	c.Ctx.Input.SetData("signinApplication", application.Name)
	assessment, denied := c.assessSigninRisk(user)
	if denied {
		return true
	}
	if assessment != nil && assessment.Decision == "Mfa" && !verified {
		organization, err := object.GetOrganizationByUser(user)
		if err != nil {
			c.ResponseError(err.Error())
			return true
		}
		verificationType := ""
		if method == "Magic link" {
			verificationType = object.EmailType
		}
		return checkMfaEnable(c, user, organization, verificationType)
	}
	c.SetSession(signinRiskSessionKey, "")
	c.SetSession(riskMfaSessionKey, "")
	return false
}

// getRiskMfaProps offers the email and the phone of the user for the MFA that a risky sign-in asks for, when the
// user hasn't set up the MFA that the method of the sign-in doesn't already prove
func getRiskMfaProps(user *object.User, verificationType string) []*object.MfaProps {
	props := []*object.MfaProps{}
	if user.Email != "" && verificationType != object.EmailType {
		props = append(props, &object.MfaProps{Enabled: true, MfaType: object.EmailType, Secret: util.GetMaskedEmail(user.Email)})
	}
	if user.Phone != "" && verificationType != object.SmsType {
		props = append(props, &object.MfaProps{Enabled: true, MfaType: object.SmsType, Secret: util.GetMaskedPhone(user.Phone), CountryCode: user.CountryCode})
	}
	return props
}

// getMfaAuthProps returns the MFA of the user that a sign-in verifies, including the email and the phone of a risky
// sign-in's MFA
func (c *ApiController) getMfaAuthProps(user *object.User, mfaType string) *object.MfaProps {
	props := user.GetMfaProps(mfaType, false)
	if props.Enabled {
		return props
	}
	if riskMfa, _ := c.GetSession(riskMfaSessionKey).(string); riskMfa != user.GetId() {
		return props
	}
	if mfaType == object.EmailType && user.Email != "" {
		return &object.MfaProps{Enabled: true, MfaType: mfaType, Secret: user.Email}
	}
	if mfaType == object.SmsType && user.Phone != "" {
		return &object.MfaProps{Enabled: true, MfaType: mfaType, Secret: user.Phone, CountryCode: user.CountryCode}
	}
	return props
}
