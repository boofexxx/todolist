package data

import (
	"database/sql"
	"sync"

	_ "github.com/lib/pq"
)

type Task struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Done        bool   `json:"bool"`
	Author      string `json:"author"`
}

type TaskStore struct {
	sync.RWMutex
	db *sql.DB
}

func NewTaskStore(dbData string) (*TaskStore, error) {
	db, err := sql.Open("postgres", dbData)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &TaskStore{
		RWMutex: sync.RWMutex{},
		db:      db,
	}, nil
}

func (ts *TaskStore) GetTaskByID(id int) (*Task, error) {
	ts.RLock()
	defer ts.RUnlock()

	task := Task{}
	err := ts.db.QueryRow("SELECT * FROM todolist WHERE id = $1", id).
		Scan(&task.ID, &task.Name, &task.Description, &task.Done, &task.Author)
	if err != nil {
		return nil, err
	}
	return &task, err
}

func (ts *TaskStore) ChangeTask(id int, name string, description string, done bool, author string) error {
	_, err := ts.db.Exec(
		"UPDATE todolist SET name=$1, description=$2, done=$3, author=$4 WHERE id=$5",
		name, description, done, author, id)
	return err
}

func (ts *TaskStore) DeleteTask(id int) error {
	_, err := ts.db.Exec("DELETE FROM todolist WHERE id = $1", id)
	return err
}

func (ts *TaskStore) CreateTask(name string, description string, done bool, author string) (int, error) {
	ts.Lock()
	defer ts.Unlock()

	task := &Task{
		ID:          0,
		Name:        name,
		Description: description,
		Done:        done,
		Author:      author,
	}

	err := ts.db.QueryRow(
		"INSERT INTO todolist(name, description, done) VALUES($1, $2, $3) RETURNING id",
		task.Name, task.Description, task.Done).Scan(&task.ID)
	if err != nil {
		return 0, err
	}

	return task.ID, nil
}

func (ts *TaskStore) GetAllTasks() ([]*Task, error) {
	ts.RLock()
	defer ts.RUnlock()

	// XXX: probably we can count number of tasks
	// so we can just allocate with enough capacity
	tasks := make([]*Task, 0)

	rows, err := ts.db.Query("SELECT * FROM todolist")
	if err != nil {
		return nil, err
	}
	// we can omit it. it would be closed eventually
	// if we don't break for loop ourself
	defer rows.Close()

	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.Done, &task.Author)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (ts *TaskStore) DeleteAllTasks() error {
	ts.Lock()
	defer ts.Unlock()

	_, err := ts.db.Exec("DELETE FROM todolist")
	return err
}
