package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

// Global database connection
var db *sql.DB

// Helper function to count rows in a SQL result set
// Used for verifying/debugging database operations
func getRowCount(rows *sql.Rows) (int, error) {
	count := 0
	for rows.Next() {
		count++
	}
	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

// Helper function to check if a string exists in a slice
// Used for room availability checking
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Calculates time difference between two time strings in minutes
// Used to check if a lesson/booking is long enough to be considered
func TimeDifferenceInMinutes(time1, time2 string) (int, error) {
	// Define the layout of the time format (24 hour format)
	layout := "15:04:05"

	// Parse the time strings into time.Time objects
	t1, err := time.Parse(layout, time1)
	if err != nil {
		return 0, fmt.Errorf("error parsing time1: %v", err)
	}
	t2, err := time.Parse(layout, time2)
	if err != nil {
		return 0, fmt.Errorf("error parsing time2: %v", err)
	}

	// Calculate the difference
	duration := t2.Sub(t1)

	// Convert the duration to minutes
	minutes := int(duration.Minutes())

	return minutes, nil
}

// Initialize and clean database
// WARNING: This deletes all existing data!
func openDB() {
	var err error
	// Open SQLite database file
	db, err = sql.Open("sqlite3", "./brake-room.db")
	if err != nil {
		log.Fatal(err)
	}

	// Set database connection pool settings
	db.SetMaxOpenConns(5) // For SQLite, usually one connection is enough
	db.SetMaxIdleConns(1)

	// Clear all existing data from tables
	_, err = db.Exec("DELETE FROM ocupiedRoom")
	fmt.Println("Deleted ocupiedRoom")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM rooms")
	fmt.Println("Deleted rooms")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM times")
	fmt.Println("Deleted times")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM canceled")
	fmt.Println("Deleted canceled")
	if err != nil {
		log.Fatal(err)
	}
}

// Simulates a login request to get a session token
// Uses curl to exactly replicate browser behavior
func sendLoginRequest() []byte {
	// Define the curl command as a string with all necessary headers
	// This appears to be copied from a browser's network inspector
	curlCmd := `curl --path-as-is -i -s -k -X $'POST' \
    -H $'Host: intranet.tam.ch' -H $'Content-Length: 97' -H $'Cache-Control: max-age=0' -H $'Sec-Ch-Ua: \"Chromium\";v=\"127\", \"Not)A;Brand\";v=\"99\"' -H $'Sec-Ch-Ua-Mobile: ?0' -H $'Sec-Ch-Ua-Platform: \"macOS\"' -H $'Accept-Language: en-US' -H $'Upgrade-Insecure-Requests: 1' -H $'Origin: https://intranet.tam.ch' -H $'Content-Type: application/x-www-form-urlencoded' -H $'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.6533.100 Safari/537.36' -H $'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' -H $'Sec-Fetch-Site: same-origin' -H $'Sec-Fetch-Mode: navigate' -H $'Sec-Fetch-User: ?1' -H $'Sec-Fetch-Dest: document' -H $'Referer: https://intranet.tam.ch/kzu' -H $'Accept-Encoding: gzip, deflate, br' -H $'Priority: u=0, i' \
    -b $'school=kzu; username=studentName; sturmuser=StudentName; sturmsession=87f6ff4811f8175ec2da9d08e97fdad8' \
    --data-binary $'hash=66c02d6f5436e6.99763438&loginschool=kzu&loginuser=studentName&loginpassword=studentPassword' \
    $'https://intranet.tam.ch/kzu'`

	// Execute the curl command using bash
	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return nil
	} else {
		return output
	}
}

// Convert response bytes to string lines for processing
func getreqeststring(byteInput []byte) []string {
	outputStr := string(byteInput)
	lines := strings.Split(outputStr, "\n")
	return lines
}

// Extract session token from response headers
func extractSturmsession(lines []string) string {
	var sturmSession string

	// Look for the sturmsession cookie in response headers
	for _, line := range lines {
		if strings.HasPrefix(line, "set-cookie: sturmsession=") {
			parts := strings.Split(line, ";")
			if len(parts) > 0 {
				sturmSession = strings.Split(parts[0], "=")[1]
				break
			}
		}
	}

	return sturmSession
}

// Fetch timetable data using the session token
func sendDataRequest(sturmSession string) []byte {
	start := time.Now()
	// Calculate date range starting from last Monday
	offset := (int(time.Monday) - int(start.Weekday()) + 7) % 7
	lastMonday := start.AddDate(0, 0, -offset).Truncate(24 * time.Hour).Add(3 * time.Hour) // Set to last Monday at 3 AM
	startDate := lastMonday.Unix() * 1000
	endDate := startDate + (432000 * 1000) // Add 5 days in milliseconds
	fmt.Println(startDate)
	fmt.Println(endDate)

	// Define the curl command for fetching timetable
	// Includes a long list of class IDs to check
	curlCmd := fmt.Sprintf(`curl --path-as-is -i -s -k -X $'POST' \
    -H $'Host: intranet.tam.ch' -H $'Content-Length: 1067' -H $'Sec-Ch-Ua: \"Chromium\";v=\"127\", \"Not)A;Brand\";v=\"99\"' -H $'Accept-Language: en-US' -H $'Sec-Ch-Ua-Mobile: ?0' -H $'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.6533.100 Safari/537.36' -H $'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' -H $'Accept: application/json, text/javascript, */*; q=0.01' -H $'X-Requested-With: XMLHttpRequest' -H $'Sec-Ch-Ua-Platform: \"macOS\"' -H $'Origin: https://intranet.tam.ch' -H $'Sec-Fetch-Site: same-origin' -H $'Sec-Fetch-Mode: cors' -H $'Sec-Fetch-Dest: empty' -H $'Referer: https://intranet.tam.ch/kzu/timetable/classbook' -H $'Accept-Encoding: gzip, deflate, br' -H $'Priority: u=1, i' \
    -b $'school=kzu; username=studentName; sturmuser=studentName; sturmsession=%s' \
    --data-binary $'startDate=%d&endDate=%d&classId%%5B%%5D=2586&classId%%5B%%5D=2587&classId%%5B%%5D=2588&classId%%5B%%5D=2589&classId%%5B%%5D=2590&classId%%5B%%5D=2591&classId%%5B%%5D=2592&classId%%5B%%5D=2550&classId%%5B%%5D=2571&classId%%5B%%5D=2572&classId%%5B%%5D=2573&classId%%5B%%5D=2574&classId%%5B%%5D=2575&classId%%5B%%5D=2593&classId%%5B%%5D=2594&classId%%5B%%5D=2595&classId%%5B%%5D=2596&classId%%5B%%5D=2597&classId%%5B%%5D=2598&classId%%5B%%5D=2599&classId%%5B%%5D=2600&classId%%5B%%5D=2601&classId%%5B%%5D=2602&classId%%5B%%5D=2585&classId%%5B%%5D=2583&classId%%5B%%5D=2584&classId%%5B%%5D=2581&classId%%5B%%5D=2582&classId%%5B%%5D=2580&classId%%5B%%5D=2576&classId%%5B%%5D=2577&classId%%5B%%5D=2578&classId%%5B%%5D=2579&classId%%5B%%5D=2569&classId%%5B%%5D=2570&classId%%5B%%5D=2565&classId%%5B%%5D=2566&classId%%5B%%5D=2567&classId%%5B%%5D=2568&classId%%5B%%5D=2561&classId%%5B%%5D=2562&classId%%5B%%5D=2563&classId%%5B%%5D=2564&classId%%5B%%5D=2552&classId%%5B%%5D=2553&classId%%5B%%5D=2554&classId%%5B%%5D=2555&classId%%5B%%5D=2556&classId%%5B%%5D=2557&classId%%5B%%5D=2551&classId%%5B%%5D=2558&classId%%5B%%5D=2559&classId%%5B%%5D=2560&holidaysOnly=0' \
    $'https://intranet.tam.ch/kzu/timetable/ajax-get-timetable'`, sturmSession, startDate, endDate)

	// Execute the curl command
	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Error executing command:", err)
		return nil
	} else {
		return output
	}
}

// Process the JSON response and extract room information
func getRoomsList(stringInput string) []string {
	// Parse JSON response into a map
	var result map[string]interface{}
	err := json.Unmarshal([]byte(stringInput), &result)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil
	}

	// Get the data array from response
	data, ok := result["data"].([]interface{})
	if !ok {
		fmt.Println("Error: 'data' field is not in the expected format")
		return nil
	}

	var rooms []string

	// Initialize database
	openDB()

	// Process each entry in the data array
	for _, item := range data {
		entry, ok := item.(map[string]interface{})
		if !ok {
			fmt.Println("Error: item is not in the expected format")
			continue
		}

		// Extract room information
		roomId, ok := entry["roomName"].(string)
		if !ok || len(roomId) == 0 {
			fmt.Println("Error: 'roomName' field is not in the expected format or is empty")
			continue
		}

		startTime, ok := entry["lessonStart"].(string)
		if !ok || len(startTime) == 0 {
			fmt.Println("Error: 'startTime' field is not in the expected format or is empty")
			continue
		}

		startDate, ok := entry["lessonDate"].(string)
		if !ok || len(startDate) == 0 {
			fmt.Println("Error: 'startTime' field is not in the expected format or is empty")
			continue
		}

		endtime, ok := entry["lessonEnd"].(string)
		if !ok || len(startDate) == 0 {
			fmt.Println("Error: 'endtime' field is not in the expected format or is empty")
			continue
		}

		canceled, ok := entry["timetableEntryTypeShort"].(string)
		if !ok {
			fmt.Println("Error: 'canceled' field is not in the expected format or is empty")
			continue
		}

		// Get time slot ID
		timeId := getTimeID(startTime, startDate, endtime)

		// Handle canceled rooms
		if canceled == "cancel" {
			if timeId != -1 {
				insertRoomSQL := `INSERT INTO canceled (id, room) VALUES (?, ?)`
				_, err := db.Exec(insertRoomSQL, timeId, roomId)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

        // Store occupied room
		if timeId != -1 {
			insertRoomSQL := `INSERT INTO ocupiedRoom (id, room) VALUES (?, ?)`
			_, err := db.Exec(insertRoomSQL, timeId, roomId)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return rooms
}

// Get or create time slot ID
func getTimeID(time string, date string, endtime string) int {
	// Check if lesson is long enough (45 minutes minimum)
	minuteDifference, err := TimeDifferenceInMinutes(time, endtime)
	if err != nil {
		log.Fatal(err)
	}

	var timeID int
	datetime := fmt.Sprintf("%s %s", date, time)

	// Skip if lesson is too short
	if minuteDifference < 45 {
		return -1
	} else {
		// Try to get existing time ID
		err := db.QueryRow("SELECT id FROM times WHERE startTime = ?", datetime).Scan(&timeID)
		if err == sql.ErrNo