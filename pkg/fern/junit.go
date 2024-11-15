package fern

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/monforton/fern-cli/pkg/models"
)

func ReportJunit(projectName string, reportDirectory string, fernUrl string) {
	var testRun models.TestRun
	testRun.ID = 0
	testRun.TestProjectName = projectName

	log.Default().Print("\nParsing reports...")
	if err := processDir(&testRun, reportDirectory); err != nil {
		log.Default().Println("FAILED")
		panic(err)
	}
	log.Default().Println("Parsing reports succeeded!")

	log.Default().Printf("Sending reports to %s...", fernUrl)
	if err := sendTestRun(testRun, fernUrl); err != nil {
		log.Default().Println("FAILED")
		panic(err)
	}
	log.Default().Println("Sending reports succeeded!")

}

func processDir(testRun *models.TestRun, currentPath string) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read directory: %v", err))
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			reportPath := filepath.Join(currentPath, entry.Name())
			suiteRun, err := processFile(reportPath)
			if err != nil {
				return fmt.Errorf("Failed to process file %s: %v", reportPath, err)
			}
			testRun.SuiteRuns = append(testRun.SuiteRuns, suiteRun...)
		} else {
			newPath := currentPath + "/" + entry.Name()
			if err = processDir(testRun, newPath); err != nil {
				return err
			}
		}
	}
	return err
}

func processFile(filePath string) ([]models.SuiteRun, error) {
	var testSuites models.TestSuites
	var testSuite models.TestSuite
	var suiteRuns []models.SuiteRun

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	if err := xml.Unmarshal(byteValue, &testSuites); err != nil {
		if err = xml.Unmarshal(byteValue, &testSuite); err != nil {
			return nil, fmt.Errorf("Failed to parse XML from file %s: %v", filePath, err)
		} else {
			testSuites.TestSuites = append(testSuites.TestSuites, testSuite)
		}
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

	runningTime := suiteRun.StartTime
	for i, testcase := range testSuite.TestCases {
		status := ""
		message := ""
		if len(testcase.Failures) > 0 {
			status = "Failure"
			message = testcase.Failures[0].Message + "\n" + testcase.Failures[0].Content
		} else if len(testcase.Errors) > 0 {
			status = "Failure"
			message = testcase.Errors[0].Message + "\n" + testcase.Errors[0].Content
		} else if len(testcase.Skips) > 0 {
			status = "Skipped"
		} else {
			status = "Success"
		}

		after, erri := getEndTime(runningTime, testSuite.Time)
		if erri != nil {
			err = fmt.Errorf("Failed to parse TestSuite %s: %v", testSuite.Name, err)
			return
		}

		sr := models.SpecRun{
			ID:              uint64(i),
			SuiteID:         suiteRun.ID,
			SpecDescription: testcase.ClassName,
			Status:          status,
			Message:         message,
			Tags:            []models.Tag{},
			StartTime:       runningTime,
			EndTime:         after,
		}
		runningTime = after

		suiteRun.SpecRuns = append(suiteRun.SpecRuns, sr)
	}

	return
}

func sendTestRun(testRun models.TestRun, serviceUrl string) error {
	payload, err := json.Marshal(testRun)
	if err != nil {
		return fmt.Errorf("Failed to serialize test run: %v", err)
	}

	resp, err := http.Post(serviceUrl+"/api/testrun", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Unexpected response code: %d", resp.StatusCode)
	}
	return nil
}

func getEndTime(startTime time.Time, duration string) (endTime time.Time, err error) {
	ms, err := time.ParseDuration(duration + "s")
	endTime = startTime.Add(ms)
	return
}
