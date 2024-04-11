package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (app *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	app.Router = mux.NewRouter()

	app.initializeRoutes()
}

func (app *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func (app *App) getProduct(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product := product{ID: id}
	if err := product.getProduct(app.DB); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			respondWithError(responseWriter, http.StatusNotFound, "Product not found")
		default:
			respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, product)
}

func respondWithError(responseWriter http.ResponseWriter, code int, message string) {
	respondWithJSON(responseWriter, code, map[string]string{"error": message})
}

func respondWithJSON(responseWriter http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)
	responseWriter.Write(response)
}

func (app *App) getProducts(responseWriter http.ResponseWriter, request *http.Request) {
	count, _ := strconv.Atoi(request.FormValue("count"))
	start, _ := strconv.Atoi(request.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(app.DB, start, count)
	if err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, products)
}

func (app *App) createProduct(responseWriter http.ResponseWriter, request *http.Request) {
	var product product
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&product); err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer request.Body.Close()

	if err := product.createProduct(app.DB); err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusCreated, product)
}

func (app *App) updateProduct(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product product
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&product); err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer request.Body.Close()
	product.ID = id

	if err := product.updateProduct(app.DB); err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, product)
}

func (app *App) deleteProduct(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(responseWriter, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	product := product{ID: id}
	if err := product.deleteProduct(app.DB); err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, map[string]string{"result": "success"})
}

func (app *App) orderProducts(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	field := vars["field"]
	mode := vars["mode"]

	products, err := orderProducts(app.DB, field, mode)
	if err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, products)
}

func (app *App) searchProducts(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	search := vars["search"]

	products, err := searchProducts(app.DB, search)
	if err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, products)
}

func (app *App) initializeRoutes() {
	app.Router.HandleFunc("/products", app.getProducts).Methods("GET")
	app.Router.HandleFunc("/product", app.createProduct).Methods("POST")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.getProduct).Methods("GET")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.updateProduct).Methods("PUT")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.deleteProduct).Methods("DELETE")

	app.Router.HandleFunc("/products/order/{field:(?:name|price)}/{mode:(?:asc|desc)}", app.orderProducts).Methods("GET")
	app.Router.HandleFunc("/products/search/{search:.+}", app.searchProducts).Methods("GET")
}
