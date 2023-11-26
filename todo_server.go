package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Todo struct {
	Id         int          `json:"id"`
	Title      string       `json:"title"`
	Status     int          `json:"status"`
	Created_on time.Time    `json:"created_on"`
	Due_date   sql.NullTime `json:"due_date"`
}

type PostTodo struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
}

type Response struct {
	Message string `json:"message"`
}

func get_db_instance() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	db, err := sql.Open("postgres", fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable", host, port, user, password, name))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func get_todos(todos *[]Todo) error {
	db, err := get_db_instance()
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
		if err = todo_rows.Scan(&todo.Id, &todo.Title, &todo.Status, &todo.Created_on, &todo.Due_date); err != nil {
			fmt.Println(err)
			continue
		}
		*todos = append(*todos, todo)
	}
	return nil
}

func insert_todo(post_todo *PostTodo) error {
	db, err := get_db_instance()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("INSERT INTO todos (title, status) values ($1, $2)", post_todo.Title, post_todo.Status)
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
		err = insert_todo(&post_todo)
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

func todo_action(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	db, err := get_db_instance()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if id == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		row := db.QueryRow("SELECT * FROM todos WHERE id=$1", id)
		todo := Todo{}
		err = row.Scan(&todo.Id, &todo.Title, &todo.Status, &todo.Created_on, &todo.Due_date)
		if err == sql.ErrNoRows {
			fmt.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(todo)
	case http.MethodDelete:
		row, err := db.Exec("DELETE FROM todos WHERE id = $1", id)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rows_affected, err := row.RowsAffected()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if rows_affected == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func main() {
	err := godotenv.Load("todo.env")
	if err != nil {
		panic(err)
	}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/todos", serve_todo).Methods("GET", "POST")
	router.HandleFunc("/todos/{id}", todo_action).Methods("GET", "DELETE")
	port := "8000"
	host := "localhost"
	fmt.Println("listening on port", port)
	err = http.ListenAndServe(host+":"+port, router)
	if err != nil {
		fmt.Println(err)
	}
}
