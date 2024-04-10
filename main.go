package main

// docker run -it -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres
func main() {
	app := App{}
	app.Initialize(
		"postgres",
		"postgres",
		"postgres")

	app.Run(":8010")
}
