package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/jackc/pgx/v4/pgxpool"
	
	
	_ "github.com/lib/pq"
)

type FalseError struct {}

func (e *FalseError) Error() string {
	return "The passwords do not match."
}

func connectToDb(connString string) *pgxpool.Pool {
	db, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	
	return db
}


func changeDeviceName(db *pgxpool.Pool, devName deviceName) error {
	_, err := db.Query(context.Background(), "UPDATE registered_devices SET device_name = $1 WHERE device_id=$2" ,devName.DeviceName, devName.DeviceID)
	
	if err != nil {
		return err
	}
	
	return nil
}


func deleteDeviceDB(db *pgxpool.Pool, deviceID string) error {
	_, err := db.Query(context.Background(), "DELETE FROM registered_devices WHERE device_id=$1", deviceID)

	if err != nil {
		return err
	}
	return nil
}

func getLatestDataDay(db *pgxpool.Pool, date string, deviceId string) (map[int]DeviceHourData, error) {
	rows, err := db.Query(context.Background(), `
	SELECT temperature, humidity, soil_moisture, light, time FROM plant_data 
	WHERE time::date = $1 AND device_id = $2
	`, date, deviceId)
	dayMap := make(map[int]DeviceHourData)
	if err != nil {
		log.Printf("error %s", err)
		return nil, err
	}
	for rows.Next() {
	var time time.Time
	var deviceData DeviceHourData
	deviceData.DeviceNumber = 1
	errs := rows.Scan(&deviceData.Temperature, &deviceData.Humidity, &deviceData.SoilMoisture, &deviceData.Light, &time)

	if errs != nil {
		log.Printf("%s", errs)
		return nil, errs
	}
	if val, ok := dayMap[time.Hour()]; ok {
		val.Humidity += deviceData.Humidity
		val.Temperature += deviceData.Temperature
		val.SoilMoisture += deviceData.SoilMoisture
		val.Light += deviceData.Light
		val.DeviceNumber += 1
		dayMap[time.Hour()] = val
	} else {
		deviceData.TimePeriod = time.Hour()
		dayMap[time.Hour()] = deviceData
	}
	
	}
		// dayMap is a map[int]interface.
		// loop over keys and values in the map.
		for k, v := range dayMap {
			v.Humidity /= float64(v.DeviceNumber)
			v.Temperature /= float64(v.DeviceNumber)
			v.Light /= float64(v.DeviceNumber)
			v.SoilMoisture /= float64(v.DeviceNumber)
			dayMap[k] = v
		}

	return dayMap, nil
}

func getDevicesDB(db *pgxpool.Pool, username string) ([]Device, error) {

	var id int64
	row := db.QueryRow(context.Background(), `SELECT id from auth WHERE username = $1`, username)
	
	err := row.Scan(&id)
	if err != nil {
		return nil, err
	}

	
	

	rows, errs := db.Query(context.Background(),`SELECT device_name, device_id FROM registered_devices WHERE user_id = $1`, id)

	if errs != nil {
		return nil, errs
	}
	
	var devices []Device

	defer rows.Close()
	for rows.Next() {
		var device Device
		err := rows.Scan(&device.DeviceName, &device.DeviceID)
		if err != nil {
			return nil, err
		}

		dataRow := db.QueryRow(context.Background(),`SELECT temperature, 
		humidity, soil_moisture, light, time FROM plant_data 
		WHERE device_id=$1 ORDER BY time DESC LIMIT 1
		`, device.DeviceID)

		errs := dataRow.Scan(&device.DeviceData.Temperature, &device.DeviceData.Humidity, 
		&device.DeviceData.SoilMoisture, &device.DeviceData.Light,&device.DeviceData.Timestamp,
		)
		if errs == sql.ErrNoRows {
			log.Printf("%s", errs)
		}
		devices = append(devices, device)

	}

	return devices, nil

}


func insertDevice(newDevice *NewDevice, db *pgxpool.Pool) error{
	
	row:= db.QueryRow(context.Background(), "SELECT id FROM auth WHERE username=$1", newDevice.Username)

	var id int

	err := row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("no user with id %d\n", id)
		return err
	case err != nil:
		log.Printf("query error: %v\n", err)
		return err
	}
	
	_, errs := db.Query(context.Background(),`INSERT INTO registered_devices(device_id, user_id, register_date, device_name)
	VALUES($1, $2, $3, $4)
	`, newDevice.DeviceID, id, time.Now().UTC(), newDevice.DeviceName)

	if errs != nil {
		log.Printf("%s", errs)
		return errs
	}
	return nil
}


func LogIn(db *pgxpool.Pool, user UserPass) error {
	row := db.QueryRow(context.Background(), "SELECT password FROM auth WHERE username=$1", user.Username)
	var password string
	err := row.Scan(&password)
	log.Printf("%s", db.Stat().AcquireDuration().String())
	switch {
		case err == sql.ErrNoRows:
			return err
		case err != nil:
			return err
	}

	var checker = CheckPasswordHash(user.Password, password)
	log.Println(checker)
	if !checker {
		return &FalseError{}
	}

	return nil
}


func getSession(db *pgxpool.Pool, login *Login) (Session, error) {

		hashedLogin, errs := hashBytes(login, nil)
		
		if errs != nil {
			return Session{}, errs
		}
		_, err := db.Query(context.Background(), `INSERT INTO session(session_id, usage_time, usage, device_id) VALUES ($1, $2, $3, $4) 
		ON CONFLICT (device_id) DO UPDATE SET session_id=$1, usage_time=$2, usage=$3`, 
		hashedLogin.SessionID, hashedLogin.Timestamp, hashedLogin.UsageCounter, login.DeviceID)

		if err != nil {
			return Session{}, err
		}

		return hashedLogin, nil
}

func insertNewUser(newUser UserPass , db *pgxpool.Pool) error {
	password, errs := HashPassword(newUser.Password)
	
	if errs != nil {
		log.Printf("%s", errs)
		return errs 
	}
	_, err := db.Query(context.Background(), `INSERT INTO auth(username, password) VALUES ($1, $2)`, newUser.Username, password)

	if err != nil {
		log.Printf("%s", err)
		return err
	}
	return nil
}

func insertSessionData(sessionData SessionData, db *pgxpool.Pool) (Session, error) {
	
	_, errs := db.Query(context.Background(),`
	
	INSERT INTO plant_data(device_id, time, temperature, humidity, soil_moisture, light)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, sessionData.DeviceID, time.Now().UTC(), 
	sessionData.Temperature, 
	sessionData.Humidity, 
	sessionData.SoilMoisture, 
	sessionData.Light)

	if errs != nil {
		return Session{}, errs
	}

	row := db.QueryRow(context.Background(),"SELECT session_id, usage, usage_time FROM session WHERE session_id=$1", sessionData.SessionID)
	var session Session

	err := row.Scan(&session.SessionID, &session.UsageCounter, &session.Timestamp)
	if err == sql.ErrNoRows || sessionData.UsageCounter == 0 || session.UsageCounter != sessionData.UsageCounter {
		new_session, errs := hashBytes(nil, &sessionData)

		if errs != nil {
			return Session{}, errs
		}

		db.Query(context.Background(),"UPDATE session SET session_id=$1, usage_time=$2, usage=$3 WHERE device_id=$4", 
		new_session.SessionID, new_session.Timestamp, new_session.UsageCounter, sessionData.DeviceID)
		return new_session, nil
	} else if err != nil {
		
		return Session{}, errs
	} else {
		newSession := Session{
			SessionID: sessionData.SessionID,
			UsageCounter: sessionData.UsageCounter - 1,
			Timestamp: time.Now().UTC(),
		}

		db.Query(context.Background(),"UPDATE session SET usage_time=$1, usage=$2 WHERE device_id=$3", 
		newSession.Timestamp, newSession.UsageCounter, sessionData.DeviceID)
		return newSession, nil
	}

}
