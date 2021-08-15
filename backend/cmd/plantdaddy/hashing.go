package main

// This file is used for the hashing function
import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// This function generates a random group of bytes from the os of the computer
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}





// This function hashes the login provided from the struct and 
//returns the result or the error and a message.
func hashBytes(login *Login, session *SessionData) (Session, error) {

	var b bytes.Buffer
	if login != nil {
		b.WriteString(login.DeviceID)
	} else if session != nil {
		b.WriteString(session.SessionID)
	}

	bytes, err := generateRandomBytes(32)
	// The random hash also would need to be written to a database and stored. 

	if err != nil {	
		return Session{}, err
	}
	// This function now adds the new bytes to the end of the string.
	b.Write(bytes)

	// now we hash all the bytes and return that 
	sha := sha256.New()

	_, errs := sha.Write(b.Bytes())

	if errs != nil {
		return Session{}, errs
	}
	
	 newSession := Session{
		SessionID: hex.EncodeToString(sha.Sum(nil)),
		UsageCounter: 10,
		Timestamp: time.Now(),
	 }

	return newSession, nil
	
	
}

