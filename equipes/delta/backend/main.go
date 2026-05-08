package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var db *sql.DB

func main() {
	var err error
	// ATENÇÃO: Altere "root:senha" para as suas credenciais do MySQL
	dsn := "root:andrei1248@tcp(127.0.0.1:3306)/todo_db"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Erro ao pingar o banco:", err)
	}

	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/", taskByIDHandler)

	fmt.Println("Servidor Go rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Middleware simples para CORS
func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, title, completed FROM tasks")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var t Task
			if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
				continue
			}
			tasks = append(tasks, t)
		}
		json.NewEncoder(w).Encode(tasks)

	case "POST":
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec("INSERT INTO tasks (title, completed) VALUES (?, ?)", t.Title, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		t.ID = int(id)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

func taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = db.Exec("UPDATE tasks SET completed = ? WHERE id = ?", t.Completed, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		_, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}
