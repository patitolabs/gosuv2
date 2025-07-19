package gosuv2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// Path for the student search controller.
	studentSearchControllerPath string = "/controller/buscarAlumnoController.php"
	// Path for the professor search controller.
	professorSearchControllerPath string = "/controller/buscarDocenteController.php"

	// Task to search for a student by name and lastname.
	taskSearchStudentByName string = "buscarAlumno"
	// Task to search for a student by code.
	taskSearchStudentByCode string = "buscarCodigo"
	// Task to search for a student by DNI.
	taskSearchStudentByDni string = "buscarDNI"
	// Task to search for a professor by name and lastname.
	taskSearchProfessorByName string = "buscarDocente"
)

// SearchBasicResponse is the common interface for search responses.
type SearchBasicResponse interface{}

// StudentBasicResponse represents the basic response structure for a student.
type StudentBasicResponse struct {
	StudentID   string `json:"idalumno"`
	StudentName string `json:"alumno"`
	DNI         string `json:"dni"`
}

// ProfessorBasicResponse represents the basic response structure for a professor.
type ProfessorBasicResponse struct {
	Code          string `json:"codigo"`
	ProfessorName string `json:"docente"`
	DNI           string `json:"dni"`
	WorkerID      string `json:"idtrabajador"`
}

// SearchStudentByName searches for a student by name and lastname.
func (c *SuvClient) SearchStudentByName(name, lastname string) (*[]StudentBasicResponse, error) {
	data := url.Values{
		"task":     {taskSearchStudentByName},
		"nombre":   {name},
		"apellido": {lastname},
	}

	return c.searchGeneric(data, studentSearchControllerPath, c.parseSearchStudentResponse).(*[]StudentBasicResponse), nil
}

// SearchStudentByCode searches for a student by code.
func (c *SuvClient) SearchStudentByCode(code string) (*[]StudentBasicResponse, error) {
	data := url.Values{
		"task":   {taskSearchStudentByCode},
		"codigo": {code},
	}

	return c.searchGeneric(data, studentSearchControllerPath, c.parseSearchStudentResponse).(*[]StudentBasicResponse), nil
}

// SearchStudentByDni searches for a student by DNI.
func (c *SuvClient) SearchStudentByDni(dni string) (*[]StudentBasicResponse, error) {
	data := url.Values{
		"task": {taskSearchStudentByDni},
		"dni":  {dni},
	}

	return c.searchGeneric(data, studentSearchControllerPath, c.parseSearchStudentResponse).(*[]StudentBasicResponse), nil
}

// SearchStudent searches for a student by code, name and lastname, or DNI.
func (c *SuvClient) SearchStudent(code, name, lastname, dni string) (*[]StudentBasicResponse, error) {
	if code != "" {
		return c.SearchStudentByCode(code)
	}
	if name != "" && lastname != "" {
		return c.SearchStudentByName(name, lastname)
	}
	if dni != "" {
		return c.SearchStudentByDni(dni)
	}
	return nil, fmt.Errorf("no valid search criteria provided")
}

// SearchProfessor searches for a professor by name and lastname.
func (c *SuvClient) SearchProfessor(name, lastname string) (*[]ProfessorBasicResponse, error) {
	data := url.Values{
		"task":     {taskSearchProfessorByName},
		"nombre":   {name},
		"apellido": {lastname},
	}

	return c.searchGeneric(data, professorSearchControllerPath, c.parseSearchProfessorResponse).(*[]ProfessorBasicResponse), nil
}

// searchGeneric performs a generic search using the provided data and endpoint.
// It handles the common logic of making the HTTP request and parsing the response.
func (c *SuvClient) searchGeneric(data url.Values, path string, parser func(*http.Response) SearchBasicResponse) SearchBasicResponse {
	res, err := c.urlEncodedPostRequest(data, path)
	if err != nil {
		return nil
	}

	return parser(res)
}

// parseSearchStudentResponse handles the response for student search requests.
func (c *SuvClient) parseSearchStudentResponse(res *http.Response) SearchBasicResponse {
	var students []StudentBasicResponse

	err := json.NewDecoder(res.Body).Decode(&students)
	if err != nil {
		return nil
	}

	return &students
}

// parseSearchProfessorResponse handles the response for professor search requests.
func (c *SuvClient) parseSearchProfessorResponse(res *http.Response) SearchBasicResponse {
	var professors []ProfessorBasicResponse

	err := json.NewDecoder(res.Body).Decode(&professors)
	if err != nil {
		return nil
	}

	return &professors
}
