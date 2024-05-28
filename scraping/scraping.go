package scraping

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type ListStasiun struct {
	id    string
	title string
}

type StasiunSchedule struct {
	StasiunID   string
	StasiunName string
	Arah        string
	Schedule    string
}

func RunScraping() error {
	//script logic
	var listStasiun []ListStasiun
	var allStasiunSchedules []StasiunSchedule

	// creating colly instance
	c := colly.NewCollector()

	//visit target page
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL)
	})
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Error:", err)
	})
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Page visited:", r.Request.URL)
	})

	// Extract station options
	c.OnHTML("select#fareFrom option", func(e *colly.HTMLElement) {
		// Extract the text content of the <option> element
		stationID := e.Attr("value")
		stationName := e.Text

		fmt.Printf("Found station: ID=%s, name=%s\n", stationID, stationName)

		// Initialize a new ListStasiun instance
		stasiun := ListStasiun{
			id:    stationID,
			title: stationName,
		}

		// Add the station to the list
		listStasiun = append(listStasiun, stasiun)
	})

	//visit the target URL
	c.OnHTML(".row-jadwal", func(e *colly.HTMLElement) {
		// Extract station name
		stasiunName := strings.TrimSpace(e.Attr("data-stasiun"))

		//extract the stasiun_id from the class name using regex
		classNames := e.Attr("class")
		re := regexp.MustCompile(`row-(\d+)`)
		match := re.FindStringSubmatch(classNames)
		stasiunID := ""
		if len(match) > 1 {
			stasiunID = match[1]
		}

		// Extract the arah and schedules for the current station
		e.ForEach("div.col-12.col-xl-6", func(_ int, s *colly.HTMLElement) {
			// Get the direction
			arah := strings.TrimSpace(s.DOM.Find("h3").Text())

			// Get the list of schedules
			schedules := make([]string, 0)
			s.DOM.Find("ul#schedule-b span").Each(func(_ int, schedule *goquery.Selection) {
				scheduleText := strings.TrimSpace(schedule.Text())
				schedules = append(schedules, scheduleText)
			})

			// Print out schedules for both directions
			for _, schedule := range schedules {
				allStasiunSchedules = append(allStasiunSchedules, StasiunSchedule{
					StasiunID:   stasiunID,
					StasiunName: stasiunName,
					Arah:        arah,
					Schedule:    schedule,
				})
			}
		})
	})

	// Call createCSV function after scraping is completed
	c.OnScraped(func(r *colly.Response) {

		if err := CreateCSV(listStasiun); err != nil {
			log.Printf("Error creating list stasiun : %v", err)
		}
		if err := WriteCSV(allStasiunSchedules); err != nil {
			log.Printf("Error creating stasiun schedules: %v", err)
		}
	})

	// Start scraping
	fmt.Println("Starting scraping...")
	c.Visit("https://jakartamrt.co.id/id/jadwal-keberangkatan-mrt?dari=null")
	return nil
}

func CreateCSV(listStasiun []ListStasiun) error {
	dir := "data"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}
	fmt.Println("Creating CSV...")
	// create the  CSV file
	filePath := filepath.Join(dir, "listStasiun.csv")
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Failed to create output CSV file:", err)
		return err
	}
	defer file.Close()
	fmt.Println("CSV file created successfully.")

	// Initialize the CSV writer
	writer := csv.NewWriter(file)

	// Write CSV headers
	headers := []string{"id", "stasiun"}
	if err := writer.Write(headers); err != nil {
		log.Println("Failed to write CSV headers:", err)
		return err
	}
	fmt.Println("CSV headers written successfully.")

	// Write each station as a CSV row
	for _, stasiun := range listStasiun {
		record := []string{stasiun.id, stasiun.title}
		if err := writer.Write(record); err != nil {
			log.Println("Failed to write CSV record:", err)
			return err
		}
		// Flush after writing each record
		writer.Flush()
	}
	fmt.Println("CSV file writing completed successfully.")
	return nil
}

func WriteCSV(data []StasiunSchedule) error {
	dir := "data"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}
	// create csv file
	fmt.Println("Writing CSV...")
	filepath := filepath.Join(dir, "stasiunSchedules.csv")
	file, err := os.Create(filepath)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"StasiunID", "StasiunName", "Arah", "Schedule"})

	// Write schedule data
	for _, schedule := range data {
		err := writer.Write([]string{schedule.StasiunID, schedule.StasiunName, schedule.Arah, schedule.Schedule})
		if err != nil {
			log.Fatalf("Error writing CSV record: %v", err)
		}
	}

	fmt.Println("CSV file has been written successfully.")
	return nil
}

func CleanupOldCSVFiles(directory string, maxAge time.Duration) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	now := time.Now()

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			filePath := filepath.Join(directory, file.Name())
			info, err := os.Stat(filePath)
			if err != nil {
				log.Printf("Error starting file %s:%v", filePath, err)
			}
			if now.Sub(info.ModTime()) > maxAge {
				log.Printf("Deletingold CSV  file: %s", filePath)
				err := os.Remove(filePath)
				if err != nil {
					log.Printf("Errordeleting file %s:%v", filePath, err)
				}
			}
		}
	}
	return nil
}

// CreateTable creates the necessary tables if they don't already exist.
func CreateTable(db *sql.DB) error {
	createStmt := `
	CREATE TABLE IF NOT EXISTS stations (
		id serial PRIMARY KEY,
		stasiun_name VARCHAR(255) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS schedules (
		id SERIAL PRIMARY KEY,
		station_id INT NOT NULL,
		stasiun_name VARCHAR(255),
		arah VARCHAR(255) NOT NULL,
		jadwal TIME
	);
	`
	_, err := db.Exec(createStmt)
	if err != nil {
		log.Fatalf("Error creating table: %v\n", err)
		return err
	}
	fmt.Println("Tables created successfully!")
	return nil
}

func InsertData(db *sql.DB) error {
	if err := InsertStations(db, "data/listStasiun.csv"); err != nil {
		return err
	}
	if err := InsertSchedules(db, "data/stasiunSchedules.csv"); err != nil {
		return err
	}
	return nil
}

func InsertStations(db *sql.DB, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %b\n", err)
		return err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("Error reading header row from CSV file : %v\n", err)
	}
	for {
		record, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading record from CSV: %v\n", err)
			return err
		}
		id, _ := strconv.Atoi(record[0])
		stasiun := record[1]

		_, err = db.Exec("INSERT INTO  stations (id, stasiun_name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING", id, stasiun)
		if err != nil {
			log.Fatalf("Error inserting data into stations table : %v\n", err)
			return err
		}
	}
	return nil
}

func InsertSchedules(db *sql.DB, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v\n", err)
		return err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	_, err = reader.Read() // Skip the header row
	if err != nil {
		log.Fatalf("Error reading header row from CSV file: %v\n", err)
		return err
	}
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading record from CSV : %v", err)
			return err
		}

		stasiunID, _ := strconv.Atoi(record[0])
		stasiun := record[1]
		arah := record[2]
		jadwal := record[3]

		// Parse the time format
		formattedTime, err := ParseTime(jadwal)
		if err != nil {
			log.Fatalf("Error parsing time value: %v", err)
			return err
		}

		_, err = db.Exec("INSERT INTO schedules (station_id, stasiun_name, arah, jadwal) VALUES ($1, $2, $3, $4)",
			stasiunID, stasiun, arah, formattedTime)
		if err != nil {
			log.Fatalf("Error inserting data into schedules table: %v\n", err)
			return err
		}
	}

	return nil
}

func RemoveSchedules(db *sql.DB) error {
	// Execute the delete query
	_, err := db.Exec("DELETE FROM schedules")
	if err != nil {
		log.Fatalf("Error deleting records from schedules table: %v", err)
		return err
	}
	log.Println("All records deleted from schedules table successfully.")
	return nil
}

func RemoveStations(db *sql.DB) error {
	// execute the delete query
	_, err := db.Exec("DELETE FROM stations")
	if err != nil {
		log.Fatalf("Error deleting records from stations table: %v", err)
		return err
	}
	log.Println("All record deleted from stations table successfully.")
	return nil
}
func ParseTime(t string) (string, error) {
	parsedTime, err := time.Parse("15:04", t)
	if err != nil {
		return "", fmt.Errorf("error parsing time value : %v", err)
	}
	return parsedTime.Format("15:04:00"), nil
}
