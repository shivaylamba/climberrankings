package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Gender .
type Gender string

// Athlete .
type Athlete struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Gender Gender `json:"gender"`
}

// Segment .
type Segment struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Distance           int64   `json:"distance"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	ElevationHigh      float64 `json:"elevation_high"`
	ElevationLow       float64 `json:"elevation_low"`
}

// LeaderboardEntry .
type LeaderboardEntry struct {
	ID       int64   `json:"id"`
	Segment  Segment `json:"segment"`
	Athlete  Athlete `json:"athlete"`
	Activity struct {
		ID int64 `json:"id"`
	} `json:"activity"`
	ElapsedTime int       `json:"elapsed_time"`
	StartDate   time.Time `json:"start_date"`
}

// DB .
type DB struct {
	*sql.DB
}

const athletesTable = `
	CREATE TABLE IF NOT EXISTS athletes (
		id INTEGER PRIMARY KEY,
		name TEXT,
		first_name TEXT,
		last_name TEXT,
		gender TEXT
	);`

const segmentsTable = `
	CREATE TABLE IF NOT EXISTS segments (
		id INTEGER PRIMARY KEY,
		name TEXT,
		distance INTEGER,
		total_elevation_gain REAL,
		elevation_low REAL,
		elevation_high REAL
	);`

const effortsTable = `
	CREATE TABLE IF NOT EXISTS efforts (
		id INTEGER PRIMARY KEY,
		segment_id INTEGER,
		athlete_id INTEGER,
		activity_id INTEGER,
		elapsed_time INTEGER,
		start_date DATETIME,
		FOREIGN KEY(segment_id) REFERENCES segments(id)
		FOREIGN KEY(athlete_id) REFERENCES athletes(id)
	);`

// Open .
func Open(databaseName string) (*DB, error) {
	db, err := sql.Open("sqlite3", databaseName)
	if err != nil {
		return nil, err
	}
	statement, err := db.Prepare(athletesTable)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}
	statement, err = db.Prepare(segmentsTable)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}
	statement, err = db.Prepare(effortsTable)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) writeAthlete(a *Athlete) error {
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	err = writeAthleteTx(tx, a)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func writeAthleteTx(tx *sql.Tx, a *Athlete) error {
	statement, err := tx.Prepare(`
		INSERT INTO athletes (id, name, gender)
		VALUES (?, ?, ?)`)
	defer statement.Close()
	if err != nil {
		return err
	}
	_, err = statement.Exec(a.ID, a.Name, a.Gender)
	return err
}

func (db *DB) readAthlete(athleteID int64) (*Athlete, error) {
	var a Athlete
	err := db.QueryRow(`
		SELECT (id, name, gender)
		FROM athletes
		WHERE id = ?`).Scan(
		&a.ID,
		&a.Name,
		&a.Gender)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) writeSegment(s *Segment) error {
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	err = writeSegmentTx(tx, s)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func writeSegmentTx(tx *sql.Tx, s *Segment) error {
	statement, err := tx.Prepare(`
		INSERT INTO segments (id, name, distance, total_elevation_gain, elevation_high, elevation_low)
		VALUES (?, ?, ?, ?, ?, ?)`)
	defer statement.Close()
	if err != nil {
		return err
	}
	_, err = statement.Exec(s.ID, s.Name, s.Distance, s.TotalElevationGain, s.ElevationHigh, s.ElevationLow)
	return err
}

func (db *DB) readSegment(segmentID int64) (*Segment, error) {
	var s Segment
	err := db.QueryRow(`
		SELECT (id, name, distance, total_elevation_gain, elevation_high, elevation_low)
		FROM segments
		WHERE id = ?`).Scan(
		&s.ID,
		&s.Name,
		&s.Distance,
		&s.TotalElevationGain,
		&s.ElevationHigh,
		&s.ElevationLow)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) writeEffort(e *Effort) error {
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	err = writeSegmentTx(tx, &e.Segment)
	if err != nil {
		return err
	}
	err = writeAthleteTx(tx, &e.Athlete)
	if err != nil {
		return err
	}
	statement, err := tx.Prepare(`
		INSERT INTO efforts (id, segment_id, athlete_id, activity_id, elapsed_time, start_date)
		VALUES (?, ?, ?, ?, ?, ?)`)
	defer statement.Close()
	if err != nil {
		return err
	}
	_, err = statement.Exec(e.ID, e.Segment.ID, e.Athlete.ID, e.Activity.ID, e.ElapsedTime, e.StartDate)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) readEffort(effortID int64) (*Effort, error) {
	var a Athlete
	var s Segment
	var e Effort
	err := db.QueryRow(`
		SELECT (
			id,
			activity_id,
			elapsed_time.
			start_date,
			athlete.id,
			athlete.name,
			athlete.gender,
			segment.id,
			segment.name,
			segment.distance,
			segment.total_elevation_gain,
			segment.elevation_high,
			segment.elevation_low)
		FROM efforts
		INNER JOIN athletes ON athletes.id = efforts.athlete_id
		INNER JOIN segments ON segments.id = efforts.segment_id
		WHERE id = ?`).Scan(
		&e.ID,
		&e.Activity.ID,
		&e.ElapsedTime,
		&e.StartDate,
		&a.ID,
		&a.Name,
		&a.Gender,
		&s.ID,
		&s.Name,
		&s.Distance,
		&s.TotalElevationGain,
		&s.ElevationHigh,
		&s.ElevationLow)
	if err != nil {
		return nil, err
	}
	e.Athlete = a
	e.Segment = s
	return &e, nil
}
