package routes

import (
	"database/sql"
	"go-crud/helpers"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CrudRoutes(router *gin.Engine, db *sql.DB) {

	// Insert submitted form data into MySQL.
	router.POST("/users/insert", func(c *gin.Context) {
		var user User

		// Read JSON from Postman
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON. Please provide name and email.",
			})
			return
		}
		name := strings.TrimSpace(user.Name)        //get name from form
		email := strings.TrimSpace(user.Email)      //get email data
		password := user.Password                   //Don't trim passwords; spaces may be intentional
		hashPassword, err := HashPassword(password) // ignore error for the sake of simplicity

		//to check password hashing
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to secure password",
			})
			return
		}

		// name := strings.TrimSpace(c.PostForm("name"))   //get name from form
		// email := strings.TrimSpace(c.PostForm("email")) //get email data

		// name := strings.TrimSpace("name")   //get name from form
		// email := strings.TrimSpace("email") //get email data

		// Validate the input.
		if name == "" || email == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Name, Password and email are required.",
			})
			return
		}

		if len(name) > 100 || len(email) > 150 ||
			!strings.Contains(email, "@") {
			// c.HTML(http.StatusBadRequest, "insert.html", gin.H{
			// 	"Error": "Enter a valid name and email address.",
			// })
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Enter a valid name and email address.",
			})
			return
		}

		// Insert into the database safely using placeholders.
		result, err := db.Exec(
			"INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
			name, email, hashPassword,
		)

		if err != nil {
			log.Println("Insert user failed:", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Could not save user. The email may already exist.",
			})
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Println("Could not retrieve inserted ID:", err)
		}

		//send Success Message
		// c.HTML(http.StatusOK, "insert.html", gin.H{
		// 	"Success": "User saved successfully!",
		// 	"ID":      id,
		// })

		c.JSON(http.StatusOK, gin.H{
			"Success": "User saved successfully!",
			"ID":      id,
		})
	})

	// READ ALL: GET /users
	router.GET("/users", func(c *gin.Context) {
		// Read all users from the database
		rows, err := db.Query(
			"SELECT id, name, email, password, created_at FROM users ORDER BY id DESC",
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read users from database",
			})
			return
		}
		defer rows.Close()

		// Create a slice to store users
		users := []gin.H{}

		for rows.Next() {
			var id int
			var name string
			var email string
			var password string
			var createdAt string

			//
			err := rows.Scan(&id, &name, &email, &password, &createdAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to read user data",
				})
				return
			}

			users = append(users, gin.H{
				"id":         id,
				"name":       helpers.Capitalize(name),
				"email":      email,
				"created_at": createdAt,
			})
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error while reading users",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"count":   len(users),
			"users":   users,
		})
	})

	//get a single user
	router.GET("/users/:id", func(c *gin.Context) {
		// Get the user ID from the URL
		id := c.Param("id")

		// Find the user in the database
		// var user struct {
		// 	ID        int    `json:"id"`
		// 	Name      string `json:"name"`
		// 	Email     string `json:"email"`
		// 	CreatedAt string `json:"created_at"`
		// }
		var user User

		err := db.QueryRow("SELECT id, name, email, created_at FROM users WHERE id = ?", id).Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve user",
			})
			return
		}

		// Return the user as JSON
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user":    user,
		})
	})

	/*  Update Logic   */
	router.PUT("/users/:id", func(c *gin.Context) {
		// Get the user ID from the URL
		id := c.Param("id")

		// Receive JSON data from Postman
		var user User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON. Provide name and email.",
			})
			return
		}

		// Remove unnecessary spaces
		name := strings.TrimSpace(user.Name)
		email := strings.TrimSpace(user.Email)

		// Validate input
		if name == "" || email == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Name and email are required.",
			})
			return
		}

		if len(name) > 100 || len(email) > 150 ||
			!strings.Contains(email, "@") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Enter a valid name and email address.",
			})
			return
		}

		// Update the database
		result, err := db.Exec(
			"UPDATE users SET name = ?, email = ? WHERE id = ?",
			name, email, id,
		)

		if err != nil {
			log.Println("Update user failed:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not update user. The email may already exist.",
			})
			return
		}

		// Check whether the user exists
		affected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not verify the update.",
			})
			return
		}

		if affected == 0 {
			// Check if the ID exists, since unchanged values can
			// also result in zero affected rows in MySQL.
			var exists bool
			err := db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)",
				id,
			).Scan(&exists)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Could not verify whether the user exists.",
				})
				return
			}

			if !exists {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User not found.",
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User updated successfully.",
			"user": gin.H{
				"id":    id,
				"name":  name,
				"email": email,
			},
		})
	})

	/*  DELETE Logic   */
	router.DELETE("/users/:id", func(c *gin.Context) {
		// Get the user ID from the URL
		id := c.Param("id")

		// Delete the user from the database
		result, err := db.Exec(
			"DELETE FROM users WHERE id = ?",
			id,
		)

		if err != nil {
			log.Println("Delete user failed:", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete user.",
			})
			return
		}

		// Check whether a user was deleted
		affected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not verify deletion.",
			})
			return
		}

		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found.",
			})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User deleted successfully.",
			"id":      id,
		})
	})

}
