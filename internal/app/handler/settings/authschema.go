//go:build !minimal

package settings

import appsettings "github.com/damonto/sigmo/internal/pkg/settings"

// authModeOptions lists every selectable authentication mode in the default
// build, including OIDC and Cloudflare Access.
func authModeOptions() []Option {
	return []Option{
		{Label: "settings.authNone", Value: "none"},
		{Label: "settings.authOTP", Value: "otp"},
		{Label: "settings.authOIDC", Value: "oidc"},
		{Label: "settings.authAccess", Value: "cloudflare_access"},
		{Label: "settings.authPassword", Value: "password"},
	}
}

// authSchemaFields returns the authentication settings schema for the default
// build, which includes OIDC and Cloudflare Access configuration.
func authSchemaFields() []Field {
	return []Field{
		{Key: "mode", Label: "settings.schema.auth.mode", Control: controlSelect, Options: authModeOptions()},
		numberField("sessionMaxDays", "settings.schema.auth.sessionMaxDays", "settings.schema.auth.sessionMaxDaysDescription", 1, appsettings.MaxSessionMaxDays, true),
		passwordField("password", "settings.schema.auth.password", "settings.schema.auth.passwordDescription", false),
		{Key: "otpChannels", Label: "settings.schema.auth.otpChannels", Control: controlChannelList},
		textField("oidcIssuer", "settings.schema.auth.oidcIssuer", "", "", false),
		textField("oidcClientId", "settings.schema.auth.oidcClientId", "", "", false),
		passwordField("oidcClientSecret", "settings.schema.auth.oidcClientSecret", "settings.schema.auth.keepSecret", false),
		listField("oidcAllowedSubjects", "settings.schema.auth.oidcAllowedSubjects", "", false),
		textField("accessTeamDomain", "settings.schema.auth.accessTeamDomain", "", "", false),
		textField("accessAudience", "settings.schema.auth.accessAudience", "", "", false),
		listField("accessAllowedSubjects", "settings.schema.auth.accessAllowedSubjects", "", false),
	}
}
