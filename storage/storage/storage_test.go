package storage

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"toDoList/models"
)

func setupMockDB(t *testing.T) (*Storage, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	storage := &Storage{db: db}
	cleanup := func() {
		db.Close()
	}

	return storage, mock, cleanup
}

func TestNew(t *testing.T) {
	t.Run("should return error for invalid connection string", func(t *testing.T) {
		storage, err := New("invalid connection string")
		assert.Error(t, err)
		assert.Nil(t, storage)
	})
}

func TestCreate(t *testing.T) {
	t.Run("should successfully create a task", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		list := models.List{
			ID:          "test-id",
			Title:       "Test Task",
			Description: "Test Description",
		}

		body, _ := json.Marshal(list)
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("INSERT INTO tasks").
			ExpectExec().
			WithArgs(list.ID, list.Title, list.Description).
			WillReturnResult(sqlmock.NewResult(1, 1))

		storage.Create(w, req)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle invalid JSON", func(t *testing.T) {
		storage, _, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		storage.Create(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should handle prepare statement error", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		list := models.List{ID: "test-id", Title: "Test", Description: "Desc"}
		body, _ := json.Marshal(list)
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("INSERT INTO tasks").WillReturnError(sql.ErrConnDone)

		storage.Create(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle exec error", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		list := models.List{ID: "test-id", Title: "Test", Description: "Desc"}
		body, _ := json.Marshal(list)
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("INSERT INTO tasks").
			ExpectExec().
			WithArgs(list.ID, list.Title, list.Description).
			WillReturnError(sql.ErrNoRows)

		storage.Create(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetAllTasks(t *testing.T) {
	t.Run("should successfully get all tasks", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		w := httptest.NewRecorder()

		rows := sqlmock.NewRows([]string{"id", "title", "description", "completed"}).
			AddRow("id1", "Task 1", "Description 1", false).
			AddRow("id2", "Task 2", "Description 2", true)

		mock.ExpectPrepare("SELECT id, title, description, completed").
			ExpectQuery().
			WillReturnRows(rows)

		storage.GetAllTasks(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var lists []models.List
		err := json.Unmarshal(w.Body.Bytes(), &lists)
		assert.NoError(t, err)
		assert.Len(t, lists, 2)
		assert.Equal(t, "id1", lists[0].ID)
		assert.Equal(t, "Task 1", lists[0].Title)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle prepare error", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		w := httptest.NewRecorder()

		mock.ExpectPrepare("SELECT id, title, description, completed").
			WillReturnError(sql.ErrConnDone)

		storage.GetAllTasks(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle query error", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		w := httptest.NewRecorder()

		mock.ExpectPrepare("SELECT id, title, description, completed").
			ExpectQuery().
			WillReturnError(sql.ErrNoRows)

		storage.GetAllTasks(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle scan error", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		w := httptest.NewRecorder()

		rows := sqlmock.NewRows([]string{"id", "title", "description", "completed"}).
			AddRow("id1", nil, "Description 1", false)

		mock.ExpectPrepare("SELECT id, title, description, completed").
			ExpectQuery().
			WillReturnRows(rows)

		storage.GetAllTasks(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAllIDs(t *testing.T) {
	t.Run("should successfully get all IDs", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodGet, "/tasks/ids", nil)
		w := httptest.NewRecorder()

		rows := sqlmock.NewRows([]string{"id"}).
			AddRow("id1").
			AddRow("id2").
			AddRow("id3")

		mock.ExpectPrepare("SELECT id").
			ExpectQuery().
			WillReturnRows(rows)

		storage.AllIDs(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result struct {
			IDs []string `json:"ids"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result.IDs, 3)
		assert.Equal(t, "id1", result.IDs[0])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle prepare error", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodGet, "/tasks/ids", nil)
		w := httptest.NewRecorder()

		mock.ExpectPrepare("SELECT id").
			WillReturnError(sql.ErrConnDone)

		storage.AllIDs(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteTask(t *testing.T) {
	t.Run("should successfully delete a task", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		id := struct {
			ID string `json:"id"`
		}{ID: "test-id"}
		body, _ := json.Marshal(id)
		req := httptest.NewRequest(http.MethodDelete, "/tasks", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("DELETE FROM tasks").
			ExpectExec().
			WithArgs("test-id").
			WillReturnResult(sqlmock.NewResult(0, 1))

		storage.DeleteTask(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle invalid JSON", func(t *testing.T) {
		storage, _, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodDelete, "/tasks", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		storage.DeleteTask(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should handle task not found", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		id := struct {
			ID string `json:"id"`
		}{ID: "non-existent"}
		body, _ := json.Marshal(id)
		req := httptest.NewRequest(http.MethodDelete, "/tasks", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("DELETE FROM tasks").
			ExpectExec().
			WithArgs("non-existent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		storage.DeleteTask(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCompleteTask(t *testing.T) {
	t.Run("should successfully complete a task", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		id := struct {
			ID string `json:"id"`
		}{ID: "test-id"}
		body, _ := json.Marshal(id)
		req := httptest.NewRequest(http.MethodPut, "/tasks/complete", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("UPDATE tasks").
			ExpectExec().
			WithArgs("test-id").
			WillReturnResult(sqlmock.NewResult(0, 1))

		rows := sqlmock.NewRows([]string{"id", "title", "description", "completed"}).
			AddRow("test-id", "Test Task", "Test Description", true)

		mock.ExpectPrepare("SELECT id, title, description, completed").
			ExpectQuery().
			WithArgs("test-id").
			WillReturnRows(rows)

		storage.CompleteTask(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var list models.List
		err := json.Unmarshal(w.Body.Bytes(), &list)
		assert.NoError(t, err)
		assert.Equal(t, "test-id", list.ID)
		assert.True(t, list.Completed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle invalid JSON", func(t *testing.T) {
		storage, _, cleanup := setupMockDB(t)
		defer cleanup()

		req := httptest.NewRequest(http.MethodPut, "/tasks/complete", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		storage.CompleteTask(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("should handle task not found", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		id := struct {
			ID string `json:"id"`
		}{ID: "non-existent"}
		body, _ := json.Marshal(id)
		req := httptest.NewRequest(http.MethodPut, "/tasks/complete", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("UPDATE tasks").
			ExpectExec().
			WithArgs("non-existent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		storage.CompleteTask(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle prepare error on update", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		id := struct {
			ID string `json:"id"`
		}{ID: "test-id"}
		body, _ := json.Marshal(id)
		req := httptest.NewRequest(http.MethodPut, "/tasks/complete", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("UPDATE tasks").
			WillReturnError(sql.ErrConnDone)

		storage.CompleteTask(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle query error after update", func(t *testing.T) {
		storage, mock, cleanup := setupMockDB(t)
		defer cleanup()

		id := struct {
			ID string `json:"id"`
		}{ID: "test-id"}
		body, _ := json.Marshal(id)
		req := httptest.NewRequest(http.MethodPut, "/tasks/complete", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mock.ExpectPrepare("UPDATE tasks").
			ExpectExec().
			WithArgs("test-id").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectPrepare("SELECT id, title, description, completed").
			ExpectQuery().
			WithArgs("test-id").
			WillReturnError(sql.ErrNoRows)

		storage.CompleteTask(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
