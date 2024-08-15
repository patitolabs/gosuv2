package gosuv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SuvGradesResponse represents the response structure for grades
type SuvGradesResponse struct {
	PaymentStatus  string
	EnrollmentType string
	Courses        []SuvCurrentCourseGrades
	Semester       string
}

// SuvCurrentCoursesGrades represents the structure of each course's grades for the current semester.
type SuvCurrentCourseGrades struct {
	IdCurso       string   `json:"idcurso"`
	Curso         string   `json:"curso"`
	Vez           string   `json:"vez"`
	Promedio1     string   `json:"promedio1"`
	Promedio2     string   `json:"promedio2"`
	Promedio3     string   `json:"promedio3"`
	Promedio4     string   `json:"promedio4"`
	Promedio5     string   `json:"promedio5"`
	Promedio6     string   `json:"promedio6"`
	Sustitutorio  string   `json:"sustitutorio"`
	Promedio      string   `json:"promedio"`
	Aplazado      string   `json:"aplazado"`
	PromedioFinal string   `json:"pfinal"`
	Inhabilitado  string   `json:"inh"`
	Pesos         []string `json:"pesos"`
	Estados       []string `json:"estados"`
	EstadoFinal   string   `json:"estado_final"`
}

// GetSuvGradesResponse retrieves the current semester and its grades from SUV2.
func (c *SuvClient) GetSuvGradesResponse() (*SuvGradesResponse, error) {
	data := url.Values{
		"task": {"verNotasPeriodoActual"},
	}

	res, err := c.urlEncodedPostRequest(data, "/controller/alumnoController.php")
	if err != nil {
		return nil, err
	}

	suvGradesResponse, err := handleSuvGradesResponse(res)
	if err != nil {
		return nil, err
	}

	return suvGradesResponse, nil
}

// handleSuvGradesResponse handles the response from the grades request.
func handleSuvGradesResponse(res *http.Response) (*SuvGradesResponse, error) {
	bodyBuf := new(bytes.Buffer)
	bodyBuf.ReadFrom(res.Body)

	suvGradesResponse, err := unmarshalSuvGradesResponse(bodyBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return suvGradesResponse, nil
}

// unmarshalSuvGradesResponse unmarshals the JSON response into a SuvGradesResponse struct.
func unmarshalSuvGradesResponse(data []byte) (*SuvGradesResponse, error) {
	// Unmarshal the JSON response into a slice of interface{}.
	var rawResponse []interface{}
	if err := json.Unmarshal(data, &rawResponse); err != nil {
		return nil, err
	}

	// Check if the response has the expected structure.
	if len(rawResponse) != 4 {
		return nil, fmt.Errorf("unexpected response length: %d", len(rawResponse))
	}

	// Create an instance of SuvGradesResponse and populate it with the response data.
	suvGradesResponse := &SuvGradesResponse{}

	// Asign the first, second and fourth fields
	var ok bool
	if suvGradesResponse.PaymentStatus, ok = rawResponse[0].(string); !ok {
		return nil, fmt.Errorf("unexpected type for payment status: %T", rawResponse[0])
	}
	if suvGradesResponse.EnrollmentType, ok = rawResponse[1].(string); !ok {
		return nil, fmt.Errorf("unexpected type for enrollment type: %T", rawResponse[1])
	}
	if suvGradesResponse.Semester, ok = rawResponse[3].(string); !ok {
		return nil, fmt.Errorf("unexpected type for semester: %T", rawResponse[3])
	}

	// Marshal the third field (array of objects) back into JSON.
	coursesJSON, err := json.Marshal(rawResponse[2])
	if err != nil {
		return nil, fmt.Errorf("error marshaling items: %w", err)
	}

	// Unmarshal the JSON array of objects into a slice of SuvCurrentCourseGrades.
	if err := json.Unmarshal(coursesJSON, &suvGradesResponse.Courses); err != nil {
		return nil, fmt.Errorf("error unmarshaling items: %w", err)
	}

	// Remove trailing spaces and slashes from the payment status, enrollment type and semester.
	suvGradesResponse.PaymentStatus = strings.TrimSpace(strings.Trim(rawResponse[0].(string), "\""))
	suvGradesResponse.EnrollmentType = strings.Trim(rawResponse[1].(string), "\"")
	suvGradesResponse.Semester = strings.Trim(rawResponse[3].(string), "\"")

	return suvGradesResponse, nil
}
