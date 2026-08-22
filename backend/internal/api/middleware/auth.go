package middleware

import (
	"strings"

	"github.com/aicenter/aicenter/internal/auth"
	"github.com/aicenter/aicenter/internal/database"
	"github.com/aicenter/aicenter/internal/permission"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth verifies a Bearer token and injects the caller's identity into
// gin context so downstream middleware and handlers can read it.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"code": 401, "message": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"code": 401, "message": "Invalid authorization format"})
			c.Abort()
			return
		}

		claims := &auth.Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"code": 401, "message": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequirePermission authorizes the request against the caller's role grants.
//
// It reads the role name that JWTAuth placed in context, checks the
// permission registry (authoritative "which groups own which permission"
// catalog, built at startup from seed) and falls back to the database
// (role_permissions) so that admin-side changes take effect immediately
// without a restart.
//
// superadmin always passes. 403 is returned when the caller is authenticated
// but lacks the permission.
func RequirePermission(permName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(403, gin.H{"code": 403, "message": "Role not found"})
			c.Abort()
			return
		}
		roleName := strings.TrimSpace(role.(string))

		if permission.RegistryInstance().HasPermission(roleName, permName) {
			c.Next()
			return
		}

		if ok := grantedFromDB(roleName, permName); ok {
			c.Next()
			return
		}

		c.JSON(403, gin.H{
			"code":     403,
			"message":  "Forbidden",
			"detail":   "role " + roleName + " does not have permission " + permName,
			"permission": permName,
		})
		c.Abort()
	}
}

func grantedFromDB(roleName, permName string) bool {
	db := database.Get()
	if db == nil {
		return false
	}
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM role_permissions rp
			JOIN roles r ON r.id = rp.role_id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE r.name = ? AND p.name = ?
		)
	`, roleName, permName).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}