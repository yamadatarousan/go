package repository

import (
	"encoding/json"
	"os"
)

type FileRepo struct {
	filename string
}

func NewFileRepo(filename string) *FileRepo {
	return &FileRepo{filename: filename}
}

func (f *FileRepo) Save(todo Todo) error {
	todos, _ := f.FindAll()
	todos = append(todos, todo)
	data, _ := json.Marshal(todos)
	return os.WriteFile(f.filename, data, 0644)
}

func (f *FileRepo) FindAll() ([]Todo, error) {
	data, err := os.ReadFile(f.filename)
	if err != nil {
		return []Todo{}, nil
	}
	var todos []Todo
	json.Unmarshal(data, &todos)
	return todos, nil
}
