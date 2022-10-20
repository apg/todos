package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var NotFound = errors.New("not found")

// UghRM is a really bad ORM.
type UghRM struct {
	db *sql.DB
}

func NewUghRM(db *sql.DB) *UghRM {
	return &UghRM{
		db: db,
	}
}

func (m *UghRM) Insert(table string, values map[string]any) *Result[int] {
	columns := []string{}
	args := []any{}
	qs := []string{}

	i := 1
	for k, v := range values {
		if k != "id" {
			columns = append(columns, k)
			args = append(args, v)
			qs = append(qs, fmt.Sprintf("$%d", i))
			i++
		}
	}

	var id int
	err := m.db.QueryRow(
		fmt.Sprintf("insert into %s (%s) values (%s) returning id",
			table, strings.Join(columns, ","), strings.Join(qs, ", ")),
		args...).Scan(&id)

	if err != nil {
		return Err[int](err)
	}
	return OK(id)
}

func (m *UghRM) Update(table string, pk int, values map[string]any) *Result[bool] {
	columns := []string{}
	args := []any{}

	i := 1
	for k, v := range values {
		columns = append(columns, fmt.Sprintf("%s = $%d", k, i))
		i++
		args = append(args, v)
	}

	args = append(args, pk)

	res, err := m.db.Exec(
		fmt.Sprintf("update %s set %s where id = $%d",
			table, strings.Join(columns, ", "), i),
		args...)
	if err != nil {
		return Err[bool](err)
	}
	aff, err := res.RowsAffected()
	return OK(aff == 1)
}

func (m *UghRM) Delete(table string, pk int) *Result[bool] {
	res, err := m.db.Exec(fmt.Sprintf("delete from %s where id = $1", table), pk)
	if err != nil {
		return Err[bool](err)
	}
	aff, err := res.RowsAffected()
	return OK(aff == 1)
}

func (m *UghRM) All(table string, direction int) *Result[[]map[string]any] {
	rows, err := m.db.Query(fmt.Sprintf("select * from %s order by id %s", table, whichDirection(direction)))
	if err != nil {
		return Err[[]map[string]any](err)
	}
	defer rows.Close()

	return slurp(rows)
}

func (m *UghRM) Get(table string, pk int) *Result[map[string]any] {
	rows, err := m.db.Query(
		fmt.Sprintf("select * from %s where id = $1", table),
		pk)
	if err != nil {
		return Err[map[string]any](err)
	}
	defer rows.Close()

	rs := slurp(rows)
	if ok := rs.IsErr(); !ok {
		out := rs.OK()
		return OK(out[0])
	}
	return Err[map[string]any](NotFound)
}

func (m *UghRM) Where(table string, clause string, params []any, direction int) *Result[[]map[string]any] {
	rows, err := m.db.Query(
		fmt.Sprintf("select * from %s where %s order by id %s",
			table, clause, whichDirection(direction)),
		params...)
	if err != nil {
		return Err[[]map[string]any](err)
	}
	defer rows.Close()
	return slurp(rows)
}

func slurp(rows *sql.Rows) *Result[[]map[string]any] {
	var out []map[string]any
	columns, err := rows.Columns()
	if err != nil {
		return Err[[]map[string]any](err)
	}

	for rows.Next() {
		columnData := make([]any, len(columns))
		columnPointers := make([]any, len(columns))
		for i, _ := range columns {
			columnPointers[i] = &columnData[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return Err[[]map[string]any](err)
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]any)
		for i, colName := range columns {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		out = append(out, m)
	}
	if err = rows.Err(); err != nil {
		return Err[[]map[string]any](err)
	}

	return OK(out)
}

func whichDirection(x int) string {
	if x < 0 {
		return "DESC"
	}
	return "ASC"
}
