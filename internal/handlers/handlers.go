package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/boofexxx/todolist/internal/data"
)

type ServerMux struct {
	*http.ServeMux

	Logger *log.Logger
	store  *data.TaskStore
}

func NewServerMux(mux *http.ServeMux, logger *log.Logger, dbData string) (*ServerMux, error) {
	db, err := data.NewTaskStore(dbData)
	if err != nil {
		return nil, err
	}
	return &ServerMux{
		ServeMux: mux,
		Logger:   logger,
		store:    db,
	}, nil
}

func (s *ServerMux) TaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tasks/" {
		switch r.Method {
		case http.MethodGet:
			s.getTaskHandler(w, r)
		case http.MethodPost:
			s.createTaskHandler(w, r)
		case http.MethodDelete:
			s.deleteTasksHandler(w, r)
		default:
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
		}
	} else {
		idString := strings.TrimPrefix(r.URL.Path, "/tasks/")
		id, err := strconv.Atoi(idString)
		if err != nil {
			http.Error(w, "expected id to be integer number", http.StatusExpectationFailed)
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.getTaskByIDHandler(w, r, id)
		case http.MethodPut:
			s.changeTaskByIDHandler(w, r, id)
		case http.MethodDelete:
			s.deleteTaskByIDHandler(w, r, id)
		default:
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
		}
	}
}

func (s *ServerMux) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(strings.Join(r.Header["Accept"], " "), "application/json") {
		http.Error(w, "expected application/json Accept", http.StatusExpectationFailed)
		return
	}

	type ResponseTasks struct {
		Tasks []*data.Task `json:"tasks"`
	}
	tasks, err := s.store.GetAllTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json, err := json.Marshal(&ResponseTasks{tasks})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (s *ServerMux) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(strings.Join(r.Header["Content-Type"], " "), "application/json") {
		http.Error(w, "expected application/json Content-Type", http.StatusExpectationFailed)
		return
	}
	if !strings.Contains(strings.Join(r.Header["Accept"], " "), "application/json") {
		http.Error(w, "expected application/json Accept", http.StatusExpectationFailed)
		return
	}

	type RequestTask struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Done        bool   `json:"done"`
		Author      string `json:"author"`
	}
	task := RequestTask{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := s.store.CreateTask(task.Name, task.Description, task.Done, task.Author)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type ResponseID struct {
		ID int `json:"id"`
	}
	json, err := json.Marshal(ResponseID{id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (s *ServerMux) deleteTasksHandler(w http.ResponseWriter, r *http.Request) {
	err := s.store.DeleteAllTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *ServerMux) getTaskByIDHandler(w http.ResponseWriter, r *http.Request, id int) {
	if !strings.Contains(strings.Join(r.Header["Accept"], " "), "application/json") {
		http.Error(w, "expected application/json Accept", http.StatusBadRequest)
		return
	}
	task, err := s.store.GetTaskByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (s *ServerMux) changeTaskByIDHandler(w http.ResponseWriter, r *http.Request, id int) {
	if !strings.Contains(strings.Join(r.Header["Content-Type"], " "), "application/json") {
		http.Error(w, "expected application/json Content-Type", http.StatusExpectationFailed)
		return
	}

	type RequestTask struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Done        bool   `json:"done"`
		Author      string `json:"author"`
	}
	task := RequestTask{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.store.ChangeTask(id, task.Name, task.Description, task.Done, task.Author); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *ServerMux) deleteTaskByIDHandler(w http.ResponseWriter, r *http.Request, id int) {
	if err := s.store.DeleteTask(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
