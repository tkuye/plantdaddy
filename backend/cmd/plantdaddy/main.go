package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"os"
	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v4/pgxpool"
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
	api := &API{db: connectToDb(os.Getenv("CONNSTRING"))}
	http.HandleFunc("/auth-device", api.logIn)
	http.HandleFunc("/api/new-device", api.newDevice)
	http.HandleFunc("/api/login", api.logInApp)
	http.HandleFunc("/new-data",api.newSessionData)
	http.HandleFunc("/api/new-user", api.newUser)
	http.HandleFunc("/api/devices", api.getDevices)
	log.Println("Listening for requests at http://localhost:8000/")
	log.Fatal(http.ListenAndServe(":8000", nil))
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
			return
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
			return
		}

		// Decoder our data into json.
		decoder := json.NewDecoder(r.Body)
		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&login)

		if jsonDecoder(err, w) != nil {
			return
		}
		// From this struct we must now return a bit id to the device. 
		session, err := getSession(api.db, &login)

		if err != nil {
			w.Write([]byte(err.Error()))
		}
	
	w.Header().Set("Content-Type", "application/json")
	
	json.NewEncoder(w).Encode(session)
	defer r.Body.Close()

    }

	}


func (api * API) logInApp(w http.ResponseWriter, r *http.Request) {
	log.Printf("New request %s", r.URL)
	if r.Method == "POST" {
		if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var login UserPass;
		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&login)

		if jsonDecoder(err, w) != nil {
			log.Printf("JSON ERROR %s", err)
			return
		}

		errs := LogIn(api.db, login)

		if errs != nil {
			log.Printf("%s",errs.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errs.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)

		
	}

}

func (api * API) newUser(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
	log.Printf("New Request %s", r.URL)

	if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}

		var newUser UserPass;
		decoder := json.NewDecoder(r.Body)
		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&newUser)

		if jsonDecoder(err, w) != nil {
			log.Printf("JSON ERROR %s", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Check if out counter has reached zero and return new session
		var errs = insertNewUser(newUser, api.db)

		if errs != nil {
			w.Write([]byte(err.Error()))
			
			return
			
		} else {
			w.WriteHeader(http.StatusOK)
		}
		
	}
}

func (api * API) newSessionData(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
	log.Printf("New Request %s", r.URL)



	if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}

		var session SessionData;
		decoder := json.NewDecoder(r.Body)
		
		// Do not allow certain fields that are not approved 
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&session)

		if jsonDecoder(err, w) != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")

		
		// Check if out counter has reached zero and return new session
		var newSession, errs  = insertSessionData(session, api.db)

		if errs != nil {
			
			w.Write([]byte(errs.Error()))
		}
		json.NewEncoder(w).Encode(newSession)
		
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