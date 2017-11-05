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
	WelcomeMessage = "Welcome, what do you want to order?"
	key            = "mpTE2wR5Fx0T3GjYwHpug"

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

	// check if there's a book and/or an author stored in the session
	//_, bookFound := session["book"]
	_, authorFound := session["author"]

	// book requests
	isGetBook := strings.HasPrefix(strings.ToLower(message), "get book")
	if isGetBook {
		bookTitle := strings.TrimPrefix(message, "get book")
		if len(bookTitle) != 0 {
			book := controller.GetBookByTitle(bookTitle, key)
			session["book"] = book // book is a JSON map
			return fmt.Sprintf("OK, I found the book %s", book["title"]), nil
		} else {
			return "", fmt.Errorf("Please enter a book title!")
		}
	}

	// reviews requests
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

	// get the author <author name>
	isGetTheAuthor := strings.HasPrefix(strings.ToLower(message), "get the author")
	if isGetTheAuthor {
		authorName := strings.TrimPrefix(message, "get the author")
		if len(authorName) != 0 {
			author := controller.GetAuthorInfo(authorName, key)
			session["author"] = author
			return fmt.Sprintf("OK, I found the author %s! What do you want to know?", author["name"]), nil
		}
		return "", fmt.Errorf("Please enter an author name!")

	}

	// get author <attribute>
	isGetAuthor := strings.HasPrefix(strings.ToLower(message), "get author")
	if isGetAuthor {
		attribute := strings.TrimPrefix(message, "get author ")

		isValidAttribute := (len(attribute) != 0 &&
			(strings.EqualFold(attribute, "number of works") ||
				strings.EqualFold(attribute, "works") ||
				strings.EqualFold(attribute, "gender") ||
				strings.EqualFold(attribute, "hometown") ||
				strings.EqualFold(attribute, "info")))

		if isValidAttribute && !authorFound {
			return "", fmt.Errorf("Please enter an author name!")
		} else if authorFound {
			author := session["author"].(controller.JSON)
			works := strings.Join(author["bookTitles"].([]string), "\n")

			// get author number of works
			if strings.EqualFold(attribute, "number of works") {
				if strings.EqualFold(author["worksCount"].(string), "") {
					return fmt.Sprintf("The author's number of works is not available"), nil
				}
				return fmt.Sprintf(author["worksCount"].(string)), nil

				// get author gender
			} else if strings.EqualFold(attribute, "gender") {
				if strings.EqualFold(author["gender"].(string), "") {
					return fmt.Sprintf("The author's gender is not available"), nil
				}
				return fmt.Sprintf(author["gender"].(string)), nil

				// get author hometown
			} else if strings.EqualFold(attribute, "hometown") {
				if strings.EqualFold(author["hometown"].(string), "") {
					return fmt.Sprintf("The author's hometown is not available"), nil
				}
				return fmt.Sprintf(author["hometown"].(string)), nil

				// get author works
			} else if strings.EqualFold(attribute, "works") {
				if strings.EqualFold(works, "") {
					return fmt.Sprintf("The author's works are not available"), nil
				}
				return fmt.Sprintf(works), nil

				// get author info
			} else if strings.EqualFold(attribute, "info") {
				info := "Name: " + author["name"].(string) + "\n" +
					"Number of works: " + author["worksCount"].(string) + "\n" +
					"gender: " + author["gender"].(string) + "\n" +
					"hometown: " + author["hometown"].(string) + "\n" +
					"works:\n" + works

				return fmt.Sprintf(info), nil
			}
		}
	}

	return fmt.Sprintf("So, you want %s! What else?", strings.ToLower("test")), nil
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
