package routes

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// auth datatype
type AuthUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// validation
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// generate jwt token
func GenerateToken(userID int) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return "", os.ErrInvalid
	}

	claims := jwt.RegisteredClaims{
		Subject:   strings.TrimSpace(time.Duration(userID).String()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

// logout logic
func Logout(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Logout successful. Please remove your token from the client.",
	})
}

func RegisterAuthRoutes(router *gin.Engine, db *sql.DB) {
	//login api
	router.POST("/login", func(c *gin.Context) {
		var input LoginInput //variable input

		//should accept json data
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Provide a valid email and password.",
				"details": err.Error(),
			})
			return
		}

		// user variable
		var user AuthUser

		// get data from database
		err := db.QueryRow(
			"SELECT id, name, email, password FROM users WHERE email = ?",
			strings.TrimSpace(input.Email),
		).Scan(&user.ID, &user.Name, &user.Email, &user.Password)

		if err != nil {
			// Don't reveal whether the email exists.
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password.",
			})
			return
		}

		// Compare the the unhashed password
		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(input.Password),
		); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid email or password.",
			})
			return
		}

		token, err := GenerateToken(user.ID) //token address
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not generate access token.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Login successful.",
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   900,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
			},
		})
	})

	router.POST("/logout", AuthMiddleware(), Logout)
}
