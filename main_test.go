package main

import (
	"log"
	"os"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
)

var app App

func TestMain(m *testing.M) {
	app.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := app.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	app.DB.Exec("DELETE FROM products")
	app.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {

	clearTable()

	var jsonStr = []byte(`{"name":"test product", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is app map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		app.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func TestUpdateProduct(t *testing.T) {

	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func getProductsFromBytes(bytes []byte) []map[string]interface{} {
	var arr []map[string]interface{}
	_ = json.Unmarshal(bytes, &arr)
	return arr
}

func checkLengthOfProducts(t *testing.T, body []byte, expected int) {
	count := len(getProductsFromBytes(body))
	if count != expected {
		t.Errorf("Expected an array of size %d. Got %d", expected, count)
	}
}

func TestSearchProducts(t *testing.T) {
	clearTable()
	addProducts(4)

	req, _ := http.NewRequest("GET", "/products/search/product", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	checkLengthOfProducts(t, response.Body.Bytes(), 4)

	req, _ = http.NewRequest("GET", "/products/search/product 1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	checkLengthOfProducts(t, response.Body.Bytes(), 1)

	req, _ = http.NewRequest("GET", "/products/search/product 6", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	checkLengthOfProducts(t, response.Body.Bytes(), 0)
}

func checkNamesOfProducts(t *testing.T, body []byte, expected []string) {
	products := getProductsFromBytes(body)
	for i, expectedField := range expected {
		actualField := products[i]["name"]
		if actualField != expectedField {
			t.Errorf("Expected product name to be '%s'. Got '%s'", expectedField, actualField)
		}
	}
}

func TestOrderProducts(t *testing.T) {
	clearTable()
	addProducts(5)

	req, _ := http.NewRequest("GET", "/products/order/nam/as", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	req, _ = http.NewRequest("GET", "/products/order/name/asc", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	checkNamesOfProducts(t, response.Body.Bytes(), []string{"Product 0", "Product 1", "Product 2", "Product 3", "Product 4"})

	req, _ = http.NewRequest("GET", "/products/order/name/desc", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	checkNamesOfProducts(t, response.Body.Bytes(), []string{"Product 4", "Product 3", "Product 2", "Product 1", "Product 0"})

	req, _ = http.NewRequest("GET", "/products/order/price/asc", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	checkNamesOfProducts(t, response.Body.Bytes(), []string{"Product 0", "Product 1", "Product 2", "Product 3", "Product 4"})

	req, _ = http.NewRequest("GET", "/products/order/price/desc", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	checkNamesOfProducts(t, response.Body.Bytes(), []string{"Product 4", "Product 3", "Product 2", "Product 1", "Product 0"})
}
