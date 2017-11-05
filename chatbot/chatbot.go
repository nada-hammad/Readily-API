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
		"1.0-  get book <titleName>\n" +
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
		"2.1-  get author number of works\n" +
		"2.2-  get author works\n" +
		"2.3-  get author gender\n" +
		"2.4-  get author hometown\n" +
		"2.5-  get author info\n" +
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

	// check if there's a book and/or an author stored in the session
	_, bookFound := session["book"]
	_, authorFound := session["author"]

	// book requests
	isGetTheBook := strings.HasPrefix(strings.ToLower(message), "get the book")
	if isGetTheBook {
		bookTitle := strings.TrimPrefix(message, "get the book")
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
	isGetBook := strings.HasPrefix(strings.ToLower(message), "get book")
	if isGetBook {
		attribute := strings.TrimPrefix(message, "get book ")

		isValidAttribute := (len(attribute) != 0 &&
			(strings.EqualFold(attribute, "numPages") ||
				strings.EqualFold(attribute, "format") ||
				strings.EqualFold(attribute, "authors") ||
				strings.EqualFold(attribute, "isbn") ||
				strings.EqualFold(attribute, "rating") ||
				strings.EqualFold(attribute, "publicationYear") ||
				strings.EqualFold(attribute, "description") ||
				strings.EqualFold(attribute, "language_code") ||
				strings.EqualFold(attribute, "publisher")))

		if isValidAttribute && !bookFound {
			return "", fmt.Errorf("Please enter an author name!")
		} else if bookFound {
			book := session["book"].(controller.JSON)
			authors := strings.Join(book["authors"].([]string), ", ")

			// get author number of works
			if strings.EqualFold(attribute, "number of pages") {
				if strings.EqualFold(book["numPages"].(string), "") {
					return fmt.Sprintf("The book's number of pages is not available"), nil
				}
				return fmt.Sprintf(book["numPages"].(string)), nil

				// get author gender
			} else if strings.EqualFold(attribute, "format") {
				if strings.EqualFold(book["format"].(string), "") {
					return fmt.Sprintf("The book's format is not available"), nil
				}
				return fmt.Sprintf(book["format"].(string)), nil

				// get author hometown
			} else if strings.EqualFold(attribute, "isbn") {
				if strings.EqualFold(book["isbn"].(string), "") {
					return fmt.Sprintf("The book's isbn is not available"), nil
				}
				return fmt.Sprintf(book["isbn"].(string)), nil
				// get author works
			}else if strings.EqualFold(attribute, "rating") {
				if strings.EqualFold(book["rating"].(string), "") {
					return fmt.Sprintf("The book's rating is not available"), nil
				}
				return fmt.Sprintf(book["publicationYear"].(string)), nil
				// get author works
			} else if strings.EqualFold(attribute, "publication year") {
				if strings.EqualFold(book["publicationYear"].(string), "") {
					return fmt.Sprintf("The book's publication year is not available"), nil
				}
				return fmt.Sprintf(book["publicationYear"].(string)), nil
				// get author works
			}else if strings.EqualFold(attribute, "description") {
				if strings.EqualFold(book["description"].(string), "") {
					return fmt.Sprintf("The book's description is not available"), nil
				}
				return fmt.Sprintf(book["description"].(string)), nil
				// get author works
			}else if strings.EqualFold(attribute, "language code") {
				if strings.EqualFold(book["language_code"].(string), "") {
					return fmt.Sprintf("The book's language_code is not available"), nil
				}
				return fmt.Sprintf(book["language_code"].(string)), nil
				// get author works
			}else if strings.EqualFold(attribute, "authors") {
				if strings.EqualFold(authors, "") {
					return fmt.Sprintf("The book's authors are not available"), nil
				}
				return fmt.Sprintf(authors), nil

				// get author info
			} else if strings.EqualFold(attribute, "info") {
				info := "Number of pages: " + book["numPages"].(string) + "\n" +
					"Format: " + book["format"].(string) + "\n" +
					"ISBN: " + book["isbn"].(string) + "\n" +
					"Publication Year: " + book["publicationYear"].(string) + "\n" +
					"Description: " + book["description"].(string) + "\n" +
					"Language code: " + book["language_code"].(string) + "\n" +
					"Publisher: " + book["publisher"].(string) + "\n" +
					"Authors:\n" + authors

				return fmt.Sprintf(info), nil
			}
		}
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

