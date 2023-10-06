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
    rows, err := env.DB.Query("SELECT * FROM recipes;")
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
	row := env.DB.QueryRow("SELECT * FROM recipes WHERE id=$1;", id)
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
	ingredients := c.QueryArray("ingredients")

	if ingredients == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You MUST give at least one ingredient"})
		return
	}

	q := "SELECT * FROM recipes WHERE ingredients *~ '$1';"
	regex := strings.Join(ingredients, "|")

	row := env.DB.QueryRow(q, regex)
	err, recipe := db.ScanRecipe(row)

    switch err {

    case nil:
        c.JSON(http.StatusOK, gin.H{"recipe": recipe})

    case sql.ErrNoRows:
		e := fmt.Sprintf("Not found receipe for ingredients")
		c.JSON(http.StatusNotFound, gin.H{"message": e})

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
