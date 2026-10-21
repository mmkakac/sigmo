//go:build minimal

package settings

import appsettings "github.com/damonto/sigmo/internal/pkg/settings"

// authModeOptions omits OIDC and Cloudflare Access, which are not compiled
// into minimal builds.
func authModeOptions() []Option {
	return []Option{
		{Label: "settings.authNone", Value: "none"},
		{Label: "settings.authPassword", Value: "password"},
		{Label: "settings.authOTP", Value: "otp"},
	}
}

// authSchemaFields omits the OIDC and Cloudflare Access settings in minimal
// builds so the web UI hides them automatically.
func authSchemaFields() []Field {
	return []Field{
		{Key: "mode", Label: "settings.schema.auth.mode", Control: controlSelect, Options: authModeOptions()},
		numberField("sessionMaxDays", "settings.schema.auth.sessionMaxDays", "settings.schema.auth.sessionMaxDaysDescription", 1, appsettings.MaxSessionMaxDays, true),
		passwordField("password", "settings.schema.auth.password", "settings.schema.auth.passwordDescription", false),
		{Key: "otpChannels", Label: "settings.schema.auth.otpChannels", Control: controlChannelList},
	}
}
