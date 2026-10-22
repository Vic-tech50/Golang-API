	package routes
	
	// rows, err := db.Query(
		// 	"SELECT id, name, email, password, created_at FROM users ORDER BY id DESC",
		// )

		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error": "Failed to read users from database",
		// 	})
		// 	return
		// }
		// defer rows.Close()

		// // Create a slice to store users
		// users := []gin.H{}

		// for rows.Next() {
		// 	var id int
		// 	var name string
		// 	var email string
		// 	var password string
		// 	var createdAt string

		// 	//
		// 	err := rows.Scan(&id, &name, &email, &password, &createdAt)
		// 	if err != nil {
		// 		c.JSON(http.StatusInternalServerError, gin.H{
		// 			"error": "Failed to read user data",
		// 		})
		// 		return
		// 	}

		// 	users = append(users, gin.H{
		// 		"id":         id,
		// 		"name":       helpers.Capitalize(name),
		// 		"email":      email,
		// 		"created_at": createdAt,
		// 	})
		// }

		// if err := rows.Err(); err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error": "Error while reading users",
		// 	})
		// 	return
		// }
// http.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	// })

	// http.ListenAndServe(":8080", nil)
	// Configure using environment variables.
	//
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASSWORD")
	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")

	// if user == "" || host == "" || port == "" {
	// 	log.Fatal("Set DB_USER, DB_HOST and DB_PORT")
	// }

	// dsn := user + ":" + password + "@tcp(" +
	// 	host + ":" + port + ")/go_crud?parseTime=true"

	// var err error
	// db, err = sql.Open("mysql", dsn)
	// if err != nil {
	// 	log.Fatal("Database configuration failed:", err)
	// }
	// defer db.Close()

	// if err = db.Ping(); err != nil {
	// 	log.Fatal("Cannot connect to MySQL:", err)
	// };