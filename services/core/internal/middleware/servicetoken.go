// Package middleware provides HTTP middleware for the core service.
package middleware

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
)

// ServiceTokenHeader is the HTTP header first-party services use to
// authenticate service-to-service requests against core.
const ServiceTokenHeader = "X-Service-Token"

// RequireServiceToken returns a Fiber middleware that authenticates the
// caller via the X-Service-Token header. A request is accepted when the
// presented token matches the configured token of any service listed in
// allowed. Comparison is constant-time to prevent timing attacks, and
// token values are never included in logs or error responses.
func RequireServiceToken(tokens map[string]string, allowed ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		presented := c.Get(ServiceTokenHeader)
		if presented == "" {
			return serviceTokenDenied(c)
		}
		for _, name := range allowed {
			expected, ok := tokens[name]
			if !ok || expected == "" {
				continue
			}
			if subtle.ConstantTimeCompare([]byte(presented), []byte(expected)) == 1 {
				// Record the authenticated caller for request logging.
				c.Locals("service_name", name)
				return c.Next()
			}
		}
		return serviceTokenDenied(c)
	}
}

// serviceTokenDenied responds with the standard error shape. The message
// is deliberately generic so callers cannot probe which tokens exist.
func serviceTokenDenied(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error":   "invalid_service_token",
		"message": "A valid service token is required",
	})
}
