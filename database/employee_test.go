package database

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bradleyjkemp/cupaloy/v2"
)

func TestCreateEmployee(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := NewEmployee(db)

	tests := []struct {
		name     string
		employee Employee
		wantErr  bool
		before   func(emp Employee, t *testing.T)
		after    func(t *testing.T)
	}{
		{
			name: "Successful Insert",
			employee: Employee{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   50000.0,
			},
			wantErr: false,
			before: func(emp Employee, t *testing.T) {
				mock.ExpectQuery(`INSERT INTO employees`).WithArgs(emp.Name, emp.Position, emp.Salary).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			after: func(t *testing.T) {
				mock.ExpectationsWereMet()
			},
		},
		{
			name: "Failed Insert",
			employee: Employee{
				Position: "Engineer",
				Salary:   50000.0,
			},
			wantErr: true,
			before: func(emp Employee, t *testing.T) {
				query := mock.ExpectQuery(`INSERT INTO employees`).WithArgs(emp.Name, emp.Position, emp.Salary).WillReturnError(errors.New("failed to   insert"))
				if query == nil {
					t.Errorf("error")
				}
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.employee, t)
			res, err := edb.CreateEmployee(tt.employee)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEmployee() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.after(t)
			cupaloy.SnapshotT(t, struct {
				Employee Employee
				Error    error
			}{
				res,
				err,
			})
		})
	}
}

func TestGetEmployeeByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := NewEmployee(db)

	tests := []struct {
		name    string
		id      int
		wantErr bool
		before  func(id int, t *testing.T)
		after   func(t *testing.T)
	}{
		{
			name:    "Successful Get",
			id:      1,
			wantErr: false,
			before: func(id int, t *testing.T) {
				rows := sqlmock.NewRows([]string{"id", "name", "position", "salary"}).AddRow(1, "John Doe", "Engineer", 50000.0)
				mock.ExpectQuery(`SELECT id, name, position, salary FROM employees WHERE id=\$1`).WithArgs(id).WillReturnRows(rows)
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
		{
			name:    "Failed Get",
			id:      1,
			wantErr: true,
			before: func(id int, t *testing.T) {
				mock.ExpectQuery(`SELECT id, name, position, salary FROM employees WHERE id=\$1`).WithArgs(id).WillReturnError(errors.New("failed to get"))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.id, t)
			res, err := edb.GetEmployeeByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEmployeeByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.after(t)
			cupaloy.SnapshotT(t, struct {
				Employee Employee
				Error    error
			}{res, err})
		})
	}
}

func TestUpdateEmployee(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := NewEmployee(db)

	tests := []struct {
		name     string
		employee Employee
		wantErr  bool
		before   func(emp Employee, t *testing.T)
		after    func(t *testing.T)
	}{
		{
			name: "Successful Update",
			employee: Employee{
				ID:       1,
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   50000.0,
			},
			wantErr: false,
			before: func(emp Employee, t *testing.T) {
				mock.ExpectExec(`UPDATE employees SET name=\$1, position=\$2, salary=\$3 WHERE id=\$4`).WithArgs(emp.Name, emp.Position, emp.Salary, emp.ID).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
		{
			name: "Failed Update",
			employee: Employee{
				ID:       1,
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   50000.0,
			},
			wantErr: true,
			before: func(emp Employee, t *testing.T) {
				mock.ExpectExec(`UPDATE employees SET name=\$1, position=\$2, salary=\$3 WHERE id=\$4`).WithArgs(emp.Name, emp.Position, emp.Salary, emp.ID).WillReturnError(errors.New("failed to update"))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.employee, t)
			err := edb.UpdateEmployee(tt.employee)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateEmployee() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.after(t)
		})
	}
}

func TestDeleteEmployee(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := NewEmployee(db)

	tests := []struct {
		name    string
		id      int
		wantErr bool
		before  func(id int, t *testing.T)
		after   func(t *testing.T)
	}{
		{
			name:    "Successful Delete",
			id:      1,
			wantErr: false,
			before: func(id int, t *testing.T) {
				mock.ExpectExec(`DELETE FROM employees WHERE id=\$1`).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
		{
			name:    "Failed Delete",
			id:      1,
			wantErr: true,
			before: func(id int, t *testing.T) {
				mock.ExpectExec(`DELETE FROM employees WHERE id=\$1`).WithArgs(id).WillReturnError(errors.New("failed to delete"))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.id, t)
			err := edb.DeleteEmployee(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteEmployee() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.after(t)
		})
	}
}

func TestListEmployees(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := NewEmployee(db)

	tests := []struct {
		name    string
		page    int
		perPage int
		wantErr bool
		before  func(page, perPage int, t *testing.T)
		after   func(t *testing.T)
	}{
		{
			name:    "Successful List",
			page:    1,
			perPage: 10,
			wantErr: false,
			before: func(page, perPage int, t *testing.T) {
				rows := sqlmock.NewRows([]string{"id", "name", "position", "salary", "total"}).
					AddRow(1, "John Doe", "Engineer", 50000.0, 1)
				mock.ExpectQuery(`SELECT id, name, position, salary, COUNT\(\*\) OVER\(\) AS total FROM employees ORDER BY id LIMIT \$1 OFFSET \$2`).
					WithArgs(perPage, (page-1)*perPage).
					WillReturnRows(rows)
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
		{
			name:    "Failed List",
			page:    1,
			perPage: 10,
			wantErr: true,
			before: func(page, perPage int, t *testing.T) {
				mock.ExpectQuery(`SELECT id, name, position, salary, COUNT\(\*\) OVER\(\) AS total FROM employees ORDER BY id LIMIT \$1 OFFSET \$2`).
					WithArgs(perPage, (page-1)*perPage).
					WillReturnError(errors.New("failed to list"))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(tt.page, tt.perPage, t)
			res, total, err := edb.ListEmployees(tt.page, tt.perPage)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListEmployees() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.after(t)
			cupaloy.SnapshotT(t, struct {
				Employees []Employee
				Total     int
				Error     error
			}{
				res,
				total,
				err,
			})
		})
	}

}
