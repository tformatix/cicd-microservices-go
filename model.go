package main

import (
	"database/sql"
)

type product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (product *product) getProduct(db *sql.DB) error {
	return db.QueryRow("SELECT name, price FROM products WHERE id=$1",
		product.ID).Scan(&product.Name, &product.Price)
}

func (product *product) updateProduct(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE products SET name=$1, price=$2 WHERE id=$3",
			product.Name, product.Price, product.ID)

	return err
}

func (product *product) deleteProduct(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM products WHERE id=$1", product.ID)

	return err
}

func (product *product) createProduct(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO products(name, price) VALUES($1, $2) RETURNING id",
		product.Name, product.Price).Scan(&product.ID)

	if err != nil {
		return err
	}

	return nil
}

func getProducts(db *sql.DB, start, count int) ([]product, error) {
	rows, err := db.Query(
		"SELECT id, name, price FROM products LIMIT $1 OFFSET $2",
		count, start)

	return sqlRowsToProducts(rows, err)
}

func searchProducts(db *sql.DB, search string) ([]product, error) {
	rows, err := db.Query(
		"SELECT id, name, price FROM products WHERE name ILIKE $1", "%"+search+"%")

	return sqlRowsToProducts(rows, err)
}

func orderProducts(db *sql.DB, field string, mode string) ([]product, error) {
	var query string

	switch field {
	case "name":
		switch mode {
		case "desc":
			query = "SELECT id, name, price FROM products ORDER BY name DESC"
		default:
			query = "SELECT id, name, price FROM products ORDER BY name ASC"
		}
	default:
		switch mode {
		case "desc":
			query = "SELECT id, name, price FROM products ORDER BY price DESC"
		default:
			query = "SELECT id, name, price FROM products ORDER BY price ASC"
		}
	}

	rows, err := db.Query(query)

	return sqlRowsToProducts(rows, err)
}

func sqlRowsToProducts(rows *sql.Rows, err error) ([]product, error) {
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := []product{}

	for rows.Next() {
		var p product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}
