package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var (
	tasks  = make(map[int]Task)
	nextID = 1
	mu     sync.RWMutex // 複数リクエストでの競合を防ぐための Mutex
)

// sendJSONError は指定された HTTP ステータスとメッセージで JSON エラーを返却するヘルパー関数です
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

// POST /tasks: タスクの作成
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		// 【Validation Error】JSONのパースに失敗した場合 (400 Bad Request)
		sendJSONError(w, "無効なJSONフォーマットです", http.StatusBadRequest)
		return
	}

	// 【Validation Error】必須項目(Title)が空の場合 (400 Bad Request)
	if task.Title == "" {
		sendJSONError(w, "タイトルは必須です", http.StatusBadRequest)
		return
	}

	// 【Internal Server Error】デモ用に特定の文字列で意図的にサーバーエラーを発生させる (500 Internal Server Error)
	if task.Title == "force_error" {
		sendJSONError(w, "データベースの接続に失敗しました(模擬エラー)", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	task.ID = nextID
	nextID++
	tasks[task.ID] = task
	mu.Unlock()

	// 成功時のレスポンス (201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 【Validation Error】IDが数値でない場合 (400 Bad Request)
		sendJSONError(w, "無効なタスクIDの形式です", http.StatusBadRequest)
		return
	}

	mu.RLock()
	task, exists := tasks[id]
	mu.RUnlock()

	// 【Not Found Error】タスクが存在しない場合 (404 Not Found)
	if !exists {
		sendJSONError(w, fmt.Sprintf("ID %d のタスクは見つかりませんでした", id), http.StatusNotFound)
		return
	}

	// 成功時のレスポンス (200 OK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /tasks", createTaskHandler)
	mux.HandleFunc("GET /tasks/{id}", getTaskHandler)

	fmt.Println("サーバーを起動しました: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
