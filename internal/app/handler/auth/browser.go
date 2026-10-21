//go:build !minimal

package auth

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/damonto/sigmo/internal/app/auth"
	"github.com/damonto/sigmo/internal/app/httpapi"
	"github.com/damonto/sigmo/internal/pkg/settings"
	"github.com/damonto/sigmo/internal/pkg/storage"
)

// CreateBrowserSession creates a bounded browser login transaction.
func (h *Handler) CreateBrowserSession(c *echo.Context) error {
	if allowed, err := h.allow(c); !allowed {
		return err
	}
	var req struct {
		Client        string `json:"client"`
		BasePath      string `json:"basePath"`
		State         string `json:"state"`
		CodeChallenge string `json:"codeChallenge"`
	}
	if err := httpapi.BindJSON(c, &req, "invalid_browser_request"); err != nil {
		return err
	}
	if req.Client != "web" && req.Client != "native" {
		return h.loginError(c, auth.ErrInvalidLogin)
	}
	target, err := h.store.CreateBrowserLogin(auth.BrowserRequest{
		Policy: h.settings.Snapshot().Auth, Origin: browserOrigin(c), BasePath: req.BasePath,
		Native: req.Client == "native", State: req.State, CodeChallenge: req.CodeChallenge,
	})
	if err != nil {
		return h.loginError(c, err)
	}
	return c.JSON(http.StatusCreated, struct {
		AuthorizationURL string `json:"authorizationUrl"`
	}{target})
}

// StartBrowserSession establishes the browser cookie before visiting the identity provider.
func (h *Handler) StartBrowserSession(c *echo.Context) error {
	if allowed, err := h.allow(c); !allowed {
		return err
	}
	result, err := h.store.StartBrowserLogin(c.Request().Context(), auth.BrowserStartRequest{
		Policy: h.settings.Snapshot().Auth, ID: c.Param("id"), Origin: browserOrigin(c),
	})
	if err != nil {
		return h.loginError(c, err)
	}
	c.SetCookie(&http.Cookie{Name: "sigmo_login_" + result.State, Value: result.Binding, Path: "/", MaxAge: 300, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	c.Response().Header().Set("Referrer-Policy", "no-referrer")
	return c.Redirect(http.StatusSeeOther, result.URL)
}

// OIDCCallback never forwards provider tokens to a browser or native app.
func (h *Handler) OIDCCallback(c *echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("Referrer-Policy", "no-referrer")
	query := c.Request().URL.Query()
	state := query.Get("state")
	if len(query["state"]) != 1 || len(query["code"]) != 1 || query.Has("error") || len(state) > 128 {
		return h.loginError(c, auth.ErrInvalidLogin)
	}
	cookie, err := c.Cookie("sigmo_login_" + state)
	if err != nil {
		return h.loginError(c, auth.ErrInvalidLogin)
	}
	c.SetCookie(&http.Cookie{Name: cookie.Name, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	cfg := h.settings.Snapshot().Auth
	result, err := h.store.CompleteBrowserLogin(c.Request().Context(), auth.BrowserCallback{
		Policy: cfg, Origin: browserOrigin(c), State: state, Code: c.QueryParam("code"), Binding: cookie.Value,
	})
	if err != nil {
		return h.loginError(c, err)
	}
	if h.settings.Snapshot().Auth.PolicyID() != cfg.PolicyID() {
		return h.loginError(c, auth.ErrInvalidLogin)
	}
	if result.Native {
		return c.Redirect(http.StatusSeeOther, result.RedirectURL)
	}
	tokens, err := h.store.IssueSession(c.Request().Context(), auth.SessionIssue{Session: result.Identity})
	if err != nil {
		return httpapi.Internal(c, "create_session_failed", err)
	}
	setSessionCookie(c, tokens)
	return c.Redirect(http.StatusSeeOther, result.RedirectURL)
}

func browserOrigin(c *echo.Context) string {
	// Echo v5 accepts scheme headers only from trusted peers. Proxies must preserve
	// Host and overwrite X-Forwarded-Proto; forwarded host headers are not used.
	return c.Scheme() + "://" + c.Request().Host
}

// resolveOIDCHandoff redeems an app-specific OIDC handoff code. It exists only
// in non-minimal builds, where OIDC login is compiled.
func resolveOIDCHandoff(ctx context.Context, store *auth.Store, cfg settings.Auth, code, verifier string) (storage.AuthSession, error) {
	return store.CompleteHandoff(auth.HandoffRequest{Policy: cfg, Code: code, Verifier: verifier})
}
