package main

import (
	"fmt"
	"log"
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


type filter func([]string, int) (string, []any)


var parametersToFilterFunction map[string] filter = map[string] filter{
  "ingredient": filterByIngredient,
  "tag": filterByTag,
  "id": filterById,
  "name": filterByName,
  "author": filterByAuthor,
  "budget": filterByBudget,
  "setup_time": filterBySetupTime,
  "cook_time": filterByCookTime,
  "total_time": filterByTotalTime,
  "difficulty": filterByDifficulty,
  "people_quantity": filterByPeopleQuantity,
}


func hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hey! Let's cook"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
    }
}


func (env Env) getRecipes(c *gin.Context) {
	queries := make([]string, 0)
	values := make([]any, 0)

	tx := db.BeginWithFunctions(env.DB)
	defer tx.Commit()

	for parameter, filter := range(parametersToFilterFunction) {
		inputs := c.QueryArray(parameter)

		if inputs != nil {
			offset := len(values)
			q, v := filter(inputs, offset)
			queries = append(queries, q)
			values = append(values, v...)
		}
	}


	query := db.MakeRecipeQuery(strings.Join(queries, " AND "))
	log.Printf(query)
	rows, err := tx.Query(query, values...)
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
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
    }
}


func filterBy(field string, comparator string, valueWrapper string, valueCombination string, filters []string, offset int) (string, []any) {
	queries := make([]string, 0)
	values := make([]any, 0)

	for idx, filter := range(filters) {
		queries = append(queries, fmt.Sprintf("%s %s $%d", field, comparator, idx + offset + 1))
		values = append(values, valueWrapper + filter + valueWrapper)
	}

	query := "(" + strings.Join(queries, valueCombination) + ")"
	return query, values
}


func filterByIngredient(ingredients []string, offset int) (string, []any) {
	query, values := filterBy("ingredient", "ILIKE", "%", " OR ", ingredients, offset)
	query = fmt.Sprintf("recipes.id IN (SELECT recipe_id FROM ingredients WHERE %s)", query)

	return query, values
}

func filterByTag(tags []string, offset int) (string, []any) {
	query, values := filterBy("tag", "ILIKE", "%",  " OR ",tags, offset)
	query = fmt.Sprintf("recipes.id IN (SELECT recipe_id FROM tags WHERE %s)", query)

	return query, values
}

func filterById(ids []string, offset int) (string, []any) {
	return filterBy("recipes.id", "=", "",  " OR ", ids, offset)
}

func filterByName(names []string, offset int) (string, []any) {
	return filterBy("name", "ILIKE", "%",  " OR ", names, offset)
}

func filterByAuthor(authors []string, offset int) (string, []any) {
	return filterBy("author", "ILIKE", "%",  " OR ", authors, offset)
}

func filterByBudget(budgets []string, offset int) (string, []any) {
	return filterBy("budget", "ILIKE", "%",  " OR ", budgets, offset)
}

func filterBySetupTime(setupTimes []string, offset int) (string, []any) {
	return filterBy("setup_time", "ILIKE", "%",  " OR ", setupTimes, offset)
}

func filterByCookTime(cookTimes []string, offset int) (string, []any) {
	return filterBy("cook_time", "ILIKE", "%",  " OR ", cookTimes, offset)
}

func filterByTotalTime(totalTimes []string, offset int) (string, []any) {
	return filterBy("total_time", "ILIKE", "%",  " OR ", totalTimes, offset)
}

func filterByDifficulty(difficulties []string, offset int) (string, []any) {
	return filterBy("difficulty", "ILIKE", "%",  " OR ", difficulties, offset)
}

func filterByPeopleQuantity(peopleQuantities []string, offset int) (string, []any) {
	return filterBy("people_quantity", "ILIKE", "%",  " OR ", peopleQuantities, offset)
}


func main() {
	env := new(Env)
	env.DB = db.Initialize()

    router := gin.Default()
    router.GET("/", hello)
    router.GET("/recipes", env.getRecipes)
    router.GET("/recipes/:id", env.getRecipeByID)

    router.Run()
}
