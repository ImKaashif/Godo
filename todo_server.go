package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Todo struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type PostTodo struct {
	Title string `json:"title"`
}

type Response struct {
	Message string `json:"message"`
}

func get_todos(todos *[]Todo) error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	db, err := sql.Open("postgres", fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable", host, port, user, password, name))
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer db.Close()
	todo_rows, err := db.Query("select * from todos")
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer todo_rows.Close()
	for todo_rows.Next() {
		todo := Todo{}
		if err = todo_rows.Scan(&todo.Id, &todo.Title); err != nil {
			fmt.Println(err)
			continue
		}
		*todos = append(*todos, todo)
	}
	return nil
}

func insert_todo(todo_title string) error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	db, err := sql.Open("postgres", fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable", host, port, user, password, name))
	defer db.Close()
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO todos (title) values ($1)", todo_title)
	if err != nil {
		return err
	}
	return nil
}

func serve_todo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		todos := make([]Todo, 0)
		err := get_todos(&todos)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Something went wrong"})
		}
		json.NewEncoder(w).Encode(todos)
	case http.MethodPost:
		post_todo := PostTodo{}
		err := json.NewDecoder(r.Body).Decode(&post_todo)

		if err != nil {
			fmt.Println("error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "Something went wrong"})
			return
		}
		err = insert_todo(post_todo.Title)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Something went wrong"})
			return
		}
		w.WriteHeader(http.StatusCreated)

	default:
		fmt.Fprintf(w, "Method not allowed")
	}

}

func main() {
	err := godotenv.Load("todo.env")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/todo", serve_todo)
	port := "8000"
	host := "localhost"
	fmt.Println("listening on port", port)
	err = http.ListenAndServe(host+":"+port, nil)
	if err != nil {
		fmt.Println(err)
	}
}
