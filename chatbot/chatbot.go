package chatbot

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	cors "github.com/heppu/simple-cors"
	controller "github.com/nada-hammad/readily/controller"
)

var (
	// WelcomeMessage A constant to hold the welcome message
	WelcomeMessage = "" +
		"Welcome to Readily, your favourite chatbot library! \n" +
		"Here's a list of commands you can try: \n" +
		"0- help\n" +
		"1.0- get book <titleName>\n" +
		"1.1-  get book title (after using command 1)\n" +
		"1.2-  get book link  (after using command 1)\n" +
		"1.3-  get book number of pages (after using command 1)\n" +
		"1.4-  get book format (after using command 1)\n" +
		"1.5-  get book authors (after using command 1)\n" +
		"1.6-  get book average rating (after using command 1)\n" +
		"1.7-  get book publication year (after using command 1)\n" +
		"1.8-  get book description (after using command 1)\n" +
		"1.9-  get book language code (after using command 1)\n" +
		"1.10- get book publisher (after using command 1)\n" +
		"1.11- get book similar books (after using command 1)\n" +
		"2.0-  get author <authorName>\n" +
		"2.1-  get author works count\n" +
		"2.2-  get author gender\n" +
		"2.3-  get author hometown\n" +
		"2.4-  get author book titles\n" +
		"3-    get latest reviews\n"

	key = "mpTE2wR5Fx0T3GjYwHpug"

	// sessions = {
	//   "uuid1" = Session{...},
	//   ...
	// }
	sessions = map[string]Session{}

	processor = sampleProcessor
)

type (
	// Session Holds info about a session
	Session map[string]interface{}

	// JSON Holds a JSON object
	//JSON map[string]interface{}

	// Processor Alias for Process func
	Processor func(session Session, message string) (string, error)
)

func sampleProcessor(session Session, message string) (string, error) {
	isGetBook := strings.HasPrefix(strings.ToLower(message), "get book")
	if isGetBook {
		bookTitle := strings.TrimPrefix(message, "get book")
		if len(bookTitle) != 0 {
			book := controller.GetBookByTitle(bookTitle, key)
			session["book"] = book // book is a JSON map
			return fmt.Sprintf("OK, I found the %s", book["title"]), nil
		} else {
			return "", fmt.Errorf("Please enter a book title!")
		}
	}

	isGetLatestReviews := strings.HasPrefix(strings.ToLower(message), "get latest reviews")
	if isGetLatestReviews {
		reviews := controller.GetRecentReviews(key)
		reviewsArr := reviews["reviews"].([]controller.Review)
		allReviews := ""
		for _, review := range reviewsArr {
			allReviews += fmt.Sprintf("Book title: %s \n", review.BookTitle)
			allReviews += fmt.Sprintf("Body: %s \n", review.Body)
		}
		return allReviews, nil
	}

	return fmt.Sprintf(WelcomeMessage), nil
}

// withLog Wraps HandlerFuncs to log requests to Stdout
func withLog(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := httptest.NewRecorder()
		fn(c, r)
		log.Printf("[%d] %-4s %s\n", c.Code, r.Method, r.URL.Path)

		for k, v := range c.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)
		c.Body.WriteTo(w)
	}
}

// writeJSON Writes the JSON equivilant for data into ResponseWriter w
func writeJSON(w http.ResponseWriter, data controller.JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ProcessFunc Sets the processor of the chatbot
func ProcessFunc(p Processor) {
	processor = p
}

// handleWelcome Handles /welcome and responds with a welcome message and a generated UUID
func handleWelcome(w http.ResponseWriter, r *http.Request) {
	// Generate a UUID.
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))

	// Create a session for this UUID
	sessions[uuid] = Session{}

	// Write a JSON containg the welcome message and the generated UUID
	writeJSON(w, controller.JSON{
		"uuid":    uuid,
		"message": WelcomeMessage,
	})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// Make sure only POST requests are handled
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	// Make sure a UUID exists in the Authorization header
	uuid := r.Header.Get("Authorization")
	if uuid == "" {
		http.Error(w, "Missing or empty Authorization header.", http.StatusUnauthorized)
		return
	}

	// Make sure a session exists for the extracted UUID
	session, sessionFound := sessions[uuid]
	if !sessionFound {
		http.Error(w, fmt.Sprintf("No session found for: %v.", uuid), http.StatusUnauthorized)
		return
	}

	// Parse the JSON string in the body of the request
	data := controller.JSON{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Make sure a message key is defined in the body of the request
	_, messageFound := data["message"]
	if !messageFound {
		http.Error(w, "Missing message key in body.", http.StatusBadRequest)
		return
	}

	// Process the received message
	message, err := processor(session, data["message"].(string))
	if err != nil {
		http.Error(w, err.Error(), 422 /* http.StatusUnprocessableEntity */)
		return
	}

	// Write a JSON containg the processed response
	writeJSON(w, controller.JSON{
		"message": message,
	})
}

// handle Handles /
func handle(w http.ResponseWriter, r *http.Request) {
	body :=
		"<!DOCTYPE html><html><head><title>Chatbot</title></head><body><pre style=\"font-family: monospace;\">\n" +
			"Available Routes:\n\n" +
			"  GET  /welcome -> handleWelcome\n" +
			"  POST /chat    -> handleChat\n" +
			"  GET  /        -> handle        (current)\n" +
			"</pre></body></html>"
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, body)
}

func bookHandle(w http.ResponseWriter, r *http.Request) {
	book := controller.GetBookByTitle("Harry potter", "mpTE2wR5Fx0T3GjYwHpug")
	//book := controller.GetBookByTitle("The Autobiography of Malcolm X", "mpTE2wR5Fx0T3GjYwHpug")
	fmt.Println(book)
	writeJSON(w, book)

	// author := controller.GetAuthorInfo("Dan Brown", "mpTE2wR5Fx0T3GjYwHpug")
	// fmt.Println(author)
	// writeJSON(w, author)

	// reviews := controller.GetRecentReviews("mpTE2wR5Fx0T3GjYwHpug")
	// fmt.Println(reviews)
	// writeJSON(w, reviews)
}

// Engage Gives control to the chatbot
func Engage(addr string) error {
	// HandleFuncs
	mux := http.NewServeMux()
	mux.HandleFunc("/welcome", withLog(handleWelcome))
	mux.HandleFunc("/chat", withLog(handleChat))
	mux.HandleFunc("/", withLog(handle))
	mux.HandleFunc("/book", withLog(bookHandle))
	mux.HandleFunc("/review", withLog(bookHandle))
	return http.ListenAndServe(addr, cors.CORS(mux))
}
