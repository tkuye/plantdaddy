package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)


func connectToDb(connString string) *sql.DB {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	
	return db
}


func insertDevice(login Login, username string, db *sql.DB) {
	
	row:= db.QueryRow("SELECT id FROM auth WHERE username=$1", username)

	var id int

	err := row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("no user with id %d\n", id)
	case err != nil:
		log.Fatalf("query error: %v\n", err)
	}
	
	_, errs := db.Query(`INSERT INTO registered_devices(device_id, user_id, register_date)
	VALUES($1, $2, $3)
	`, login.DeviceID, id, time.Now().UTC())

	if errs != nil {
		log.Fatal(errs)
	}
}

func getSession(db *sql.DB, login *Login) Session {

		hashedLogin, errs := hashBytes(login, nil)
		if errs != nil {
			log.Fatal(errs)
		}
		_, err := db.Query(`INSERT INTO session(session_id, usage_time, usage, device_id) VALUES ($1, $2, $3, $4) 
		ON CONFLICT (device_id) DO UPDATE SET session_id=$1, usage_time=$2, usage=$3`, 
		hashedLogin.SessionID, hashedLogin.Timestamp, hashedLogin.UsageCounter, login.DeviceID)

		if err != nil {
			log.Fatal(err)
		}

		return hashedLogin
}

func insertSessionData(sessionData SessionData, db *sql.DB) Session {
	_, errs := db.Query(`
	INSERT INTO plant_data(device_id, time, temperature, humidity, soil_moisture, light)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, sessionData.DeviceID, time.Now().UTC(), 
	sessionData.Temperature, 
	sessionData.Humidity, 
	sessionData.SoilMoisture, 
	sessionData.Light)

	if errs != nil {
		log.Fatal(errs)
	}

	row := db.QueryRow("SELECT session_id, usage, usage_time FROM session WHERE session_id=$1", sessionData.SessionID)
	var session Session

	err := row.Scan(&session.SessionID, &session.UsageCounter, &session.Timestamp)
	if err == sql.ErrNoRows || sessionData.UsageCounter == 0 || session.UsageCounter != sessionData.UsageCounter {
		new_session, errs := hashBytes(nil, &sessionData)

		if errs != nil {
			log.Fatal(errs)
		}

		db.Query("UPDATE session SET session_id=$1, usage_time=$2, usage=$3 WHERE device_id=$4", 
		new_session.SessionID, new_session.Timestamp, new_session.UsageCounter, sessionData.DeviceID)
		println(new_session.SessionID)
		return new_session
	} else if err != nil {
		log.Fatal(err)
		return Session{}
	} else {
		newSession := Session{
			SessionID: sessionData.SessionID,
			UsageCounter: sessionData.UsageCounter - 1,
			Timestamp: time.Now().UTC(),
		}

		
		db.Query("UPDATE session SET usage_time=$1, usage=$2 WHERE device_id=$3", 
		newSession.Timestamp, newSession.UsageCounter, sessionData.DeviceID)
		return newSession

	}

}
