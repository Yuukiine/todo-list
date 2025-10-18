package storage

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"toDoList/models"
	e "toDoList/pkg/errors"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("postgres", path)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Create(w http.ResponseWriter, r *http.Request) {
	log.Println("storage.Create()...")
	var list models.List
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.Create(): ", err)

		return
	}

	stmt, err := s.db.Prepare(`
		INSERT INTO tasks(id, title, description) 
		VALUES($1, $2, $3)
	`)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.Create(): ", err)

		return
	}

	_, err = stmt.Exec(list.ID, list.Title, list.Description)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.Create(): ", err)

		return
	}
}

func (s *Storage) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	log.Println("storage.GetAllTasks()...")
	stmt, err := s.db.Prepare(`
		SELECT id, title, description, completed
		FROM tasks
	`)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.GetAllTasks(): ", err)

		return
	}

	var lists []models.List
	rows, err := stmt.Query()
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.GetAllTasks(): ", err)

		return
	}
	defer rows.Close()

	for rows.Next() {
		var list models.List

		if err = rows.Scan(&list.ID, &list.Title, &list.Description, &list.Completed); err != nil {
			e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
			log.Println("storage.GetAllTasks(): ", err)

			return
		}

		lists = append(lists, list)
	}

	if err = rows.Err(); err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.GetAllTasks(): ", err)

		return
	}

	w.WriteHeader(http.StatusOK)
	b, err := json.MarshalIndent(lists, "", "	")
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.GetAllTasks(): ", err)

		return
	}
	if _, err = w.Write(b); err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.GetAllTasks(): ", err)

		return
	}
}

func (s *Storage) AllIDs(w http.ResponseWriter, r *http.Request) {
	log.Println("storage.AllIDs()...")
	var a struct {
		IDs []string `json:"ids"`
	}

	stmt, err := s.db.Prepare(`
		SELECT id
		FROM tasks
	`)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.AllIDs(): ", err)

		return
	}

	var ids []string
	rows, err := stmt.Query()
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.AllIDs(): ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string

		if err = rows.Scan(&id); err != nil {
			e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
			log.Println("storage.AllIDs(): ", err)
			return
		}

		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.AllIDs(): ", err)
		return
	}

	a.IDs = ids
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(a); err != nil {
		panic(err)
	}
}

func (s *Storage) DeleteTask(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteTask()...")
	var id struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.DeleteTask(): ", err)
		return
	}

	stmt, err := s.db.Prepare(`
		DELETE FROM tasks
		WHERE id = $1
	`)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.DeleteTask(): ", err)
		return
	}

	result, err := stmt.Exec(id.ID)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.DeleteTask(): ", err)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.DeleteTask(): ", err)
		return
	}

	if rows == 0 {
		e.SendJSONError(w, e.ErrNotExists, http.StatusNotFound)
		log.Println("storage.DeleteTask(): ", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Storage) CompleteTask(w http.ResponseWriter, r *http.Request) {
	log.Println("CompleteTask()...")
	var id struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.CompleteTask(): ", err)
		return
	}

	stmt, err := s.db.Prepare(`
		UPDATE tasks
		SET completed = true
		WHERE id = $1
	`)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.CompleteTask(): ", err)
		return
	}

	result, err := stmt.Exec(id.ID)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.CompleteTask(): ", err)
		return
	}

	rows, err := result.RowsAffected()
	log.Println(result.RowsAffected())
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.CompleteTask(): ", err)
		return
	}

	if rows == 0 {
		e.SendJSONError(w, e.ErrNotExists, http.StatusNotFound)
		log.Println("storage.CompleteTask(): ", err)
		return
	}

	stmt, err = s.db.Prepare(`
		SELECT id, title, description, completed
		FROM tasks
		WHERE id = $1
	`)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.CompleteTask(): ", err)
		return
	}

	var list models.List
	err = stmt.QueryRow(id.ID).Scan(&list.ID, &list.Title, &list.Description, &list.Completed)
	if err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.CompleteTask(): ", err)
		return
	}

	b, err := json.MarshalIndent(list, "", "	")

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		e.SendJSONError(w, e.ErrInternalServer, http.StatusInternalServerError)
		log.Println("storage.CompleteTask(): ", err)
	}
}
