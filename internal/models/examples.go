package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.api.template/internal/validator"
)

var (
	ErrExampleRecordNotFound = errors.New("example record not found")
)

type Example struct {
	Id            int64     `json:"id"`
	ExampleValue1 float64   `json:"-"`
	ExampleValue2 string    `json:"-"`
	ExampleValue3 string    `json:"example_value_3"`
	CreatedAt     time.Time `json:"created_at"`
}

type ExampleDBConnection struct {
	DB *sql.DB
}

// Insert an example in DB
func (e ExampleDBConnection) Insert(example *Example) error {
	query := `
			INSERT INTO examples (example_value_1, example_value_2, example_value_3)
			VALUES($1, $2, $3)
			RETURNING id, created_at`

	// This context statement limits the query to be executed in 3s
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return e.DB.QueryRowContext(ctx, query,
		example.ExampleValue1,
		example.ExampleValue2,
		example.ExampleValue3).Scan(
		&example.Id,
		&example.CreatedAt)
}

// Get an example from DB
func (e ExampleDBConnection) Get(id int64) (*Example, error) {
	if id < 1 {
		return nil, ErrExampleRecordNotFound
	}

	query := `
		SELECT example_value_1, example_value_2, example_value_3, created_at
		FROM examples
		WHERE id = $1`

	example := Example{Id: id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&example.ExampleValue1,
		&example.ExampleValue2,
		&example.ExampleValue3,
		&example.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrExampleRecordNotFound
		} else {
			return nil, err
		}
	}

	return &example, nil
}

// Get all examples from DB
func (e ExampleDBConnection) GetAll(exampleValue2 string, exampleValue3 string, filters *Filters) ([]*Example, *Metadata, error) {
	totalRecords := 0
	result := []*Example{}

	// it’s not possible to use placeholder parameters for column names or SQL keywords (including ASC and DESC)
	// thats why use fmt.Sprintf for "ORDER BY"
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, example_value_1, example_value_2, example_value_3, created_at
		FROM examples
		WHERE (LOWER(example_value_2) = LOWER($1) OR $1 = '') 
		AND (LOWER(example_value_3) = LOWER($2) OR $2 = '') 
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.SortColumn, filters.SortDirection)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := e.DB.QueryContext(ctx, query,
		exampleValue2,
		exampleValue3,
		filters.limit(),
		filters.offset())
	if err != nil {
		return nil, &Metadata{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var row Example

		err := rows.Scan(
			&totalRecords,
			&row.Id,
			&row.ExampleValue1,
			&row.ExampleValue2,
			&row.ExampleValue3,
			&row.CreatedAt)
		if err != nil {
			return nil, &Metadata{}, err
		}

		result = append(result, &row)
	}

	if err = rows.Err(); err != nil {
		return nil, &Metadata{}, err
	}

	metadata := InitMetadata(totalRecords, filters.Page, filters.PageSize)

	return result, metadata, nil
}

// Update an example from DB
func (e ExampleDBConnection) Update(example *Example) error {
	query := "UPDATE examples Set"
	parameterCount := 1
	args := []any{}

	if validator.NotCero(example.ExampleValue1) {
		query += fmt.Sprintf(" example_value_1 = $%d,", parameterCount)
		parameterCount++
		args = append(args, example.ExampleValue1)
	}
	if validator.NotBlank(example.ExampleValue2) {
		query += fmt.Sprintf(" example_value_2 = $%d,", parameterCount)
		parameterCount++
		args = append(args, example.ExampleValue2)
	}
	if validator.NotBlank(example.ExampleValue3) {
		query += fmt.Sprintf(" example_value_3 = $%d,", parameterCount)
		parameterCount++
		args = append(args, example.ExampleValue3)
	}

	query = query[:len(query)-1]
	query += fmt.Sprintf(" WHERE id = $%d", parameterCount)
	query += " RETURNING example_value_1, example_value_2, example_value_3, created_at"
	args = append(args, example.Id)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, query, args...).Scan(
		&example.ExampleValue1,
		&example.ExampleValue2,
		&example.ExampleValue3,
		&example.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrExampleRecordNotFound
		} else {
			return err
		}
	}

	return nil
}

// Delete an example from DB
func (e ExampleDBConnection) Delete(id int64) error {
	if id < 1 {
		return ErrExampleRecordNotFound
	}

	query := `
		DELETE FROM examples
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := e.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrExampleRecordNotFound
	}

	return nil
}

// Automatically used when trying to encode this type to json.
func (e Example) MarshalJSON() ([]byte, error) {

	exampleValue12 := fmt.Sprintf("%v %s", e.ExampleValue1, e.ExampleValue2)

	// Use an Alias to avoid an infinity loop when call json.Marshal in the return
	type ExampleAlias Example

	aux := struct {
		ExampleAlias
		ExampleConcatValue string `json:"example_concat_value,omitempty"`
	}{
		ExampleAlias:       ExampleAlias(e),
		ExampleConcatValue: exampleValue12,
	}

	return json.Marshal(aux)
}

// validate input example fields
func (e *Example) ValidateExample(r *http.Request) *validator.Validator {
	v := validator.Validator{}

	if r.Method == http.MethodPost {
		v.Check(validator.NotBlank(e.ExampleValue2), "example_value_2", "must be provided")
		v.Check(validator.NotBlank(e.ExampleValue3), "example_value_3", "must be provided")
	}

	v.Check(validator.MinNumber(e.ExampleValue1, 0), "example_value_1", "must be a positive number")
	v.Check(validator.MaxChars(e.ExampleValue2, 4), "example_value_2", "must not be more than 4 bytes long")
	v.Check(validator.MaxChars(e.ExampleValue3, 50), "example_value_3", "must not be more than 50 bytes long")

	return &v
}
