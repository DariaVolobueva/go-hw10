package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TaskStore struct {
	sync.RWMutex
	tasks  map[int]Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

func (ts *TaskStore) Add(task Task) int {
	ts.Lock()
	defer ts.Unlock()
	task.ID = ts.nextID
	ts.tasks[task.ID] = task
	ts.nextID++
	return task.ID
}

func (ts *TaskStore) Get(id int) (Task, bool) {
	ts.RLock()
	defer ts.RUnlock()
	task, ok := ts.tasks[id]
	return task, ok
}

func (ts *TaskStore) Update(id int, task Task) bool {
	ts.Lock()
	defer ts.Unlock()
	if _, ok := ts.tasks[id]; !ok {
		return false
	}
	task.ID = id
	ts.tasks[id] = task
	return true
}

func (ts *TaskStore) Delete(id int) bool {
	ts.Lock()
	defer ts.Unlock()
	if _, ok := ts.tasks[id]; !ok {
		return false
	}
	delete(ts.tasks, id)
	return true
}

func (ts *TaskStore) GetAll() []Task {
	ts.RLock()
	defer ts.RUnlock()
	tasks := make([]Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

var store = NewTaskStore()

func main() {
	http.HandleFunc("/tasks", handleTasks)
	http.HandleFunc("/tasks/", handleTask)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks := store.GetAll()
		json.NewEncoder(w).Encode(tasks)
	case http.MethodPost:
		var task Task
		json.NewDecoder(r.Body).Decode(&task)
		id := store.Add(task)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "{\"id\": %d}", id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	id := 0
	fmt.Sscanf(r.URL.Path, "/tasks/%d", &id)
	if id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, ok := store.Get(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(task)
	case http.MethodPut:
		var task Task
		json.NewDecoder(r.Body).Decode(&task)
		if store.Update(id, task) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case http.MethodDelete:
		if store.Delete(id) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}