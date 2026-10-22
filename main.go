// @title Go CRUD API
// @version 1.0
// @description API documentation for my Go CRUD application.
// @host localhost:8080
// @BasePath /

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"go-crud/config"
	"go-crud/routes"
	"time"

	// "fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"go-crud/models"

	// "golang/routes"

	"github.com/gin-contrib/cors" // Provides CORS middleware for Gin.
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-crud/docs"
)

type Product struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	Stock       int       `json:"stock" binding:"gte=0"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

var db *sql.DB

func RateLimiter() gin.HandlerFunc {
	type client struct {
		limiter *rate.Limiter
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		if _, exists := clients[ip]; !exists {
			// Allow 10 requests per second with a burst of 20
			clients[ip] = &client{limiter: rate.NewLimiter(2, 20)}
		}
		cl := clients[ip]
		mu.Unlock()

		if !cl.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

func main() {

	// Load variables from the .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using environment variables already set")
	}

	// Connect to MySQL through GORM.
	if err := config.Connect(); err != nil {
		log.Fatal(err)
	}

	// Close the database connection when the application exits.
	sqlDB, err := config.DB.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// Create the directory for uploaded files.
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("Could not create uploads directory: ", err)
	}

	// Create or update the products table.
	if err := config.DB.AutoMigrate(&models.StripePayment{}); err != nil {
		log.Fatal("Database migration failed: ", err)
	}

	// 	// Connect to MySQL first.
	// config.ConnectDatabase()

	// // Create the payments table if it does not exist.
	// err := config.DB.AutoMigrate(&models.Payment{})
	// if err != nil {
	// 	log.Fatal("Failed to create payments table:", err)
	// }

	// log.Println("Payments table created successfully")

	router := gin.Default()
	router.Use(RateLimiter()) //import middleware globally

	// Limit in-memory multipart form processing to 8 MB.
	router.MaxMultipartMemory = 8 << 20

	// Serve uploaded files through URLs such as /uploads/image.png.
	router.Static("/uploads", "./uploads")

	// Configure which frontend origins can access your API.
	router.Use(cors.New(cors.Config{
		// Allow requests from your React development server.
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
		},

		// Allow the HTTP methods your API uses.
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		// Allow common headers sent by frontend applications.
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},

		// Allow cookies and credentialed requests if needed.
		// Set this to false if you do not use browser cookies.
		AllowCredentials: true,
	}))

	routes.RegisterMe(router) //import route on different folder with main.go
	routes.AdminRoute(router) //import route on different folder with main.go
	// routes.EmailRoute(router)     //import route on different folder with main.go
	routes.CrudRoutes(router, db) //import route on different folder with main.go
	routes.RegisterAuthRoutes(router, db)
	routes.GetConfigRoute(router)
	routes.SMSRoutes(router)
	routes.RegisterPaymentRoutes(router)
	routes.RegisterFlutterwaveRoutes(router)
	routes.RegisterStripeRoutes(router)

	//Api Versioning
	{
		v1 := router.Group("/v1")
		routes.EmailRoute(v1)
		routes.ReadWriteFileRoute(v1)
		routes.HttpClientRoute(v1)
		routes.CacheRoute(v1)
		routes.CrudFileRoutes(v1)

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

	router.GET("/products", getProducts)
	router.GET("/products/:id", getProduct)
	router.POST("/products", createProduct)
	router.PUT("/products/:id", updateProduct)
	router.DELETE("/products/:id", deleteProduct)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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
	encryptData, err := config.Encrypt(product.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Encryption failed.",
		})
		return
	}
	if product.Name == "" || len(product.Name) > 150 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name must contain 1 to 150 characters",
		})
		return
	}

	result, err := db.Exec(`
		INSERT INTO products (name, description, price, stock)
		VALUES (?, ?, ?, ?)`,
		product.Name, encryptData,
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
		decryptData, err := config.Decrypt(p.Description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not decrypt product description",
			})
			return
		}
		p.Description = decryptData

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
