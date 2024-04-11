package main

import "os"

// docker run -it -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres
func main() {
	app := App{}

	app.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	app.Run(":8010")
}
