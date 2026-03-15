package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"profile-services/controllers.go"
	"profile-services/db"

	_ "github.com/lib/pq"
)

func main() {
	db, err := db.InitDB()
	if err != nil {
		slog.Error("database connection refused", "error", err)
		os.Exit(66)
	}

	c := controllers.NewSession(db)

	http.HandleFunc("/", c.GetUser)
	http.HandleFunc("/", c.UpdateUser)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
