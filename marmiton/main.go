package main

import (
	"fmt"
	"strings"
    "net/http"
	"database/sql"
    "github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"ArnaudParant/marmiton/db"
)


type Env struct {
    DB *sql.DB
}


func hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hey! Let's cook"})
}


func (env Env) getRecipes(c *gin.Context) {
	tx := db.BeginWithFunctions(env.DB)
	defer tx.Commit()

    rows, err := tx.Query(db.MakeRecipeQuery(""))
	defer rows.Close()

    switch err {

    case nil:
        recipes := make([]db.Recipe, 0)

        for rows.Next() {
			err, recipe := db.ScanRecipe(rows)

            if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to fetch results"})
                return
            }
            recipes = append(recipes, recipe)
        }

        c.JSON(http.StatusOK, gin.H{"recipes": recipes})

    case sql.ErrNoRows:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Not any recipes found"})

    default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error"})
    }
}


func (env Env) getRecipeByID(c *gin.Context) {
	id := c.Param("id")

	tx := db.BeginWithFunctions(env.DB)
	defer tx.Commit()

	row := tx.QueryRow(db.MakeRecipeQuery("recipes.id = $1"), id)
	err, recipe := db.ScanRecipe(row)

    switch err {

    case nil:
        c.JSON(http.StatusOK, gin.H{"recipe": recipe})

    case sql.ErrNoRows:
		e := fmt.Sprintf("Not found receipe for id: %v", id)
		c.JSON(http.StatusNotFound, gin.H{"message": e})

    default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error"})
    }
}


func (env Env) getRecipeByIngredients(c *gin.Context) {
	ingredients := c.QueryArray("ingredient")

	if ingredients == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You MUST give at least one ingredient"})
		return
	}

	tx := db.BeginWithFunctions(env.DB)
	defer tx.Commit()

	var where_query []string
	var values []any
	for idx, ingredient := range(ingredients) {
		where_query = append(where_query, fmt.Sprintf("ingredient ILIKE $%d", idx+1))
		values = append(values, "%" + ingredient + "%")
	}

	query := fmt.Sprintf("recipes.id IN (SELECT recipe_id FROM ingredients WHERE %s)", strings.Join(where_query, " OR "))


	rows, err := env.DB.Query(db.MakeRecipeQuery(query), values...)
	defer rows.Close()

    switch err {

    case nil:
        recipes := make([]db.Recipe, 0)

        for rows.Next() {
			err, recipe := db.ScanRecipe(rows)

            if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to fetch results"})
                return
            }
            recipes = append(recipes, recipe)
        }

        c.JSON(http.StatusOK, gin.H{"recipes": recipes})

    case sql.ErrNoRows:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Not any recipes found"})

    default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown error"})
    }
}


func main() {
	env := new(Env)
	env.DB = db.Initialize()

    router := gin.Default()
    router.GET("/", hello)
    router.GET("/recipes", env.getRecipes)
    router.GET("/recipes/ingredients", env.getRecipeByIngredients)
    router.GET("/recipes/:id", env.getRecipeByID)

    router.Run()
}
