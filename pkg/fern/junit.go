package fern

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/monforton/fern-cli/pkg/models"
)

func Report(reportDirectory string) {

	// Use os.ReadDir to list directory contents
	entries, err := os.ReadDir(reportDirectory)
	if err != nil {
		panic(fmt.Sprintf("Failed to read directory: %v", err))
	}

	startTime := time.Now()
	var suiteRuns []SuiteRun

	for _, entry := range files {
		if !entry.IsDir() {
			reportPath := filepath.Join(reportDirectory, entry.Name())
			suiteRun, err := processFile(reportPath)
			if err != nil {
				panic(fmt.Sprintf("Failed to process file %s: %v", reportPath, err))
			}
			suiteRuns = append(suiteRuns, suiteRun)
		}
	}

	testRun := TestRun{
		TestProjectName: "TestProj",
		StartTime:       startTime,
		EndTime:         time.Now(),
		SuiteRuns:       suiteRuns,
	}

	// if err := sendTestRun(testRun, "http://localhost:8080"); err != nil {
	// 	panic(fmt.Sprintf("Failed to send test run: %v", err))
	// }

}

func processFile(filePath string) ([]models.SuiteRun, error) {
	var testSuites models.TestSuites
	var suiteRuns []models.SuiteRun

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	if err := xml.Unmarshal(byteValue, &testSuites); err != nil {
		return nil, fmt.Errorf("Failed to parse XML from file %s: %v", filePath, err)
	}

	for _, suite := range testSuites.TestSuites {
		run, err := parseTestSuite(suite)
		if err != nil {
			return nil, err
		}
		suiteRuns = append(suiteRuns, run)
	}
	return suiteRuns, err
}

func parseTestSuite(testSuite models.TestSuite) (suiteRun models.SuiteRun, err error) {
	suiteRun.SuiteName = testSuite.Name
	suiteRun.TestRunID = 0
	suiteRun.StartTime, err = time.Parse(time.RFC3339, testSuite.Timestamp)
	suiteRun.EndTime, err = getEndTime(suiteRun.StartTime, testSuite.Time)
	if err != nil {
		err = fmt.Errorf("Failed to parse TestSuite %s: %v", testSuite.Name, err)
		return
	}

	return
}

func getEndTime(startTime time.Time, duration string) (endTime time.Time, err error) {
	ms, err := time.ParseDuration(duration + "s")
	endTime = startTime.Add(ms)
	return
}

func sendTestRun(testRun models.TestRun, serviceUrl string) error {
	payload, err := json.Marshal(testRun)
	if err != nil {
		return fmt.Errorf("failed to serialize test run: %v", err)
	}

	resp, err := http.Post(serviceUrl+"/api/testrun", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	return nil
}
