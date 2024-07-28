package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Unit тести

func TestTaskStore_Add(t *testing.T) {
	store := NewTaskStore()
	task := Task{Title: "Test Task", Completed: false}
	
	id := store.Add(task)
	
	if id != 1 {
		t.Errorf("Expected ID to be 1, got %d", id)
	}
	
	addedTask, ok := store.Get(id)
	if !ok {
		t.Errorf("Failed to get added task")
	}
	
	if addedTask.Title != task.Title {
		t.Errorf("Expected title %s, got %s", task.Title, addedTask.Title)
	}
}

func TestTaskStore_Update(t *testing.T) {
	store := NewTaskStore()
	task := Task{Title: "Test Task", Completed: false}
	id := store.Add(task)
	
	updatedTask := Task{Title: "Updated Task", Completed: true}
	success := store.Update(id, updatedTask)
	
	if !success {
		t.Errorf("Update failed")
	}
	
	retrievedTask, ok := store.Get(id)
	if !ok {
		t.Errorf("Failed to get updated task")
	}
	
	if retrievedTask.Title != updatedTask.Title {
		t.Errorf("Expected title %s, got %s", updatedTask.Title, retrievedTask.Title)
	}
	
	if retrievedTask.Completed != updatedTask.Completed {
		t.Errorf("Expected completed %v, got %v", updatedTask.Completed, retrievedTask.Completed)
	}
}

func TestTaskStore_Delete(t *testing.T) {
	store := NewTaskStore()
	task := Task{Title: "Test Task", Completed: false}
	id := store.Add(task)
	
	success := store.Delete(id)
	if !success {
		t.Errorf("Delete failed")
	}
	
	_, ok := store.Get(id)
	if ok {
		t.Errorf("Task should have been deleted")
	}
	
	// Негативний випадок: видалення неіснуючого завдання
	success = store.Delete(9999)
	if success {
		t.Errorf("Delete should have failed for non-existent task")
	}
}

// Функціональний тест

func TestTaskAPI(t *testing.T) {
	store := NewTaskStore()
	resource := &TaskResource{s: store}
	
	// Створення нового завдання
	createTask := Task{Title: "New Task", Completed: false}
	createBody, _ := json.Marshal(createTask)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(createBody))
	rr := httptest.NewRecorder()
	
	resource.CreateOne(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var createdTask Task
	json.Unmarshal(rr.Body.Bytes(), &createdTask)
	
	// Отримання створеного завдання
	req, _ = http.NewRequest("GET", fmt.Sprintf("/tasks/%d", createdTask.ID), nil)
	rr = httptest.NewRecorder()
	
	resource.GetOne(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var retrievedTask Task
	json.Unmarshal(rr.Body.Bytes(), &retrievedTask)
	
	if retrievedTask.Title != createTask.Title {
		t.Errorf("Retrieved task does not match created task")
	}
	
	// Оновлення завдання
	updateTask := Task{Title: "Updated Task", Completed: true}
	updateBody, _ := json.Marshal(updateTask)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/tasks/%d", createdTask.ID), bytes.NewBuffer(updateBody))
	rr = httptest.NewRecorder()
	
	resource.UpdateOne(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Перевірка оновленого завдання
	req, _ = http.NewRequest("GET", fmt.Sprintf("/tasks/%d", createdTask.ID), nil)
	rr = httptest.NewRecorder()
	
	resource.GetOne(rr, req)
	
	json.Unmarshal(rr.Body.Bytes(), &retrievedTask)
	
	if retrievedTask.Title != updateTask.Title || retrievedTask.Completed != updateTask.Completed {
		t.Errorf("Updated task does not match expected values")
	}

	// Видалення завдання
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/tasks/%d", createdTask.ID), nil)
	rr = httptest.NewRecorder()

	resource.DeleteOne(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Перевірка, що завдання було видалено
	req, _ = http.NewRequest("GET", fmt.Sprintf("/tasks/%d", createdTask.ID), nil)
	rr = httptest.NewRecorder()

	resource.GetOne(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}