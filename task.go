package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type TaskStore struct {
	db *sql.DB
}

func (s *TaskStore) Add(t Task) (int, error) {
	res, err := s.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES(:date, :title, :comment, :repeat);",
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat))
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s *TaskStore) Update(t Task) (int, error) {
	res, err := s.db.Exec("UPDATE scheduler SET date=:date, title=:title, comment=:comment, repeat=:repeat WHERE id=:id;",
		sql.Named("id", t.ID),
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat))
	if err != nil {
		return 0, err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(cnt), nil
}

func (s *TaskStore) Get(id int) (Task, error) {
	row := s.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	task := Task{}

	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if (err != nil) && errors.Is(err, sql.ErrNoRows) {
		return task, fmt.Errorf("задача %d отсутствует", id)
	}

	return task, err
}

func (s *TaskStore) GetAll(search string) (Tasks, error) {
	var err error
	var rows *sql.Rows

	res := Tasks{}
	res.Tasks = make([]Task, 0)

	if len(search) > 0 {
		rows, err = s.db.Query(
			"SELECT id, date, title, comment, repeat FROM scheduler "+
				"WHERE UPPER(title) LIKE :text OR UPPER(comment) LIKE :text OR date LIKE :text "+
				"ORDER BY date LIMIT :limit ",
			sql.Named("text", search),
			sql.Named("limit", 10))
	} else {
		rows, err = s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit ",
			sql.Named("limit", 10))
	}
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		task := Task{}

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return res, err
		}

		res.Tasks = append(res.Tasks, task)
	}
	if err := rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}

func (s TaskStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	if err != nil {
		return err
	}

	return nil
}
