package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// helper function
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

//helper function
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

//helper function
func TimeDifferenceInMinutes(time1, time2 string) (int, error) {
	// Define the layout of the time format
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

//helper funciton
func openDB() {
	var err error
	db, err = sql.Open("sqlite3", "./brake-room.db")
	if err != nil {
		log.Fatal(err)
	}

	// Optionally set database connection settings
	db.SetMaxOpenConns(5) // For SQLite, usually one connection is enough
	db.SetMaxIdleConns(1)

	_, err = db.Exec("DELETE FROM ocupiedRoom")
	fmt.Println("Delted ocupiedRoom")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM rooms")
	fmt.Println("Delted rooms")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM times")
	fmt.Println("Delted times")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM canceled")
	fmt.Println("Delted canceled")
	if err != nil {
		log.Fatal(err)
	}
}

//get sturm session for further request
func sendLoginRequest() []byte {
	// Define the curl command as a string
	curlCmd := `curl --path-as-is -i -s -k -X $'POST' \
    -H $'Host: intranet.tam.ch' -H $'Content-Length: 97' -H $'Cache-Control: max-age=0' -H $'Sec-Ch-Ua: \"Chromium\";v=\"127\", \"Not)A;Brand\";v=\"99\"' -H $'Sec-Ch-Ua-Mobile: ?0' -H $'Sec-Ch-Ua-Platform: \"macOS\"' -H $'Accept-Language: en-US' -H $'Upgrade-Insecure-Requests: 1' -H $'Origin: https://intranet.tam.ch' -H $'Content-Type: application/x-www-form-urlencoded' -H $'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.6533.100 Safari/537.36' -H $'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' -H $'Sec-Fetch-Site: same-origin' -H $'Sec-Fetch-Mode: navigate' -H $'Sec-Fetch-User: ?1' -H $'Sec-Fetch-Dest: document' -H $'Referer: https://intranet.tam.ch/kzu' -H $'Accept-Encoding: gzip, deflate, br' -H $'Priority: u=0, i' \
    -b $'school=kzu; username=studentName; sturmuser=StudentName; sturmsession=87f6ff4811f8175ec2da9d08e97fdad8' \
    --data-binary $'hash=66c02d6f5436e6.99763438&loginschool=kzu&loginuser=studentName&loginpassword=studentPassword' \
    $'https://intranet.tam.ch/kzu'`

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

func getreqeststring(byteInput []byte) []string {
	outputStr := string(byteInput)
	lines := strings.Split(outputStr, "\n")

	return lines
}

//extract stumsession from return
func extractSturmsession(lines []string) string {
	var sturmSession string

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

//get lesson plan data with before gotten sturm session
func sendDataRequest(sturmSession string) []byte {
	start := time.Now()
	// Find the last Monday
	offset := (int(time.Monday) - int(start.Weekday()) + 7) % 7
	lastMonday := start.AddDate(0, 0, -offset).Truncate(24 * time.Hour).Add(3 * time.Hour) // Set to last Monday at 3 AM
	startDate := lastMonday.Unix() * 1000
	endDate := startDate + (432000 * 1000)
	fmt.Println(startDate)
	fmt.Println(endDate)
	// Define the curl command as a string
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

//format data from server as lsit of rooms
func getRoomsList(stringInput string) []string {
	// Unmarshal into a map
	var result map[string]interface{}
	err := json.Unmarshal([]byte(stringInput), &result)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil
	}

	// Access the "data" field, which is a slice of interfaces
	data, ok := result["data"].([]interface{})
	if !ok {
		fmt.Println("Error: 'data' field is not in the expected format")
		return nil
	}

	var rooms []string

	openDB()

	// Loop over each element in the "data" slice
	for _, item := range data {
		// Assert the type of each item to map[string]interface{}
		entry, ok := item.(map[string]interface{})
		if !ok {
			fmt.Println("Error: item is not in the expected format")
			continue
		}

		// Access the "roomId" field, which is a slice of interfaces
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

		// Append the roomId to the rooms slice as a string
		timeId := getTimeID(startTime, startDate, endtime)

		if canceled == "cancel" {
			if timeId != -1 {
				insertRoomSQL := `INSERT INTO canceled (id, room) VALUES (?, ?)`
				_, err := db.Exec(insertRoomSQL, timeId, roomId)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

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

//get time id to get current weeks data
func getTimeID(time string, date string, endtime string) int {
	minuteDifference, err := TimeDifferenceInMinutes(time, endtime)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the row with the given time exists in the times table
	var timeID int
	datetime := fmt.Sprintf("%s %s", date, time)

	if minuteDifference < 45 {
		return -1
	} else {
		err := db.QueryRow("SELECT id FROM times WHERE startTime = ?", datetime).Scan(&timeID)
		if err == sql.ErrNoRows {
			// Row does not exist, so insert a new row with the given time
			_, err := db.Exec("INSERT INTO times (startTime) VALUES (?)", datetime)
			if err != nil {
				log.Fatal(err)
			}

			err = db.QueryRow("SELECT id FROM times WHERE startTime = ?", datetime).Scan(&timeID)
			if err != nil {
				log.Fatal(err)
			}

		} else if err != nil {
			log.Fatal(err)
		}

		return timeID
	}
}

//helper function
func extractBody(byteInput []byte) string {
	outputStr := string(byteInput)

	// Find the start of the body
	bodyStartIndex := strings.Index(outputStr, "\r\n\r\n")
	if bodyStartIndex == -1 {
		bodyStartIndex = strings.Index(outputStr, "\n\n") // Fallback in case \r\n\r\n is not used
	}
	if bodyStartIndex != -1 {
		// Extract the body from the response
		body := outputStr[bodyStartIndex+4:] // +4 to skip the \r\n\r\n or \n\n
		return body
	} else {
		fmt.Println("Could not find the response body")
	}

	return ""
}

//compare ocupied rooms with room list
func insertEmptyRooms() {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // Commit transaction
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	times, err := tx.Query("SELECT id FROM times")
	if err != nil {
		log.Fatal(err)
	}
	defer times.Close()

	allRooms := []string{
		"Z101", "Z102", "Z103", "Z201", "Z204", "Z205", "Z206", "Z207", "Z209",
		"Z210", "Z211", "Z212", "Z214", "Z215", "Z216", "Z217", "Z218", "Z222", "Z223",
		"Z224", "Z226", "Z303", "Z304", "Z305", "Z306", "Z307", "Z309",
		"Z310", "Z311", "Z314", "Z315", "Z316", "Z317", "Z320",
		"Z404", "Z405", "Z406", "Z407", "Z408", "Z409", "Z412", "Z413", "Z414", "Z415",
		"Z416", "Z417", "Z418", "Z421", "Z422", "A", "B", "C", "D",
	}

	for times.Next() {
		var id int
		err := times.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}

		// Fetch the occupied rooms for the given id
		rooms, err := tx.Query("SELECT room FROM ocupiedRoom WHERE id = ?", id)
		if err != nil {
			log.Fatal(err)
		}

		roomSlice := []string{}
		for rooms.Next() {
			var room string
			err := rooms.Scan(&room)
			if err != nil {
				log.Fatal(err)
			}
			roomSlice = append(roomSlice, room)
		}

		rooms.Close() // Close the rooms query to free up resources

		for _, room := range allRooms {
			if !contains(roomSlice, room) {
				insertRoomSQL := `INSERT INTO rooms (id, room) VALUES (?, ?)`
				_, err := tx.Exec(insertRoomSQL, id, room)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	if err := times.Err(); err != nil {
		log.Fatal(err)
	}
}

//mark consecutive Rooms as consecutive
func consecutiveRooms() {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // Commit transaction
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	times, err := tx.Query("select id from times")
	if err != nil {
		log.Fatal(err)
	}
	defer times.Close()

	var ids []int
	for times.Next() {
		var id int
		if err := times.Scan(&id); err != nil {
			log.Fatal(err)
		}
		ids = append(ids, id)
	}

	if err = times.Err(); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(ids)-1; i++ {
		id := ids[i]

		futureRoom, err := tx.Query("select room from rooms where id = ?", id+1)
		if err != nil {
			log.Fatal(err)
		}
		defer futureRoom.Close()

		futureRoomsSlice := []string{}

		for futureRoom.Next() {
			var room string
			if err := futureRoom.Scan(&room); err != nil {
				log.Fatal(err)
			}

			futureRoomsSlice = append(futureRoomsSlice, room)
		}

		if err = futureRoom.Err(); err != nil {
			log.Fatal(err)
		}

		room, err := tx.Query("select room from rooms where id = ?", id)
		if err != nil {
			log.Fatal(err)
		}
		defer room.Close()

		for room.Next() {
			var currentRoom string
			if err := room.Scan(&currentRoom); err != nil {
				log.Fatal(err)
			}

			fmt.Println(currentRoom)
			fmt.Println(futureRoomsSlice)

			if contains(futureRoomsSlice, currentRoom) {
				_, err := tx.Exec("UPDATE rooms SET consecutive = true WHERE id = ? AND room = ?", id, currentRoom)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		if err = room.Err(); err != nil {
			log.Fatal(err)
		}
	}
}

//mark all canceled rooms as canceled
func canceledRooms() {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // Commit transaction
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	rows, err := tx.Query("SELECT * FROM canceled")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var room string
		err := rows.Scan(&id, &room)
		if err != nil {
			log.Fatal(err)
		}

		_, err = tx.Exec("INSERT INTO rooms (id, room, canceled) VALUES (?, ?, true)", id, room)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func fullLoginRequest() string {
	output := sendLoginRequest()//first login request
	lines := getreqeststring(output)//extract string
	sturmSession := extractSturmsession(lines)//extract sturmsession
	return sturmSession
}

func getRoomData(sturmSession string) {
	output := sendDataRequest(sturmSession)//send room plan request
	body := extractBody(output)//extract rooms
	getRoomsList(body)//format as lsit
	insertEmptyRooms()//insert empty rooms into db
	consecutiveRooms()//insert consecutive rooms into db
	canceledRooms()//insert canceled rooms into db
}

func main() {
	sturmSession := fullLoginRequest()//simulate login request

	getRoomData(sturmSession)//get and log data
}
