package db

import (
	"os"
	"fmt"
	"log"
	"bufio"
	"strings"
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
)


type Recipe struct {
    ID              string   `json:"id"`
    Name            string   `json:"name"`
    Author          string   `json:"author"`
    Tags            []string `json:"tags"`
    Budget          string   `json:"budget"`
    SetupTime       string   `json:"setup_time"`
    CookTime        string   `json:"cook_time"`
    TotalTime       string   `json:"total_time"`
    Difficulty      string   `json:"difficulty"`
    PeopleQuantity  string   `json:"people_quantity"`
    Ingredients     []string `json:"ingredients"`
}


type Scanner interface {
	Scan(...any) error
}


func Connect(dbname string) (*sql.DB, error) {
	var err error

	connStr := fmt.Sprintf("postgres://postgres:pass123@postgres/%v?sslmode=disable", dbname)
	db, err := sql.Open("postgres", connStr)

	return db, err
}


func ScanRecipe(row Scanner) (error, Recipe) {
	var id, name, author, budget, difficulty string
	var setup_time, cook_time, total_time, people_quantity string
	var tags, ingredients []string

	err := row.Scan(&id, &name, &author, pq.Array(&tags), &budget, &setup_time, &cook_time, &total_time, &difficulty, pq.Array(&ingredients), &people_quantity)

	if err != nil {
		log.Printf("Error occurred while reading the database rows: %v", err)
	}

	recipe := Recipe{id, name, author, tags, budget, setup_time, cook_time, total_time, difficulty, people_quantity, ingredients}

	return err, recipe
}


func Initialize() *sql.DB {
	db, err := Connect("postgres")
	if err != nil {
       panic(err)
   }

	skip := false
	_, err = db.Exec("CREATE DATABASE marmiton")
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			skip = true
		} else {
			db.Close()
			panic(err)
		}
	}

	db.Close()
	db, err = Connect("marmiton")
	if err != nil {
       panic(err)
   }

	if skip == false {

		schema := readFile("db/schema.sql")
		_, err = db.Exec(strings.Join(schema, " "))
		if err != nil {
			panic(err)
		}

		loadRecipes("db/recipes-fr2.json")

	}

	return db
}


func readFile(filepath string) []string {

	readFile, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer readFile.Close()

    fileScanner := bufio.NewScanner(readFile)
    fileScanner.Split(bufio.ScanLines)
    var fileLines []string

    for fileScanner.Scan() {
        fileLines = append(fileLines, fileScanner.Text())
    }

	return fileLines
}


func loadRecipes(filepath string) []Recipe {
	readFile, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer readFile.Close()

	decoder := json.NewDecoder(readFile)

	var recipes []Recipe

	for decoder.More() {
		var recipe Recipe
		err := decoder.Decode(&recipe)

		if err != nil {
			log.Printf("parse recipe: %w", err)
		}

		recipes = append(recipes, recipe)
	}

	return recipes
}
