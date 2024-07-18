package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

func main() {
    mux := http.NewServeMux()
    s := NewTaskStore()
    tasks := TaskResource{
        s: s,
    }

    mux.HandleFunc("GET /tasks", tasks.GetAll)
    mux.HandleFunc("POST /tasks", tasks.CreateOne)
    mux.HandleFunc("GET /tasks/{id}", tasks.GetOne)
    mux.HandleFunc("PUT /tasks/{id}", tasks.UpdateOne)
    mux.HandleFunc("DELETE /tasks/{id}", tasks.DeleteOne)

    if err := http.ListenAndServe(":8080", mux); err != nil {
        fmt.Printf("Failed to listen and serve: %v\n", err)
    }
}

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

type TaskResource struct {
    s *TaskStore
}

func (t *TaskResource) GetAll(w http.ResponseWriter, r *http.Request) {
    tasks := t.s.GetAll()
    err := json.NewEncoder(w).Encode(tasks)
    if err != nil {
        fmt.Printf("Failed to encode: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (t *TaskResource) CreateOne(w http.ResponseWriter, r *http.Request) {
    var task Task
    err := json.NewDecoder(r.Body).Decode(&task)
    if err != nil {
        fmt.Printf("Failed to decode: %v\n", err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    task.ID = t.s.Add(task)
    err = json.NewEncoder(w).Encode(task)
    if err != nil {
        fmt.Printf("Failed to encode: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (t *TaskResource) GetOne(w http.ResponseWriter, r *http.Request) {
    idVal := r.PathValue("id")
    id, err := strconv.Atoi(idVal)
    if err != nil {
        fmt.Printf("Invalid id param: %v\n", err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    task, ok := t.s.Get(id)
    if !ok {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    err = json.NewEncoder(w).Encode(task)
    if err != nil {
        fmt.Printf("Failed to encode: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (t *TaskResource) UpdateOne(w http.ResponseWriter, r *http.Request) {
    idVal := r.PathValue("id")
    id, err := strconv.Atoi(idVal)
    if err != nil {
        fmt.Printf("Invalid id param: %v\n", err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    var task Task
    err = json.NewDecoder(r.Body).Decode(&task)
    if err != nil {
        fmt.Printf("Failed to decode: %v\n", err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    if !t.s.Update(id, task) {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusOK)
}

func (t *TaskResource) DeleteOne(w http.ResponseWriter, r *http.Request) {
    idVal := r.PathValue("id")
    id, err := strconv.Atoi(idVal)
    if err != nil {
        fmt.Printf("Invalid id param: %v\n", err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    if !t.s.Delete(id) {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusOK)
}