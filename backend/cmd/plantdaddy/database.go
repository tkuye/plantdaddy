package main

import (
	"context"
	"database/sql"
	"log"
	"time"
	"github.com/jackc/pgx/v4/pgxpool"
	
	
	_ "github.com/lib/pq"
)

type FalseError struct {}

func (e *FalseError) Error() string {
	return "The passwords do not match."
}
// Conenects to the database with a given string
func connectToDb(connString string) *pgxpool.Pool {
	db, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return db
}

// DB Query to connect to database and change the device string
func changeDeviceName(db *pgxpool.Pool, devName deviceName) error {
	_, err := db.Exec(context.Background(), "UPDATE registered_devices SET device_name = $1 WHERE device_id=$2" ,devName.DeviceName, devName.DeviceID)
	
	if err != nil {
		return err
	}
	
	return nil
}

// DB Query to connect to database and and get the device data
func getDeviceDB(db *pgxpool.Pool, deviceID string) (Device, error) {

	var device Device
	device.DeviceID = deviceID
	dataRow := db.QueryRow(context.Background(),`SELECT p.temperature, 
		p.humidity, p.soil_moisture, p.light, p.time, r.device_name 
		FROM plant_data p INNER JOIN registered_devices r
		ON p.device_id = r.device_id
		WHERE p.device_id=$1 ORDER BY time DESC LIMIT 1
		`, deviceID)
	
	err := dataRow.Scan(&device.DeviceData.Temperature,
		 &device.DeviceData.Humidity, 
		 &device.DeviceData.SoilMoisture,
		&device.DeviceData.Light,
		&device.DeviceData.Timestamp, 
		&device.DeviceName,
	)
	
	if err != nil {
		log.Printf("%s", err)
		return Device{}, err
	}

	return device, nil
}

// DB Query to connect to database and delete the device data.
func deleteDeviceDB(db *pgxpool.Pool, deviceID string) error {
	_, err := db.Exec(context.Background(), "DELETE FROM registered_devices WHERE device_id=$1", deviceID)

	if err != nil {
		return err
	}
	return nil
}

// DB Query to connect to database and get the latest data.
func getLatestDataDay(db *pgxpool.Pool, date string, deviceId string) (map[int]DeviceHourData, error) {
	rows, err := db.Query(context.Background(), `
	SELECT temperature, humidity, soil_moisture, light, (time::timestamp AT TIME ZONE 'GMT') FROM plant_data 
	WHERE (time::timestamp AT TIME ZONE 'GMT')::date = $1 AND device_id = $2
	`, date, deviceId)
	dayMap := make(map[int]DeviceHourData)
	if err != nil {
		log.Printf("error %s", err)
		return nil, err
	}

	defer rows.Close()

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

// DB Query to connect to database and get all the latest devices associated with an ID 
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

// DB Query to connect to database and and insert a new device.
func insertDevice(newDevice *NewDevice, db *pgxpool.Pool) error{
	log.Printf("DEVICE: %s %s", newDevice.DeviceName, newDevice.DeviceID)
	row:= db.QueryRow(context.Background(), "SELECT id FROM auth WHERE LOWER(username)=LOWER($1)", newDevice.Username)

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
	
	_, errs := db.Exec(context.Background(),`INSERT INTO registered_devices(device_id, user_id, register_date, device_name)
	VALUES($1, $2, $3, $4)
	`, newDevice.DeviceID, id, time.Now().UTC(), newDevice.DeviceName)

	if errs != nil {
		log.Printf("%s", errs)
		return errs
	}
	return nil
}

// DB Query to connect to database and log in as the device with a given username and password.
func LogIn(db *pgxpool.Pool, user UserPass) error {
	row := db.QueryRow(context.Background(), "SELECT password FROM auth WHERE LOWER(username)=LOWER($1) OR LOWER(email)=LOWER($1)", user.Username)
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
	
	if !checker {
		return &FalseError{}
	}

	return nil
}



// DB Query to connect to database and get a new device session for the device
func getSession(db *pgxpool.Pool, login *Login) (Session, error) {

		hashedLogin, errs := hashBytes(login, nil)
		
		if errs != nil {
			return Session{}, errs
		}
		_, err := db.Exec(context.Background(), `INSERT INTO session(session_id, usage_time, usage, device_id) VALUES ($1, $2, $3, $4) 
		ON CONFLICT (device_id) DO UPDATE SET session_id=$1, usage_time=$2, usage=$3`, 
		hashedLogin.SessionID, hashedLogin.Timestamp, hashedLogin.UsageCounter, login.DeviceID)

		if err != nil {
			return Session{}, err
		}

		return hashedLogin, nil
}

// DB Query to connect to database and create a new user that can be used to create devices.
func insertNewUser(newUser UserPass , db *pgxpool.Pool) error {
	password, errs := HashPassword(newUser.Password)
	
	if errs != nil {
		log.Printf("%s", errs)
		return errs 
	}
	_, err := db.Exec(context.Background(), `INSERT INTO auth(username, password, email) VALUES ($1, $2, $3)`, newUser.Username, password, newUser.Email)

	if err != nil {
		log.Printf("%s", err)
		return err
	}
	return nil
}


// DB Query that will create a take a given session and insert the plant data associated with it in the database.
func insertSessionData(sessionData SessionData, db *pgxpool.Pool) (Session, error) {
	
	_, errs := db.Exec(context.Background(),`
	INSERT INTO plant_data(device_id, time, temperature, humidity, soil_moisture, light)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, sessionData.DeviceID, time.Now().UTC(), 
	sessionData.Temperature, 
	sessionData.Humidity, 
	sessionData.SoilMoisture, 
	sessionData.Light)
	if errs == sql.ErrNoRows {
		return Session{
			SessionID: "",
			UsageCounter: 0,
			Timestamp: time.Now(),
		}, errs
	}

	
	if errs != nil {
		log.Printf("%s", errs)
		return Session{
			SessionID: "",
			UsageCounter: 0,
			Timestamp: time.Now(),
		}, errs
	}

	row := db.QueryRow(context.Background(),"SELECT session_id, usage, usage_time FROM session WHERE session_id=$1", sessionData.SessionID)
	var session Session

	err := row.Scan(&session.SessionID, &session.UsageCounter, &session.Timestamp)
	if err == sql.ErrNoRows || sessionData.UsageCounter == 0 || session.UsageCounter != sessionData.UsageCounter {
		new_session, errs := hashBytes(nil, &sessionData)

		if errs != nil {
			log.Printf("%s", errs)
			return Session{}, errs
		}

		_, err = db.Exec(context.Background(),"UPDATE session SET session_id=$1, usage_time=$2, usage=$3 WHERE device_id=$4", 
		new_session.SessionID, new_session.Timestamp, new_session.UsageCounter, sessionData.DeviceID)

		if err != nil {
			return Session{}, errs
		}

		return new_session, nil
	} else if err != nil {
		log.Printf("%s", errs)
		return Session{}, errs
	} else {
		newSession := Session{
			SessionID: sessionData.SessionID,
			UsageCounter: sessionData.UsageCounter - 1,
			Timestamp: time.Now().UTC(),
		}

		_, err = db.Exec(context.Background(),"UPDATE session SET usage_time=$1, usage=$2 WHERE device_id=$3", 
		newSession.Timestamp, newSession.UsageCounter, sessionData.DeviceID)
		if err != nil {
			return Session{}, err
		}
		return newSession, nil
	}

}
