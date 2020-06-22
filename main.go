package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/model"
)

var db *sql.DB

func createTable() {
	ctb := `
	CREATE TABLE IF NOT EXISTS customer (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`
	_, err := db.Exec(ctb)
	if err != nil {
		log.Fatal("Can't create table", err)
	}
}

func main() {
	var err error

	var dbURL string
	dbURL = os.Getenv("DATABASE_URL")
	log.Println("DATABASE_URL:", dbURL)

	if "" != dbURL {
		log.Println("Initial database connection")
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatal("can't connect to database", err)
		}
		defer db.Close()
		createTable()
	} else {
		log.Println("Can not initial database connection")
	}

	// Init gin
	log.Println("Initial gin middleware")
	r := gin.Default()

	// Configure middleware
	r.Use(configureMiddleware)

	// Configure 5 endpoints
	r.POST("/customers", createCustomerHandler)
	r.GET("/customers/:id", getCustomerByIDHandler)
	r.GET("/customers", getCustomersHandler)
	r.PUT("/customers/:id", updateCustomerByIDHandler)
	r.DELETE("/customers/:id", deleteCustomerByIDHandler)

	// Listening and serving HTTP on 127.0.0.1:2019
	r.Run(":2019")
}

func configureMiddleware(ctx *gin.Context) {
	// Pre Handler
	log.Println("Execute pre middleware")

	authKey := ctx.GetHeader("Authorization")
	if authKey != "token2019" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		ctx.Abort()
		return
	}

	// Execture Handler
	ctx.Next()

	// Post Handler
	log.Println("Execute post middleware")
}

func createCustomerHandler(ctx *gin.Context) {
	var c model.Customer
	err := ctx.ShouldBindJSON(&c)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	row := db.QueryRow("INSERT INTO customer (name, email, status) VALUES ($1, $2, $3) RETURNING id", c.Name, c.Email, c.Status)

	err = row.Scan(&c)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	ctx.JSON(http.StatusCreated, c)
}

func getCustomerByIDHandler(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "id required"})
		return
	}

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer WHERE id=$1")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	row := stmt.QueryRow(id)

	c := model.Customer{}

	err = row.Scan(&c.ID, &c.Name, &c.Email, &c.Status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Data not found"})
		return
	}

	ctx.JSON(http.StatusOK, c)
}

func getCustomersHandler(ctx *gin.Context) {
	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	var customers = []model.Customer{}

	for rows.Next() {
		c := model.Customer{}
		err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Status)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		customers = append(customers, c)
	}

	ctx.JSON(http.StatusOK, customers)
}

func updateCustomerByIDHandler(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "id required"})
		return
	}

	c := model.Customer{}

	err := ctx.ShouldBindJSON(&c)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stmt, err := db.Prepare("UPDATE customer SET name=$2, email=$3, status=$4 WHERE id=$1;")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if _, err := stmt.Exec(id, c.Name, c.Email, c.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.ID = id
	ctx.JSON(http.StatusOK, c)
}

func deleteCustomerByIDHandler(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "id required"})
		return
	}

	stmt, err := db.Prepare("DELETE FROM customer WHERE id=$1")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if _, err := stmt.Exec(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Customer deleted"})
}
