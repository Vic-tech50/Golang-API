
package routes

import (
    "net/http"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        parts := strings.SplitN(header, " ", 2)

        if len(parts) != 2 ||
            !strings.EqualFold(parts[0], "Bearer") ||
            parts[1] == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Missing or invalid authorization token.",
            })
            return
        }

        secret := os.Getenv("JWT_SECRET")
        if len(secret) < 32 {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "error": "Authentication is not configured.",
            })
            return
        }

        claims := &jwt.RegisteredClaims{}

        token, err := jwt.ParseWithClaims(
            parts[1],
            claims,
            func(token *jwt.Token) (interface{}, error) {
                if token.Method != jwt.SigningMethodHS256 {
                    return nil, jwt.ErrSignatureInvalid
                }
                return []byte(secret), nil
            },
        )

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid or expired token.",
            })
            return
        }

        c.Set("userID", claims.Subject)
        c.Next()
    }
}


/* Set your JWT secret
Generate a random secret in PowerShell:


$bytes = New-Object byte[] 32
[Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
$secret = [Convert]::ToBase64String($bytes)
$env:JWT_SECRET = $secret
*/