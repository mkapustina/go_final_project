package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mkapustina/go_final_project/internal/models"
)

const selectSQL = "SELECT id, date, title, comment, repeat FROM scheduler"

type TaskStore struct {
	Db *sql.DB
}

func (s *TaskStore) Add(t models.Task) (int, error) {
	res, err := s.Db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES(:date, :title, :comment, :repeat);",
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

func (s *TaskStore) Update(t models.Task) (int, error) {
	res, err := s.Db.Exec("UPDATE scheduler SET date=:date, title=:title, comment=:comment, repeat=:repeat WHERE id=:id;",
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

func (s *TaskStore) Get(id int) (models.Task, error) {
	row := s.Db.QueryRow(selectSQL+" WHERE id = :id",
		sql.Named("id", id))

	task := models.Task{}

	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if (err != nil) && errors.Is(err, sql.ErrNoRows) {
		return task, fmt.Errorf("задача %d отсутствует", id)
	}

	return task, err
}

func (s *TaskStore) GetAll(search string) (models.Tasks, error) {
	const rowsLimit = 10
	const limitSql = " ORDER BY date LIMIT :limit"
	var err error
	var rows *sql.Rows
	var sqlText string

	res := models.Tasks{}
	res.Tasks = make([]models.Task, 0)
	sqlText = selectSQL

	if len(search) > 0 {
		searchDT, errDt := time.Parse("02.01.2006", search)
		if errDt == nil {
			search = searchDT.Format("20060102")

			sqlText += " WHERE date LIKE :text "
		} else {
			search = "%" + strings.ToLower(search) + "%"
			sqlText += " WHERE LOWER(title) LIKE :text OR LOWER(comment) LIKE :text"

		}
		sqlText += limitSql
		rows, err = s.Db.Query(sqlText,
			sql.Named("text", search),
			sql.Named("limit", rowsLimit))
	} else {
		sqlText += limitSql
		rows, err = s.Db.Query(sqlText,
			sql.Named("limit", rowsLimit))
	}
	if err != nil {
		return res, err
	}

	defer rows.Close()

	if rows == nil {
		return res, nil
	}

	for rows.Next() {
		task := models.Task{}

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

func (s *TaskStore) Delete(id int) error {
	_, err := s.Db.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	if err != nil {
		return err
	}

	return nil
}
