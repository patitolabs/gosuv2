package gosuv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// Task to retrieve the current semester grades.
	taskGetCurrentSemesterGrades string = "verNotasPeriodoActual"

	// Passed status.
	PassedStatus int = 1
	// Not passed status.
	NotPassedStatus int = 0
)

// SuvGradesResponse represents the response structure for grades
type SuvGradesResponse struct {
	PaymentStatus  string
	EnrollmentType string
	Courses        []SuvCurrentCourseGrades
	Semester       string
}

// SuvCurrentCourseGrades represents the structure of each course's grades for the current semester.
type SuvCurrentCourseGrades struct {
	CourseID     int       `json:"idcurso,string"`
	CourseName   string    `json:"curso"`
	Attempt      int       `json:"vez,string"`
	Average1     float32   `json:"promedio1,string"`
	Average2     float32   `json:"promedio2,string"`
	Average3     float32   `json:"promedio3,string"`
	Average4     float32   `json:"promedio4,string"`
	Average5     float32   `json:"promedio5,string"`
	Average6     float32   `json:"promedio6,string"`
	Substitute   float32   `json:"sustitutorio,string"`
	Average      float32   `json:"promedio,string"`
	Postponed    float32   `json:"aplazado,string"`
	FinalAverage float32   `json:"pfinal,string"`
	Disabled     bool      `json:"inh,string"`
	Weights      []float32 `json:"pesos"`
	Statuses     []int     `json:"estados"`
	FinalStatus  int       `json:"estado_final,string"`
}

// GetSuvGradesResponse retrieves the current semester and its grades from SUV2.
func (c *SuvClient) GetSuvGradesResponse() (*SuvGradesResponse, error) {
	data := url.Values{
		"task": {taskGetCurrentSemesterGrades},
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

	// Assign the first, second and fourth fields
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

	// Process courses with custom handling for string-to-numeric conversions
	var rawCourses []map[string]interface{}
	if err := json.Unmarshal(coursesJSON, &rawCourses); err != nil {
		return nil, fmt.Errorf("error unmarshaling courses: %w", err)
	}

	// Process each course
	suvGradesResponse.Courses = make([]SuvCurrentCourseGrades, len(rawCourses))
	for i, rawCourse := range rawCourses {
		course := SuvCurrentCourseGrades{}

		// Convert string ID to int
		if idStr, ok := rawCourse["idcurso"].(string); ok {
			course.CourseID, _ = strconv.Atoi(idStr)
		}

		// Assign course name
		if curso, ok := rawCourse["curso"].(string); ok {
			course.CourseName = curso
		}

		// Convert string Vez to int
		if vezStr, ok := rawCourse["vez"].(string); ok {
			course.Attempt, _ = strconv.Atoi(vezStr)
		}

		// Convert string averages to float32
		if p1, ok := rawCourse["promedio1"].(string); ok && p1 != "" {
			val, _ := strconv.ParseFloat(p1, 32)
			course.Average1 = float32(val)
		}
		if p2, ok := rawCourse["promedio2"].(string); ok && p2 != "" {
			val, _ := strconv.ParseFloat(p2, 32)
			course.Average2 = float32(val)
		}
		if p3, ok := rawCourse["promedio3"].(string); ok && p3 != "" {
			val, _ := strconv.ParseFloat(p3, 32)
			course.Average3 = float32(val)
		}
		if p4, ok := rawCourse["promedio4"].(string); ok && p4 != "" {
			val, _ := strconv.ParseFloat(p4, 32)
			course.Average4 = float32(val)
		}
		if p5, ok := rawCourse["promedio5"].(string); ok && p5 != "" {
			val, _ := strconv.ParseFloat(p5, 32)
			course.Average5 = float32(val)
		}
		if p6, ok := rawCourse["promedio6"].(string); ok && p6 != "" {
			val, _ := strconv.ParseFloat(p6, 32)
			course.Average6 = float32(val)
		}

		// Convert other numeric fields
		if sust, ok := rawCourse["sustitutorio"].(string); ok && sust != "" {
			val, _ := strconv.ParseFloat(sust, 32)
			course.Substitute = float32(val)
		}
		if prom, ok := rawCourse["promedio"].(string); ok && prom != "" {
			val, _ := strconv.ParseFloat(prom, 32)
			course.Average = float32(val)
		}
		if apl, ok := rawCourse["aplazado"].(string); ok && apl != "" {
			val, _ := strconv.ParseFloat(apl, 32)
			course.Postponed = float32(val)
		}
		if pf, ok := rawCourse["pfinal"].(string); ok && pf != "" {
			val, _ := strconv.ParseFloat(pf, 32)
			course.FinalAverage = float32(val)
		}
		if inh, ok := rawCourse["inh"].(string); ok {
			inhVal, _ := strconv.Atoi(inh)
			course.Disabled = inhVal == 1
		}

		if ef, ok := rawCourse["estado_final"].(string); ok {
			course.FinalStatus, _ = strconv.Atoi(ef)
		}

		// Convert string arrays to numeric arrays
		if pesos, ok := rawCourse["pesos"].([]interface{}); ok {
			course.Weights = make([]float32, len(pesos))
			for j, peso := range pesos {
				if pesoStr, ok := peso.(string); ok && pesoStr != "" {
					val, _ := strconv.ParseFloat(pesoStr, 32)
					course.Weights[j] = float32(val)
				}
			}
		}

		if estados, ok := rawCourse["estados"].([]interface{}); ok {
			course.Statuses = make([]int, len(estados))
			for j, estado := range estados {
				if estadoStr, ok := estado.(string); ok && estadoStr != "" {
					course.Statuses[j], _ = strconv.Atoi(estadoStr)
				}
			}
		}

		suvGradesResponse.Courses[i] = course
	}

	// Remove trailing spaces and slashes from the payment status, enrollment type and semester.
	suvGradesResponse.PaymentStatus = strings.TrimSpace(strings.Trim(rawResponse[0].(string), "\""))
	suvGradesResponse.EnrollmentType = strings.Trim(rawResponse[1].(string), "\"")
	suvGradesResponse.Semester = strings.Trim(rawResponse[3].(string), "\"")

	return suvGradesResponse, nil
}
