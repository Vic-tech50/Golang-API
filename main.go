package main

import (
	"database/sql"
	"errors"
	"fmt"
	"go-crud/routes"

	// "fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	// "golang/routes"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"gte=0"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

var db *sql.DB

func main() {

	// http.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	// })

	// http.ListenAndServe(":8080", nil)
	// Configure using environment variables.
	//
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	if user == "" || host == "" || port == "" {
		log.Fatal("Set DB_USER, DB_HOST and DB_PORT")
	}

	dsn := user + ":" + password + "@tcp(" +
		host + ":" + port + ")/go_crud?parseTime=true"

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Database configuration failed:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot connect to MySQL:", err)
	}

	// Create the table if it does not exist.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(150) NOT NULL,
			description TEXT NOT NULL,
			price DECIMAL(12,2) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				ON UPDATE CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Table initialization failed:", err)
	}

	router := gin.Default()

	routes.RegisterMe(router) //import route on different folder with main.go
	routes.AdminRoute(router) //import route on different folder with main.go
	// routes.EmailRoute(router)     //import route on different folder with main.go
	routes.CrudRoutes(router, db) //import route on different folder with main.go
	routes.RegisterAuthRoutes(router, db)

	//Api Versioning
	{
		v1 := router.Group("/v1")
		routes.EmailRoute(v1)
		routes.ReadWriteFileRoute(v1)
	}

	// 4. Add your protected routes here
	protected := router.Group("/api")

	protected.Use(routes.AuthMiddleware())

	protected.GET("/profile", func(c *gin.Context) {
		userID, _ := c.Get("userID")

		c.JSON(200, gin.H{
			"message": "You are authenticated!",
			"user_id": userID,
		})
	})

	protected.GET("/list", func(c *gin.Context) {
		userID, _ := c.Get("userID")

		c.JSON(200, gin.H{
			"message": "You are authenticated!",
			"user_id": userID,
			"route":   "The List of User",
			"Action":  "Message for the world",
		})
	})
	// RegisterMe(router) //import route in the same folder with main.go

	// router.LoadHTMLGlob("templates/*.html")

	// // router.GET("/show", func(c *gin.Context) {
	// // 	c.HTML(200, "show.html", nil)
	// // 	// c.String(200, "Hello World")
	// // })

	// // router.GET("/view", viewRes)

	// // router.GET("/ping", func(c *gin.Context) {
	// // 	// Return JSON response
	// // 	c.JSON(http.StatusOK, gin.H{
	// // 		"message": "pong",
	// // 	})
	// // })
	router.GET("/products", getProducts)
	router.GET("/products/:id", getProduct)
	router.POST("/products", createProduct)
	router.PUT("/products/:id", updateProduct)
	router.DELETE("/products/:id", deleteProduct)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(router.Run(":8080"))
}

// Convert and validate the ID from the URL.
func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return 0, false
	}
	return id, true
}

func viewRes(c *gin.Context) {

	fmt.Println("Just Testing")
	c.String(200, "This is cool in this place, is okay.Just Testing")

}

// CREATE: POST /products
func createProduct(c *gin.Context) {
	var product Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input. Name and positive price are required; stock cannot be negative.",
		})
		return
	}

	product.Name = strings.TrimSpace(product.Name)
	if product.Name == "" || len(product.Name) > 150 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name must contain 1 to 150 characters",
		})
		return
	}

	result, err := db.Exec(`
		INSERT INTO products (name, description, price, stock)
		VALUES (?, ?, ?, ?)`,
		product.Name, product.Description,
		product.Price, product.Stock,
	)
	if err != nil {
		log.Println("Create product:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create product",
		})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Product created but ID could not be retrieved",
		})
		return
	}

	product.ID = uint64(id)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"data":    product,
	})
}

// READ ALL: GET /products
func getProducts(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, name, description, price, stock,
		       created_at, updated_at
		FROM products
		ORDER BY id DESC
	`)
	if err != nil {
		log.Println("Get products:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve products",
		})
		return
	}
	defer rows.Close()

	products := make([]Product, 0)

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description,
			&p.Price, &p.Stock,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not read product data",
			})
			return
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error reading product records",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Products retrieved successfully",
		"data":    products,
	})
}

// READ ONE: GET /products/:id
func getProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var p Product

	err := db.QueryRow(`
		SELECT id, name, description, price, stock,
		       created_at, updated_at
		FROM products
		WHERE id = ?`, id,
	).Scan(
		&p.ID, &p.Name, &p.Description,
		&p.Price, &p.Stock,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	if err != nil {
		log.Println("Get product:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": p,
	})
}

// UPDATE: PUT /products/:id
func updateProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var product Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input. Name and positive price are required; stock cannot be negative.",
		})
		return
	}

	product.Name = strings.TrimSpace(product.Name)
	if product.Name == "" || len(product.Name) > 150 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name must contain 1 to 150 characters",
		})
		return
	}

	result, err := db.Exec(`
		UPDATE products
		SET name = ?, description = ?, price = ?, stock = ?
		WHERE id = ?`,
		product.Name, product.Description,
		product.Price, product.Stock, id,
	)
	if err != nil {
		log.Println("Update product:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not update product",
		})
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not verify update",
		})
		return
	}

	if affected == 0 {
		// MySQL may report zero when values are unchanged.
		var exists bool
		err := db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)",
			id,
		).Scan(&exists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not verify product",
			})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
			return
		}
	}

	product.ID = id

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"data":    product,
	})
}

// DELETE: DELETE /products/:id
func deleteProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	result, err := db.Exec(
		"DELETE FROM products WHERE id = ?", id,
	)
	if err != nil {
		log.Println("Delete product:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not delete product",
		})
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not verify deletion",
		})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}
