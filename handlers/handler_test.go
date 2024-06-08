package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/go-chi/chi/v5"
	"github.com/theluckiestsoul/employeemanager/database"
)

func dumpResponse(t *testing.T, r *http.Response) string {
	t.Helper()
	body, err := httputil.DumpResponse(r, true)
	if err != nil {
		t.Fatalf("failed to dump response: %v", err)
	}
	return string(body)
}

func TestEmployeeCreateParamsValidate(t *testing.T) {
	tests := []struct {
		name      string
		params    EmployeeParams
		wantError bool
	}{
		{
			name: "valid params",
			params: EmployeeParams{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   5000.0,
			},
			wantError: false,
		},
		{
			name: "invalid name",
			params: EmployeeParams{
				Name:     "",
				Position: "Engineer",
				Salary:   5000.0,
			},
			wantError: true,
		},
		{
			name: "invalid position",
			params: EmployeeParams{
				Name:     "John Doe",
				Position: "",
				Salary:   5000.0,
			},
			wantError: true,
		},
		{
			name: "invalid salary",
			params: EmployeeParams{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   0.0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.validate()
			if (err != nil) != tt.wantError {
				t.Errorf("validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCreateEmployeeHandler(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := database.NewEmployee(db)

	tests := []struct {
		name           string
		params         *EmployeeParams
		wantError      bool
		before         func(t *testing.T, emp *EmployeeParams)
		after          func(t *testing.T)
		expectedStatus int
	}{
		{
			name: "create employee",
			params: &EmployeeParams{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   5000.0,
			},
			wantError:      false,
			expectedStatus: http.StatusCreated,
			before: func(t *testing.T, emp *EmployeeParams) {
				mock.ExpectQuery(`INSERT INTO employees`).WithArgs(emp.Name, emp.Position, emp.Salary).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
		{
			name: "invalid name",
			params: &EmployeeParams{
				Name:     "",
				Position: "Engineer",
				Salary:   5000.0,
			},
			wantError:      true,
			expectedStatus: http.StatusBadRequest,
			before: func(t *testing.T, emp *EmployeeParams) {
			},
			after: func(t *testing.T) {
			},
		},
		{
			name: "invalid position",
			params: &EmployeeParams{
				Name:     "John Doe",
				Position: "",
				Salary:   5000.0,
			},
			wantError:      true,
			expectedStatus: http.StatusBadRequest,
			before: func(t *testing.T, emp *EmployeeParams) {
			},
			after: func(t *testing.T) {},
		},
		{
			name: "invalid salary",
			params: &EmployeeParams{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   0.0,
			},
			wantError:      true,
			expectedStatus: http.StatusBadRequest,
			before: func(t *testing.T, emp *EmployeeParams) {
			},
			after: func(t *testing.T) {},
		},
		{
			name: "failed insert",
			params: &EmployeeParams{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   5000.0,
			},
			wantError:      true,
			expectedStatus: http.StatusInternalServerError,
			before: func(t *testing.T, emp *EmployeeParams) {
				mock.ExpectQuery(`INSERT INTO employees`).WithArgs(emp.Name, emp.Position, emp.Salary).WillReturnError(errors.New("failed to   insert"))
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
			tt.before(t, tt.params)
			defer tt.after(t)

			h := NewHandler(edb)
			body, _ := json.Marshal(tt.params)
			req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			h.CreateEmployeeHandler(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
			res := rr.Result()
			defer res.Body.Close()

			cupaloy.SnapshotT(t, dumpResponse(t, res))

		})
	}
}

func TestUpdateEmployeeHandler(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := database.NewEmployee(db)

	tests := []struct {
		name           string
		params         *EmployeeParams
		id             int
		wantError      bool
		before         func(t *testing.T, emp *EmployeeParams, id int)
		after          func(t *testing.T)
		expectedStatus int
	}{
		{
			name: "update employee",
			params: &EmployeeParams{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   5000.0,
			},
			id:             1,
			wantError:      false,
			expectedStatus: http.StatusOK,
			before: func(t *testing.T, emp *EmployeeParams, id int) {
				mock.ExpectExec(`UPDATE employees SET name=\$1, position=\$2, salary=\$3 WHERE id=\$4`).WithArgs(emp.Name, emp.Position, emp.Salary, id).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			after: func(t *testing.T) {
				err := mock.ExpectationsWereMet()
				if err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			},
		},
		{
			name: "invalid name",
			params: &EmployeeParams{
				Name:     "",
				Position: "Engineer",
				Salary:   5000.0,
			},
			id:             1,
			wantError:      true,
			expectedStatus: http.StatusBadRequest,
			before: func(t *testing.T, emp *EmployeeParams, id int) {
			},
			after: func(t *testing.T) {
			},
		},
		{
			name: "invalid position",
			params: &EmployeeParams{
				Name:     "John Doe",
				Position: "",
				Salary:   5000.0,
			},
			id:             1,
			wantError:      true,
			expectedStatus: http.StatusBadRequest,
			before: func(t *testing.T, emp *EmployeeParams, id int) {
			},
			after: func(t *testing.T) {},
		},
		{
			name: "invalid salary",
			params: &EmployeeParams{
				Name:     "John Doe",
				Position: "Engineer",
				Salary:   0.0,
			},
			id:             1,
			wantError:      true,
			expectedStatus: http.StatusBadRequest,
			before: func(t *testing.T, emp *EmployeeParams, id int) {
			},
			after: func(t *testing.T) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(t, tt.params, tt.id)
			defer tt.after(t)

			h := NewHandler(edb)
			body, _ := json.Marshal(tt.params)

			r := chi.NewRouter()
			r.Put("/employees/{id}", h.UpdateEmployeeHandler)

			url := fmt.Sprintf("/employees/%d", tt.id)
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
			res := rr.Result()
			defer res.Body.Close()

			cupaloy.SnapshotT(t, dumpResponse(t, res))
		})
	}
}

func TestGetEmployeeHandler(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := database.NewEmployee(db)

	tests := []struct {
		name           string
		id             int
		wantError      bool
		before         func(id int, t *testing.T)
		after          func(t *testing.T)
		expectedStatus int
	}{
		{
			name:           "get employee",
			id:             1,
			wantError:      false,
			expectedStatus: http.StatusOK,
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
			name:           "failed get",
			id:             1,
			wantError:      true,
			expectedStatus: http.StatusNotFound,
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
			defer tt.after(t)

			h := NewHandler(edb)

			r := chi.NewRouter()
			r.Get("/employees/{id}", h.GetEmployeeHandler)

			url := fmt.Sprintf("/employees/%d", tt.id)
			req, _ := http.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
			res := rr.Result()
			defer res.Body.Close()

			cupaloy.SnapshotT(t, dumpResponse(t, res))
		})
	}
}

func TestDeleteEmployeeHandler(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := database.NewEmployee(db)

	tests := []struct {
		name           string
		id             int
		wantError      bool
		before         func(id int, t *testing.T)
		after          func(t *testing.T)
		expectedStatus int
	}{
		{
			name:           "delete employee",
			id:             1,
			wantError:      false,
			expectedStatus: http.StatusNoContent,
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
			name:           "failed delete",
			id:             1,
			wantError:      true,
			expectedStatus: http.StatusNotFound,
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
			defer tt.after(t)

			h := NewHandler(edb)

			r := chi.NewRouter()
			r.Delete("/employees/{id}", h.DeleteEmployeeHandler)

			url := fmt.Sprintf("/employees/%d", tt.id)
			req, _ := http.NewRequest("DELETE", url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
			res := rr.Result()
			defer res.Body.Close()

			cupaloy.SnapshotT(t, dumpResponse(t, res))
		})
	}
}

func TestListEmployeesHandler(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	edb := database.NewEmployee(db)

	tests := []struct {
		name           string
		wantError      bool
		before         func(page, perPage int, t *testing.T)
		after          func(t *testing.T)
		expectedStatus int
		page           int
		perPage        int
	}{
		{
			name:           "list employees",
			wantError:      false,
			expectedStatus: http.StatusOK,
			page:           1,
			perPage:        10,
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
			name:           "failed list",
			wantError:      true,
			page:           1,
			perPage:        10,
			expectedStatus: http.StatusInternalServerError,
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
			defer tt.after(t)

			h := NewHandler(edb)

			r := chi.NewRouter()
			r.Get("/employees", h.ListEmployeesHandler)

			req, _ := http.NewRequest("GET", "/employees", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
			res := rr.Result()
			defer res.Body.Close()

			cupaloy.SnapshotT(t, dumpResponse(t, res))
		})
	}
}
