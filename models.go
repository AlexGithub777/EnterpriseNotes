package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

type Cost struct {
	Id             int    `json:"id"`
	ElectricAmount int    `json:"electric_amount"`
	ElectricPrice  int    `json:"electric_price"`
	WaterAmount    int    `json:"water_amount"`
	WaterPrice     int    `json:"water_price"`
	CheckedDate    string `json:"checked_date"`
}

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     int    `json:"role"`
}

func readData(fileName string) ([][]string, error) {
	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	// skip first line as this is the CSV header
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

// import the JSON data into a collection
func (a *App) importData() error {
	log.Printf("Creating tables...")
	// Create table as required, along with attribute constraints
	sql := `DROP TABLE IF EXISTS "cost";
	CREATE TABLE "cost" (
		id INTEGER PRIMARY KEY NOT NULL,
		electric_amount INTEGER,
		electric_price INTEGER,
		water_amount INTEGER,
		water_price INTEGER,
		checked_date DATE
	);`
	_, err := a.db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Table cost table created.")

	sql = `DROP TABLE IF EXISTS "users";
	CREATE TABLE "users" (
		id SERIAL PRIMARY KEY NOT NULL,
		username VARCHAR(255) NOT NULL,
		password VARCHAR(255) NOT NULL,
		role INTEGER DEFAULT 2 NOT NULL
	);
	CREATE UNIQUE INDEX users_by_id ON users (id);`
	_, err = a.db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Table users created.")

	log.Printf("Inserting data...")

	//prepare the cost insert query
	stmt, err := a.db.Prepare("INSERT INTO cost VALUES($1,$2,$3,$4,$5,$6)")
	if err != nil {
		log.Fatal(err)
	}

	// open the CSV file for importing in PG database
	data, err := readData("data/costs.csv")
	if err != nil {
		log.Fatal(err)
	}

	var c Cost
	// prepare the SQL for multiple inserts
	for _, data := range data {
		c.Id, _ = strconv.Atoi(data[0])
		c.ElectricAmount, _ = strconv.Atoi(data[1])
		c.ElectricPrice, _ = strconv.Atoi(data[2])
		c.WaterAmount, _ = strconv.Atoi(data[3])
		c.WaterPrice, _ = strconv.Atoi(data[4])
		c.CheckedDate = data[5]

		_, err := stmt.Exec(c.Id, c.ElectricAmount, c.ElectricPrice, c.WaterAmount, c.WaterPrice, c.CheckedDate)
		if err != nil {
			log.Fatal(err)
		}
	}

	//prepare the users insert query
	stmt, err = a.db.Prepare("INSERT INTO users VALUES($1,$2,$3,$4)")
	if err != nil {
		log.Fatal(err)
	}

	// open the CSV file for importing in PG database
	data, err = readData("data/users.csv")
	if err != nil {
		log.Fatal(err)
	}

	var u User
	// prepare the SQL for multiple inserts
	for _, data := range data {
		u.Id, _ = strconv.Atoi(data[0])
		u.Username = data[1]
		u.Password = data[2]
		u.Role, _ = strconv.Atoi(data[3])
		_, err := stmt.Exec(u.Id, u.Username, u.Password, u.Role)

		if err != nil {
			log.Fatal(err)
		}
	}

	// create temp file to notify data imported
	//can use database directly but this is an example
	// https://golangbyexample.com/touch-file-golang/
	file, err := os.Create("./imported")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	return err
}
