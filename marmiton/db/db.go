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


func Connect(dbname string) *sql.DB {
	var err error

	connStr := fmt.Sprintf("postgres://postgres:pass123@postgres/%v?sslmode=disable", dbname)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
       panic(err)
   }

	return db
}

func BeginWithFunctions(db *sql.DB) (*sql.Tx) {
    tx, err := db.Begin()

	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`
       -- Create a function that always returns the first non-NULL value:
       CREATE OR REPLACE FUNCTION public.first_agg (anyelement, anyelement)
         RETURNS anyelement
         LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE AS
       'SELECT $1';

       -- Then wrap an aggregate around it:
       CREATE OR REPLACE AGGREGATE public.first (anyelement) (
         SFUNC    = public.first_agg,
         STYPE    = anyelement,
         PARALLEL = safe
       );

       CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;
       `)

	if err != nil {
		panic(err)
	}

	return tx
}


func MakeRecipeQuery(where string) string {
	where_query := ""
	if where != "" {
		where_query = "WHERE " + where
	}

	query_recipes_with_ingredients := fmt.Sprintf(`
       SELECT recipes.id, name, author, budget, difficulty, setup_time, cook_time, total_time,
        array_agg(ingredient ORDER BY index ASC) AS ingredients, people_quantity
          FROM recipes
          LEFT JOIN ingredients AS i ON recipes.id = i.recipe_id
          %s
          GROUP BY recipes.id
    `, where_query)

	final_query := fmt.Sprintf(`
       SELECT ri.id, first(name), first(author), array_agg(tag ORDER BY index ASC) AS tags,
        first(budget), first(difficulty), first(setup_time), first(cook_time), first(total_time),
        first(ingredients), first(people_quantity)
          FROM (%s) AS ri
          LEFT JOIN tags AS t ON ri.id = t.recipe_id
          GROUP BY ri.id
          ORDER BY ri.id ASC;
    `, query_recipes_with_ingredients)

	return final_query
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
	db := Connect("postgres")

	skip := false
	_, err := db.Exec("CREATE DATABASE marmiton")
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			skip = true
		} else {
			db.Close()
			panic(err)
		}
	}

	db.Close()
	db = Connect("marmiton")

	if skip == false {
		execFile(db, "db/schemas.sql")
		insertRecipies(db)
	}

	return db
}


func insertRecipies(db *sql.DB) {
	recipes := loadRecipes("db/recipes-fr2.json")
	sqlRecipe := "INSERT INTO recipes(name,author,budget,difficulty,setup_time,cook_time,total_time,people_quantity) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"
	sqlIngredient := "INSERT INTO ingredients(recipe_id, ingredient, index) VALUES "
	sqlTag := "INSERT INTO tags(recipe_id, tag, index) VALUES "
	total := float64(len(recipes))
	id := 0

	log.Printf("Initializing database ...")

	for rowIdx, recipe := range recipes {
		err := db.QueryRow(sqlRecipe, recipe.Name, recipe.Author, recipe.Budget, recipe.Difficulty, recipe.SetupTime, recipe.CookTime, recipe.TotalTime, recipe.PeopleQuantity).Scan(&id)
		if err != nil {
			panic(err)
		}

		var valuesFmt []string
		var values []any
		values = append(values, id)

		for idx, ingredient := range(recipe.Ingredients) {
			valuesFmt = append(valuesFmt, fmt.Sprintf("($1,$%d,$%d)", idx*2+2, idx*2+3))
			values = append(values, ingredient, idx)
		}

		// log.Printf("Id: %d", id)
		// log.Printf("Ingredients: %s", strings.Join(recipe.Ingredients, ", "))
		// log.Printf("Query: %s", sqlIngredient + strings.Join(valuesFmt, ", "))
		// log.Printf("---")

		_, err = db.Exec(sqlIngredient + strings.Join(valuesFmt, ","), values...)
		if err != nil {
			panic(err)
		}

		valuesFmt = []string{}
		values = []any{id}
		for idx, tag := range(recipe.Tags) {
			valuesFmt = append(valuesFmt, fmt.Sprintf("($1,$%d,$%d)", idx*2+2, idx*2+3))
			values = append(values, tag, idx)
		}

		_, err = db.Exec(sqlTag + strings.Join(valuesFmt, ","), values...)
		if err != nil {
			panic(err)
		}

		if rowIdx % 500 == 0 {
			log.Printf("%.0f%%", float64(rowIdx) / total * 100)
		}

	}

	log.Printf("Database initialized")

}


func execFile(db *sql.DB, filepath string) {

	data := readFile(filepath)
	_, err := db.Exec(strings.Join(data, " "))

	if err != nil {
		panic(err)
	}

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
