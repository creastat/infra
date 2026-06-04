package middleware

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth returns a gin.HandlerFunc that validates a JWT token.
// It extracts "tenant_id" / "tid" and "sub" (user_id) from the claims
// and sets them in the gin context.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := parseBearer(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Extract tenant_id and user_id/sub
		if tid, ok := claims["tenant_id"].(string); ok && tid != "" {
			c.Set("tenant_id", tid)
		}
		if tid, ok := claims["tid"].(string); ok && tid != "" {
			c.Set("tenant_id", tid)
		}
		if sub, ok := claims["sub"].(string); ok {
			c.Set("user_id", sub)
		}

		c.Next()
	}
}

// ServiceAuth returns a gin.HandlerFunc that validates a service-to-service JWT token.
// It checks if the token has the required scopes.
func ServiceAuth(secret string, requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := parseBearer(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired service token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Extract scopes
		var tokenScopes []string
		if sc, ok := claims["scopes"].([]interface{}); ok {
			for _, s := range sc {
				if str, ok := s.(string); ok {
					tokenScopes = append(tokenScopes, str)
				}
			}
		}

		// Verify required scopes
		for _, reqScope := range requiredScopes {
			found := false
			for _, tokenScope := range tokenScopes {
				if tokenScope == reqScope {
					found = true
					break
				}
			}
			if !found {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Missing required scope: %s", reqScope)})
				return
			}
		}

		// Set service name and tenant_id in context
		if sub, ok := claims["sub"].(string); ok {
			c.Set("service_name", sub)
		}
		if tid, ok := claims["tenant_id"].(string); ok && tid != "" {
			c.Set("tenant_id", tid)
		}
		if tid, ok := claims["tid"].(string); ok && tid != "" {
			c.Set("tenant_id", tid)
		}

		c.Next()
	}
}

type InternalVerifierConfig struct {
	HMACSecret          string
	Ed25519PublicKeyB64 string
	TrustedIssuers      []string
}

// RequireInternal validates short-lived internal JWTs and enforces audience/scopes.
// It reads verification settings from environment:
//   - INTERNAL_JWT_HS_SECRET (or JWT_SECRET fallback)
//   - INTERNAL_JWT_PUBLIC_KEY_B64
//   - INTERNAL_JWT_TRUSTED_ISSUERS (comma-separated, optional)
func RequireInternal(audience string, scopes ...string) gin.HandlerFunc {
	trustedIssuers := splitCSV(os.Getenv("INTERNAL_JWT_TRUSTED_ISSUERS"))
	if len(trustedIssuers) == 0 {
		trustedIssuers = []string{"edge-api", "identity-service"}
	}
	secret := strings.TrimSpace(os.Getenv("INTERNAL_JWT_HS_SECRET"))
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("JWT_SECRET"))
	}
	cfg := InternalVerifierConfig{
		HMACSecret:          secret,
		Ed25519PublicKeyB64: strings.TrimSpace(os.Getenv("INTERNAL_JWT_PUBLIC_KEY_B64")),
		TrustedIssuers:      trustedIssuers,
	}
	return RequireInternalWithConfig(cfg, audience, scopes...)
}

func RequireInternalWithConfig(cfg InternalVerifierConfig, audience string, scopes ...string) gin.HandlerFunc {
	var pub ed25519.PublicKey
	if cfg.Ed25519PublicKeyB64 != "" {
		raw, err := base64.StdEncoding.DecodeString(cfg.Ed25519PublicKeyB64)
		if err == nil && len(raw) == ed25519.PublicKeySize {
			pub = ed25519.PublicKey(raw)
		}
	}
	requiredScopes := normalizeScopes(scopes)
	audience = strings.TrimSpace(audience)

	return func(c *gin.Context) {
		tokenString, ok := parseBearer(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			switch token.Method.(type) {
			case *jwt.SigningMethodHMAC:
				if strings.TrimSpace(cfg.HMACSecret) == "" {
					return nil, fmt.Errorf("internal hmac secret is empty")
				}
				return []byte(cfg.HMACSecret), nil
			case *jwt.SigningMethodEd25519:
				if len(pub) == 0 {
					return nil, fmt.Errorf("internal ed25519 public key is empty")
				}
				return pub, nil
			default:
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		})
		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired internal token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}
		if audience != "" && !claimHasAudience(claims, audience) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid token audience"})
			return
		}
		if len(cfg.TrustedIssuers) > 0 {
			issuer, _ := claims["iss"].(string)
			if !stringInSet(issuer, cfg.TrustedIssuers) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "untrusted token issuer"})
				return
			}
		}

		tokenScopes := extractScopeList(claims["scopes"])
		for _, reqScope := range requiredScopes {
			if !stringInSet(reqScope, tokenScopes) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Missing required scope: %s", reqScope)})
				return
			}
		}

		sub, _ := claims["sub"].(string)
		if sub != "" {
			c.Set("service_name", sub)
		}
		if tenantID := tenantFromClaims(claims); tenantID != "" {
			c.Set("tenant_id", tenantID)
		}
		if actorType := actorTypeFromClaims(claims); actorType != "" {
			c.Set("actor_type", actorType)
			if actorType == "user" && sub != "" {
				c.Set("user_id", sub)
			}
		}
		c.Set("token_scopes", tokenScopes)
		c.Set("internal_claims", claims)
		c.Next()
	}
}

func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, ok := TenantIDFromContext(c)
		if !ok || strings.TrimSpace(tenantID) == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tenant context required"})
			return
		}
		c.Next()
	}
}

func RequireActorType(expected string) gin.HandlerFunc {
	expected = strings.TrimSpace(expected)
	return func(c *gin.Context) {
		actor, ok := ActorFromContext(c)
		if !ok || actor != expected {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "actor type not allowed"})
			return
		}
		c.Next()
	}
}

func ActorFromContext(c *gin.Context) (string, bool) {
	raw, ok := c.Get("actor_type")
	if !ok {
		return "", false
	}
	value, ok := raw.(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func TenantIDFromContext(c *gin.Context) (string, bool) {
	raw, ok := c.Get("tenant_id")
	if !ok {
		return "", false
	}
	value, ok := raw.(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func parseBearer(header string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

func extractScopeList(raw any) []string {
	switch v := raw.(type) {
	case []string:
		return normalizeScopes(v)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return normalizeScopes(out)
	case string:
		return normalizeScopes(strings.Fields(v))
	default:
		return nil
	}
}

func normalizeScopes(scopes []string) []string {
	out := make([]string, 0, len(scopes))
	seen := map[string]struct{}{}
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return out
}

func claimHasAudience(claims jwt.MapClaims, expected string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return true
	}
	if aud, ok := claims["aud"].(string); ok {
		return aud == expected
	}
	if audList, ok := claims["aud"].([]any); ok {
		for _, item := range audList {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	}
	if audList, ok := claims["aud"].([]string); ok {
		for _, aud := range audList {
			if aud == expected {
				return true
			}
		}
	}
	return false
}

func tenantFromClaims(claims jwt.MapClaims) string {
	if tid, _ := claims["tenant_id"].(string); tid != "" {
		return tid
	}
	if tid, _ := claims["tid"].(string); tid != "" {
		return tid
	}
	return ""
}

func actorTypeFromClaims(claims jwt.MapClaims) string {
	if act, ok := claims["act"].(map[string]any); ok {
		if typ, _ := act["type"].(string); typ != "" {
			return typ
		}
	}
	if act, ok := claims["act"].(string); ok && act != "" {
		return act
	}
	return ""
}

func stringInSet(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
