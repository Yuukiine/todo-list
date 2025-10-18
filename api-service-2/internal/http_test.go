package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"toDoList/models"
)

func TestCreateHandler(t *testing.T) {
	t.Run("should successfully create a task", func(t *testing.T) {
		list := models.List{
			Title:       "Test List",
			Description: "Test Description",
		}
		body, err := json.Marshal(list)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		w := httptest.NewRecorder()

		CreateHandler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var created models.List
		err = json.NewDecoder(w.Body).Decode(&created)
		require.NoError(t, err)

		assert.Equal(t, list.Title, created.Title)
		assert.Equal(t, list.Description, created.Description)
		assert.NotEmpty(t, created.ID)
		assert.False(t, created.Completed)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader([]byte(`{invalid}`)))
		w := httptest.NewRecorder()

		CreateHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid title field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader([]byte(`{"title": "te"}`)))
		w := httptest.NewRecorder()

		CreateHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for missing title", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader([]byte(`{"description": "test"}`)))
		w := httptest.NewRecorder()

		CreateHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for empty request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader([]byte("")))
		w := httptest.NewRecorder()

		CreateHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestListHandler(t *testing.T) {
	t.Run("should successfully retrieve all tasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/list", nil)
		w := httptest.NewRecorder()

		ListHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var lists []models.List
		err := json.NewDecoder(w.Body).Decode(&lists)
		require.NoError(t, err)
		assert.NotNil(t, lists)
	})

	t.Run("should return empty array when no tasks exist", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/list", nil)
		w := httptest.NewRecorder()

		ListHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var lists []models.List
		err := json.NewDecoder(w.Body).Decode(&lists)
		require.NoError(t, err)
	})
}

func TestDeleteHandler(t *testing.T) {
	t.Run("should successfully delete an existing task", func(t *testing.T) {
		list := models.List{
			Title:       "Task to Delete",
			Description: "This will be deleted",
		}
		body, _ := json.Marshal(list)

		createReq := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		createW := httptest.NewRecorder()
		CreateHandler(createW, createReq)

		require.Equal(t, http.StatusCreated, createW.Code)

		var created models.List
		err := json.NewDecoder(createW.Body).Decode(&created)
		require.NoError(t, err)
		require.NotEmpty(t, created.ID)

		deleteBody := map[string]string{"id": created.ID}
		deleteBodyBytes, _ := json.Marshal(deleteBody)

		deleteReq := httptest.NewRequest(http.MethodDelete, "/delete", bytes.NewReader(deleteBodyBytes))
		deleteW := httptest.NewRecorder()
		DeleteHandler(deleteW, deleteReq)

		assert.Equal(t, http.StatusNoContent, deleteW.Code)
		assert.Empty(t, deleteW.Body.String())
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/delete", bytes.NewReader([]byte(`{"ip": "12345"}`)))
		w := httptest.NewRecorder()

		DeleteHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/delete", bytes.NewReader([]byte(`{invalid json}`)))
		w := httptest.NewRecorder()

		DeleteHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/delete", bytes.NewReader([]byte(`{"id": "12345"}`)))
		w := httptest.NewRecorder()

		DeleteHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 for non-existent ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/delete", bytes.NewReader([]byte(`{"id": "123456"}`)))
		w := httptest.NewRecorder()

		DeleteHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 for empty ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/delete", bytes.NewReader([]byte(`{"id": ""}`)))
		w := httptest.NewRecorder()

		DeleteHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDoneHandler(t *testing.T) {
	t.Run("should successfully mark task as completed", func(t *testing.T) {
		list := models.List{
			Title:       "Task to Complete",
			Description: "This will be marked as done",
		}
		body, _ := json.Marshal(list)

		createReq := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		createW := httptest.NewRecorder()
		CreateHandler(createW, createReq)

		require.Equal(t, http.StatusCreated, createW.Code)

		var created models.List
		err := json.NewDecoder(createW.Body).Decode(&created)
		require.NoError(t, err)
		require.NotEmpty(t, created.ID)
		require.False(t, created.Completed)

		doneBody := map[string]string{"id": created.ID}
		doneBodyBytes, _ := json.Marshal(doneBody)

		doneReq := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader(doneBodyBytes))
		doneW := httptest.NewRecorder()
		DoneHandler(doneW, doneReq)

		assert.Equal(t, http.StatusOK, doneW.Code)

		var completed models.List
		err = json.NewDecoder(doneW.Body).Decode(&completed)
		require.NoError(t, err)
		assert.True(t, completed.Completed)
		assert.Equal(t, created.ID, completed.ID)
		assert.Equal(t, created.Title, completed.Title)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader([]byte(`{"ip": "12345"}`)))
		w := httptest.NewRecorder()

		DoneHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader([]byte(`{invalid json}`)))
		w := httptest.NewRecorder()

		DoneHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader([]byte(`{"id": "12345"}`)))
		w := httptest.NewRecorder()

		DoneHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 for non-existent ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader([]byte(`{"id": "123456"}`)))
		w := httptest.NewRecorder()

		DoneHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 for empty ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader([]byte(`{"id": ""}`)))
		w := httptest.NewRecorder()

		DoneHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should handle marking already completed task", func(t *testing.T) {
		list := models.List{
			Title:       "Task Already Done",
			Description: "Will be completed twice",
		}
		body, _ := json.Marshal(list)

		createReq := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		createW := httptest.NewRecorder()
		CreateHandler(createW, createReq)

		var created models.List
		json.NewDecoder(createW.Body).Decode(&created)

		doneBody := map[string]string{"id": created.ID}
		doneBodyBytes, _ := json.Marshal(doneBody)

		doneReq1 := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader(doneBodyBytes))
		doneW1 := httptest.NewRecorder()
		DoneHandler(doneW1, doneReq1)

		doneReq2 := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader(doneBodyBytes))
		doneW2 := httptest.NewRecorder()
		DoneHandler(doneW2, doneReq2)

		assert.Equal(t, http.StatusOK, doneW2.Code)

		var completed models.List
		json.NewDecoder(doneW2.Body).Decode(&completed)
		assert.True(t, completed.Completed)
	})
}

func TestDeleteHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "empty request body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing id field",
			requestBody:    `{"title": "something"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "null id",
			requestBody:    `{"id": null}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/delete", bytes.NewReader([]byte(tt.requestBody)))
			w := httptest.NewRecorder()

			DeleteHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDoneHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "empty request body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing id field",
			requestBody:    `{"title": "something"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "null id",
			requestBody:    `{"id": null}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/done", bytes.NewReader([]byte(tt.requestBody)))
			w := httptest.NewRecorder()

			DoneHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
