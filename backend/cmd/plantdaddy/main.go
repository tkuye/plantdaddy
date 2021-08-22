package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"time"
	
)

// Requires initial server request config for login
type API struct {
	db * pgxpool.Pool
}



func main() {
	// Initial web server configuration
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "http: ", log.Flags())

	
	api := &API{db: connectToDb(os.Getenv("CONNSTRING"))}
	http.HandleFunc("/auth-device", api.logIn)
	http.HandleFunc("/api/new-device", api.newDevice)
	http.HandleFunc("/api/login", api.logInApp)
	http.HandleFunc("/new-data",api.newSessionData)
	http.HandleFunc("/api/new-user", api.newUser)
	http.HandleFunc("/api/devices", api.getDevices)
	http.HandleFunc("/api/get-daily-data", api.getDailyData)
	http.HandleFunc("/api/device-name", api.changeDeviceName)
	http.HandleFunc("/api/delete-device", api.deleteDevice)
	log.Println("Listening for requests at http://localhost:8000/")
	server := &http.Server{
		ReadTimeout: 5 * time.Second,
    	WriteTimeout:5 * time.Second,
		Addr:           ":8000",
		IdleTimeout:  5 * time.Second,
		ErrorLog: logger,
	}
	log.Fatal(server.ListenAndServe())
}


func (api * API) deleteDevice(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	
	if r.Method == "DELETE" {
		deviceID := r.URL.Query().Get("deviceID")

		deleteDeviceDB(api.db, deviceID)
	}
}

func (api * API) changeDeviceName(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	if r.Method == "POST" {
		jsonEncoder := json.NewDecoder(r.Body)
		jsonEncoder.DisallowUnknownFields()
		var name deviceName
		
		err := jsonEncoder.Decode(&name)

		if jsonDecoder(err, w) != nil {
			return
		}

		errs := changeDeviceName(api.db, name)

		if errs != nil {
			http.Error(w, "Error updating device name", http.StatusInternalServerError)
		}

	}
}	
	
func (api * API) getDevices(w http.ResponseWriter, r *http.Request){

	if r.Method == "GET" {
		log.Printf("New request: %s", r.URL)
		
		username := r.URL.Query().Get("username")
		if username != "" {
			devices, err := getDevicesDB(api.db, username)

			if err != nil {
				log.Printf("%s", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			jsonEncoder := json.NewEncoder(w)
			jsonEncoder.Encode(devices)
			defer r.Body.Close()
		}

		
	}
}




func (api * API) newDevice(w http.ResponseWriter, r *http.Request) {
	var newDevice NewDevice;

	if r.Method == "POST" {

		log.Printf("New Request %s", r.URL)

	if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			
		}

		
		decoder := json.NewDecoder(r.Body)
		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&newDevice)

		if jsonDecoder(err, w) != nil {
			return
		}

	 	insertDevice(&newDevice, api.db)

		defer r.Body.Close()
	}
}
func (api * API) logIn(w http.ResponseWriter, r *http.Request) {
	var login Login;
	
	if r.Method == "POST" {
		log.Printf("New Request %s", r.URL)
		// Check if the content type is correct
		if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			
		}

		// Decoder our data into json.
		decoder := json.NewDecoder(r.Body)

		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&login)

		if jsonDecoder(err, w) != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("%s", err.Error())
		}
		// From this struct we must now return a bit id to the device. 
		session, err := getSession(api.db, &login)

		if err != nil {
			log.Printf("%s", err.Error())
			w.Write([]byte(err.Error()))
		}
	
	w.Header().Set("Content-Type", "application/json")
	
	json.NewEncoder(w).Encode(session)
	defer r.Body.Close()

    }

	}


func (api * API) getDailyData(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()

	if r.Method == "GET" {
		deviceID := r.URL.Query().Get("deviceID")
		timePeriod := r.URL.Query().Get("timePeriod")
		
		if deviceID == "" {
			http.Error(w, "Must provide deviceID", http.StatusBadRequest)
		} 
		if timePeriod == "" {
			http.Error(w, "Must provide timePeriod", http.StatusBadRequest)
		}
		
		mapper, err := getLatestDataDay(api.db, timePeriod, deviceID)

		if err != nil {
			http.Error(w, "Error getting latest data day", http.StatusBadGateway)
		}


		errs := json.NewEncoder(w).Encode(mapper)
		if errs != nil {
			http.Error(w, "Error encoding latest data day", http.StatusBadGateway)
		}
		defer r.Body.Close()
	}
}

func (api * API) logInApp(w http.ResponseWriter, r *http.Request) {
	log.Printf("New request %s", r.URL)
	if r.Method == "POST" {
		if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			log.Printf("NOT CONTENT TYPE")
		}
		
		decoder := json.NewDecoder(r.Body)
		var login UserPass;
		log.Printf("COULD CREATE DECODER")
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&login)

		if jsonDecoder(err, w) != nil {
			log.Printf("JSON ERROR %s", err)
			
		}
		log.Printf("BEFORE LOGIN")
		errs := LogIn(api.db, login)
		log.Printf("AFTER LOGIN")
		if errs != nil {
			log.Printf("%s",errs.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errs.Error()))
			
		}

		w.WriteHeader(http.StatusOK)

		defer r.Body.Close()
	}

}

func (api * API) newUser(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
	log.Printf("New Request %s", r.URL)

	if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			
		}

		var newUser UserPass;
		decoder := json.NewDecoder(r.Body)
		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&newUser)

		if jsonDecoder(err, w) != nil {
			log.Printf("JSON ERROR %s", err)
			
		}

		w.Header().Set("Content-Type", "application/json")

		// Check if out counter has reached zero and return new session
		var errs = insertNewUser(newUser, api.db)

		if errs != nil {
			w.Write([]byte(err.Error()))
			
			
			
		} else {
			w.WriteHeader(http.StatusOK)
		}
		defer r.Body.Close()
	}
}

func (api * API) newSessionData(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
	log.Printf("New Request %s", r.URL)



	if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			
		}

		var session SessionData;
		decoder := json.NewDecoder(r.Body)
		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&session)

		if jsonDecoder(err, w) != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/json")

		
		// Check if out counter has reached zero and return new session
		
		var newSession, errs  = insertSessionData(session, api.db)

		if errs != nil {
			
			w.Write([]byte(errs.Error()))
		}
		json.NewEncoder(w).Encode(newSession)
		
		defer r.Body.Close()
	}
	
}


// this function checks if theres any http errors if so then it will write to the response body
func jsonDecoder(err error, w http.ResponseWriter) error {

	if err != nil {
        var syntaxError *json.SyntaxError
        var unmarshalTypeError *json.UnmarshalTypeError

        switch {
        // Catch any syntax errors in the JSON and send an error message
        // which interpolates the location of the problem to make it
        // easier for the client to fix.
        case errors.As(err, &syntaxError):
            msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
            http.Error(w, msg, http.StatusBadRequest)

        // In some circumstances Decode() may also return an
        // io.ErrUnexpectedEOF error for syntax errors in the JSON. There
        // is an open issue regarding this at
        // https://github.com/golang/go/issues/25956.
        case errors.Is(err, io.ErrUnexpectedEOF):
            msg := "Request body contains badly-formed JSON"
            http.Error(w, msg, http.StatusBadRequest)

        // Catch any type errors, like trying to assign a string in the
        // JSON request body to a int field in our Person struct. We can
        // interpolate the relevant field name and position into the error
        // message to make it easier for the client to fix.
        case errors.As(err, &unmarshalTypeError):
            msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
            http.Error(w, msg, http.StatusBadRequest)

        // Catch the error caused by extra unexpected fields in the request
        // body. We extract the field name from the error message and
        // interpolate it in our custom error message. There is an open
        // issue at https://github.com/golang/go/issues/29035 regarding
        // turning this into a sentinel error.
        case strings.HasPrefix(err.Error(), "json: unknown field "):
            fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
            msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
            http.Error(w, msg, http.StatusBadRequest)

        // An io.EOF error is returned by Decode() if the request body is
        // empty.
        case errors.Is(err, io.EOF):
            msg := "Request body must not be empty"
            http.Error(w, msg, http.StatusBadRequest)

        // Catch the error caused by the request body being too large. Again
        // there is an open issue regarding turning this into a sentinel
        // error at https://github.com/golang/go/issues/30715.
        case err.Error() == "http: request body too large":
            msg := "Request body must not be larger than 1MB"
            http.Error(w, msg, http.StatusRequestEntityTooLarge)

        // Otherwise default to logging the error and sending a 500 Internal
        // Server Error response.
        default:
            log.Println(err.Error())
            http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        }
        return err
} else {
	return nil
}
}