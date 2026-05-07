package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	UserID    int    `json:"-"`
}

type User struct {
	ID       int    `json:"id,omitempty"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASS", "P@$$w0rd"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "todolist"),
	)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir banco: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}
	fmt.Println("Conectado ao MySQL com sucesso!")

	mux := http.NewServeMux()

	// Rotas Públicas (Auth)
	mux.HandleFunc("POST /api/v1/register", registerUser)
	mux.HandleFunc("POST /api/v1/login", loginUser)

	// Rotas Protegidas (Tarefas)
	mux.HandleFunc("GET /api/v1/tasks", withAuth(getTasks))
	mux.HandleFunc("POST /api/v1/tasks", withAuth(createTask))
	mux.HandleFunc("PUT /api/v1/tasks/{id}", withAuth(updateTask))
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", withAuth(deleteTask))

	server := &http.Server{
		Addr:    ":8080",
		Handler: securityMiddleware(mux),
	}

	go func() {
		fmt.Println("Servidor rodando na porta 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro no servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\nDesligando servidor graciosamente...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Erro no shutdown: %v", err)
	}
}

// Utilitários de Segurança
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Handlers de Autenticação
func registerUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Input inválido", http.StatusBadRequest)
		return
	}

	u.Username = strings.TrimSpace(u.Username)
	u.Password = strings.TrimSpace(u.Password)
	if u.Username == "" || u.Password == "" {
		http.Error(w, "Usuário e senha são obrigatórios", http.StatusBadRequest)
		return
	}

	hashedPassword := hashPassword(u.Password)
	stmt, err := db.Prepare("INSERT INTO users (username, password) VALUES (?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Username, hashedPassword)
	if err != nil {
		http.Error(w, "Erro ao criar usuário (possível duplicação)", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Input inválido", http.StatusBadRequest)
		return
	}

	hashedPassword := hashPassword(u.Password)
	var id int
	err := db.QueryRow("SELECT id FROM users WHERE username = ? AND password = ?", u.Username, hashedPassword).Scan(&id)
	if err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	token := generateToken()
	_, err = db.Exec("UPDATE users SET token = ? WHERE id = ?", token, id)
	if err != nil {
		http.Error(w, "Erro ao gerar sessão", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// Middleware de Autenticação
func withAuth(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token == "" || token == authHeader {
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}

		var userID int
		err := db.QueryRow("SELECT id FROM users WHERE token = ?", token).Scan(&userID)
		if err != nil {
			http.Error(w, "Sessão inválida ou expirada", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		handler(w, r.WithContext(ctx))
	}
}

// Handlers de Tarefas (Protegidos)
func getTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	rows, err := db.Query("SELECT id, title, completed FROM tasks WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Input inválido", http.StatusBadRequest)
		return
	}

	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		http.Error(w, "O título não pode estar vazio", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO tasks (title, user_id) VALUES (?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(t.Title, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	t.ID = int(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("UPDATE tasks SET completed = true WHERE id = ? AND user_id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Tarefa não encontrada ou não pertence a você", http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("DELETE FROM tasks WHERE id = ? AND user_id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Tarefa não encontrada ou não pertence a você", http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
