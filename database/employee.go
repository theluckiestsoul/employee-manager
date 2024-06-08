package database

import (
	"database/sql"
	"errors"
)

type Employee struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Position string  `json:"position"`
	Salary   float64 `json:"salary"`
}

type EmployeeDB interface {
	CreateEmployee(employee Employee) (Employee, error)
	GetEmployeeByID(id int) (Employee, error)
	UpdateEmployee(employee Employee) error
	DeleteEmployee(id int) error
	ListEmployees(page, perPage int) ([]Employee, int, error)
}

type employeeDB struct {
	db *sql.DB
}

func NewEmployee(db *sql.DB) EmployeeDB {
	return &employeeDB{db: db}
}

func (e *employeeDB) CreateEmployee(employee Employee) (Employee, error) {
	query := `INSERT INTO employees (name, position, salary) VALUES ($1, $2, $3) RETURNING id`
	err := e.db.QueryRow(query, employee.Name, employee.Position, employee.Salary).Scan(&employee.ID)
	return employee, err
}

func (e *employeeDB) GetEmployeeByID(id int) (Employee, error) {
	var employee Employee
	query := `SELECT id, name, position, salary FROM employees WHERE id=$1`
	err := e.db.QueryRow(query, id).Scan(&employee.ID, &employee.Name, &employee.Position, &employee.Salary)
	if err == sql.ErrNoRows {
		return employee, errors.New("employee not found")
	}
	return employee, err
}

func (e *employeeDB) UpdateEmployee(employee Employee) error {
	query := `UPDATE employees SET name=$1, position=$2, salary=$3 WHERE id=$4`
	result, err := e.db.Exec(query, employee.Name, employee.Position, employee.Salary, employee.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.New("employee not found")
	}
	return nil
}

func (e *employeeDB) DeleteEmployee(id int) error {
	query := `DELETE FROM employees WHERE id=$1`
	result, err := e.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.New("employee not found")
	}
	return nil
}

func (e *employeeDB) ListEmployees(page, perPage int) ([]Employee, int, error) {
	var employees []Employee
	query := `
		SELECT id, name, position, salary, COUNT(*) OVER() AS total 
		FROM employees 
		ORDER BY id 
		LIMIT $1 OFFSET $2
	`
	rows, err := e.db.Query(query, perPage, (page-1)*perPage)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var total int
	for rows.Next() {
		var employee Employee
		if err := rows.Scan(&employee.ID, &employee.Name, &employee.Position, &employee.Salary, &total); err != nil {
			return nil, 0, err
		}
		employees = append(employees, employee)
	}

	return employees, total, nil
}
