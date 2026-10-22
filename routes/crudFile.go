package routes

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-crud/config"
	"go-crud/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Maximum allowed upload size is 5 MB.
const maxImageSize = 5 << 20

// Only allow these image extensions.
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// saveImage validates and stores an uploaded image.
func saveImage(c *gin.Context) (string, error) {
	// Retrieve the image from the multipart form.
	file, err := c.FormFile("image")
	if err != nil {
		return "", err
	}

	// Reject files that exceed the upload limit.
	if file.Size > maxImageSize {
		return "", fmt.Errorf("image must not exceed 5 MB")
	}

	// Validate the file extension.
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("only JPG, JPEG, PNG, and WEBP images are allowed")
	}

	// Open the uploaded file to inspect its content.
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Read the first bytes to identify the actual file type.
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && n == 0 {
		return "", fmt.Errorf("could not read image")
	}

	// Ensure the content type matches the extension.
	contentType := http.DetectContentType(buffer[:n])
	validType := false

	switch contentType {
	case "image/jpeg":
		validType = ext == ".jpg" || ext == ".jpeg"
	case "image/png":
		validType = ext == ".png"
	case "image/webp":
		validType = ext == ".webp"
	}

	if !validType {
		return "", fmt.Errorf("file content does not match an allowed image type")
	}

	// Generate a unique filename.
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	filename := hex.EncodeToString(randomBytes) + ext

	// Create the uploads directory if it does not exist.
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", err
	}

	// Save the uploaded file to disk.
	fullPath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		return "", err
	}

	// Return a relative URL path for the database.
	return "/" + filepath.ToSlash(fullPath), nil
}

// removeImage deletes an image stored inside uploads/.
func removeImage(imagePath string) {
	// Only process paths belonging to the uploads directory.
	if !strings.HasPrefix(imagePath, "/uploads/") {
		return
	}

	// Extract the filename to prevent directory traversal.
	filename := filepath.Base(imagePath)
	if filename == "." || filename == "/" {
		return
	}

	// Remove the local image file.
	_ = os.Remove(filepath.Join("uploads", filename))
}

// CrudFileRoutes registers all file CRUD routes.
func CrudFileRoutes(v1 *gin.RouterGroup) {

	// CREATE: POST /crudfile/insert
	v1.POST("/crudfile/insert", func(c *gin.Context) {
		// Read product fields from multipart/form-data.
		name := strings.TrimSpace(c.PostForm("name"))
		description := c.PostForm("description")

		// Validate the price.
		price, err := strconv.ParseFloat(c.PostForm("price"), 64)
		if err != nil || price < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "price must be a valid non-negative number",
			})
			return
		}

		// Validate the stock quantity.
		stock, err := strconv.Atoi(c.PostForm("stock"))
		if err != nil || stock < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "stock must be a non-negative integer",
			})
			return
		}

		// Require a product name.
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "name is required",
			})
			return
		}

		// Save an image only when one is provided.
		imagePath := ""
		if _, err := c.FormFile("image"); err == nil {
			imagePath, err = saveImage(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		// Build the product model.
		product := models.Product{
			Name:        name,
			Description: description,
			Price:       price,
			Stock:       stock,
			ImagePath:   imagePath,
		}

		// Insert the product into the database.
		if err := config.DB.Create(&product).Error; err != nil {
			removeImage(imagePath)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not create product",
			})
			return
		}

		// Return the created product.
		c.JSON(http.StatusCreated, gin.H{
			"message": "Product created successfully",
			"product": product,
		})
	})

	v1.GET("/crudfile", func(c *gin.Context) {
		var products []models.Product

		// Retrieve products in descending ID order.
		if err := config.DB.Order("id DESC").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not retrieve products",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"products": products})
	})

	// READ ONE: GET /crudfile/:id
	v1.GET("/crudfile/:id", func(c *gin.Context) {
		var product models.Product

		// Find the product by ID.
		if err := config.DB.First(&product, c.Param("id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not retrieve product",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"product": product})
	})

	// UPDATE: PUT /crudfile/:id
	v1.PUT("/crudfile/:id", func(c *gin.Context) {
		var product models.Product

		// Find the existing product before updating.
		if err := config.DB.First(&product, c.Param("id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not retrieve product",
			})
			return
		}

		// Update the name only if it was submitted.
		if value, exists := c.GetPostForm("name"); exists {
			value = strings.TrimSpace(value)
			if value == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "name cannot be empty",
				})
				return
			}
			product.Name = value
		}

		// Update the description only if it was submitted.
		if value, exists := c.GetPostForm("description"); exists {
			product.Description = value
		}

		// Validate and update the price if supplied.
		if value, exists := c.GetPostForm("price"); exists {
			price, err := strconv.ParseFloat(value, 64)
			if err != nil || price < 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "price must be a valid non-negative number",
				})
				return
			}
			product.Price = price
		}

		// Validate and update the stock if supplied.
		if value, exists := c.GetPostForm("stock"); exists {
			stock, err := strconv.Atoi(value)
			if err != nil || stock < 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "stock must be a non-negative integer",
				})
				return
			}
			product.Stock = stock
		}

		// Keep the existing image unless a replacement is supplied.
		oldImagePath := product.ImagePath
		if _, err := c.FormFile("image"); err == nil {
			newImagePath, err := saveImage(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			product.ImagePath = newImagePath
		}

		// Save the product changes.
		if err := config.DB.Save(&product).Error; err != nil {
			if product.ImagePath != oldImagePath {
				removeImage(product.ImagePath)
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not update product",
			})
			return
		}

		// Remove the old image only after a successful update.
		if oldImagePath != product.ImagePath {
			removeImage(oldImagePath)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Product updated successfully",
			"product": product,
		})
	})

	// DELETE: DELETE /crudfile/:id
	v1.DELETE("/crudfile/:id", func(c *gin.Context) {
		var product models.Product

		// Find the product to delete.
		if err := config.DB.First(&product, c.Param("id")).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not retrieve product",
			})
			return
		}

		// Delete the database record.
		if err := config.DB.Delete(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not delete product",
			})
			return
		}

		// Delete the associated image from disk.
		removeImage(product.ImagePath)

		c.JSON(http.StatusOK, gin.H{
			"message": "Product deleted successfully",
		})
	})
}
