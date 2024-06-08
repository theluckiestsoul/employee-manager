package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/theluckiestsoul/employeemanager/database"
)

var (
	ErrInvalidName     = errors.New("invalid name")
	ErrInvalidPosition = errors.New("invalid position")
	ErrInvalidSalary   = errors.New("invalid salary")
)

type handler struct {
	emp database.EmployeeDB
}

func NewHandler(db database.EmployeeDB) *handler {
	return &handler{emp: db}
}

// EmployeeResponse defines the response structure for an employee
type EmployeeResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Salary   int    `json:"salary"`
}

// EmployeeParams defines the body parameters for the CreateEmployeeHandler and UpdateEmployeeHandler
// @Param name body string true "Employee name"
// @Param position body string true "Employee position"
// @Param salary body float64 true "Employee salary"
type EmployeeParams struct {
	Name     string  `json:"name"`
	Position string  `json:"position"`
	Salary   float64 `json:"salary"`
}

func (e EmployeeParams) toEmployee() database.Employee {
	return database.Employee{
		Name:     e.Name,
		Position: e.Position,
		Salary:   e.Salary,
	}
}

func (e EmployeeParams) validate() error {
	if e.Name == "" {
		return ErrInvalidName
	}
	if e.Position == "" {
		return ErrInvalidPosition
	}
	if e.Salary <= 0 {
		return ErrInvalidSalary
	}
	return nil
}

// CreateEmployeeHandler creates a new employee
// @Summary Create a new employee
// @Description Create a new employee
// @Tags employees
// @Accept json
// @Produce json
// @Param body body EmployeeParams true "Employee body"
// @Success 201 {object} EmployeeResponse
// @Failure 400 {string} string "Invalid request payload"
// @Failure 500 {string} string "Internal server error"
// @Router /employees [post]
func (h *handler) CreateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	var employee EmployeeParams
	if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if err := employee.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	emp, err := h.emp.CreateEmployee(employee.toEmployee())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", r.URL.Path+"/"+strconv.Itoa(emp.ID))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toEmployeeResponse(emp))

}

// GetEmployeeHandler retrieves an employee by ID.
// @Summary Get an employee by ID
// @Description Get an employee by ID
// @Tags employees
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Success 200 {object} EmployeeResponse
// @Failure 400 {string} string "Invalid employee ID"
// @Failure 404 {string} string "Employee not found"
// @Router /employees/{id} [get]
func (h *handler) GetEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}
	employee, err := h.emp.GetEmployeeByID(id)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toEmployeeResponse(employee))
}

// UpdateEmployeeHandler updates an employee.
// @Summary Update an employee
// @Description Update an employee
// @Tags employees
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Param body body EmployeeParams true "Employee object that needs to be updated"
// @Success 200 {object} EmployeeResponse
// @Failure 400 {string} string "Invalid request payload"
// @Failure 404 {string} string "Employee not found"
// @Router /employees/{id} [put]
func (h *handler) UpdateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}
	var emp EmployeeParams
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if err := emp.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	empToUpdate := emp.toEmployee()
	empToUpdate.ID = id
	err = h.emp.UpdateEmployee(empToUpdate)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toEmployeeResponse(empToUpdate))
}

// DeleteEmployeeHandler deletes an employee by ID.
// @Summary Delete an employee by ID
// @Description Delete an employee by ID
// @Tags employees
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Success 204 {string} string "Employee deleted"
// @Failure 400 {string} string "Invalid employee ID"
// @Failure 404 {string} string "Employee not found"
// @Router /employees/{id} [delete]
func (h *handler) DeleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}
	if err := h.emp.DeleteEmployee(id); err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type ListEmployeesResponse struct {
	Employees []EmployeeResponse `json:"employees"`
	Total     int                `json:"total"`
}

// ListEmployeesHandler
// @Summary List employees
// @Description List employees
// @Tags employees
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Number of items per page"
// @Success 200 {object} ListEmployeesResponse
// @Failure 500 {string} string "Internal server error"
// @Router /employees [get]
func (h *handler) ListEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
	if err != nil || perPage <= 0 {
		perPage = 10
	}
	employees, total, err := h.emp.ListEmployees(page, perPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := ListEmployeesResponse{
		Employees: make([]EmployeeResponse, len(employees)),
		Total:     total,
	}
	for i, emp := range employees {
		response.Employees[i] = toEmployeeResponse(emp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func toEmployeeResponse(emp database.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:       emp.ID,
		Name:     emp.Name,
		Position: emp.Position,
		Salary:   int(emp.Salary),
	}
}
