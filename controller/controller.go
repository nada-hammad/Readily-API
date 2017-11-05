package controller

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var (
	apiRoot = "http://www.goodreads.com/"
)

// JSON Holds a JSON object
type JSON map[string]interface{}

type Response struct {
	Author  Author   `xml:"author"`
	Book    Book     `xml:"book"`
	Reviews []Review `xml:"reviews>review"`
}

type Book struct {
	ID              string   `xml:"id"`
	Title           string   `xml:"title"`
	Link            string   `xml:"link"`
	ImageURL        string   `xml:"image_url"`
	NumPages        string   `xml:"num_pages"`
	Format          string   `xml:"format"`
	Authors         []string `xml:"authors>author>name"`
	ISBN            string   `xml:"isbn"`
	AverageRating   string   `xml:"average_rating"`
	PublicationYear string   `xml:"publication_year"`
	Description     string   `xml:"description"`
	LanguageCode    string   `xml:"language_code"`
	Publisher       string   `xml:"publisher"`
	SimilarBooks    []string `xml:"similar_books>book>title"`
}

func (b Book) Author() string {
	return b.Authors[0]
}

type Author struct {
	Id         string   `xml:"id,attr"`
	ID         string   `xml:"id"`
	Name       string   `xml:"name"`
	WorksCount string   `xml:"works_count"`
	Gender     string   `xml:"gender"`
	Hometown   string   `xml:"hometown"`
	BookTitles []string `xml:"books>book>title"`
}

type Review struct {
	Body      string `xml:"body"`
	BookTitle string `xml:"book>title"`
}

// API Calls

func GetBookByTitle(title, key string) JSON {
	title = url.QueryEscape(title)
	uri := apiRoot + "book/title.xml?key=" + key + "&title=" + title
	response := &Response{}
	getData(uri, response)

	book := JSON{
		"title":           response.Book.Title,
		"link":            response.Book.Link,
		"imageURL":        response.Book.ImageURL,
		"numPages":        response.Book.NumPages,
		"format":          response.Book.Format,
		"authors":         response.Book.Authors,
		"isbn":            response.Book.ISBN,
		"rating":          response.Book.AverageRating,
		"publicationYear": response.Book.PublicationYear,
		"description":     response.Book.Description,
		"language_code":   response.Book.LanguageCode,
		"publisher":       response.Book.Publisher,
		"similarBooks":    response.Book.SimilarBooks,
	}

	return book
}

// not used
func GetBookId(isbn, key string) string {
	uri := apiRoot + "book/isbn/" + isbn + ".xml?key=" + key
	response := &Response{}
	getData(uri, response)
	return response.Book.ID
}

// not used
func GetBook(id, key string) JSON {
	uri := apiRoot + "book/show/" + id + ".xml?key=" + key
	response := &Response{}
	getData(uri, response)

	book := JSON{
		"id":              response.Book.ID,
		"title":           response.Book.Title,
		"link":            response.Book.Link,
		"imageURL":        response.Book.ImageURL,
		"numPages":        response.Book.NumPages,
		"format":          response.Book.Format,
		"authors":         response.Book.Authors,
		"isbn":            response.Book.ISBN,
		"rating":          response.Book.AverageRating,
		"publicationYear": response.Book.PublicationYear,
		"description":     response.Book.Description,
		"language_code":   response.Book.LanguageCode,
		"publisher":       response.Book.Publisher,
		"similarBooks":    response.Book.SimilarBooks,
	}

	return book
}

func GetRecentReviews(key string) JSON {
	uri := apiRoot + "review/recent_reviews?format=xml&key=" + key
	response := &Response{}
	getData(uri, response)

	reviews := JSON{
		"reviews": response.Reviews,
	}

	return reviews
}

func GetAuthorIDbyName(name string, key string) string {
	uri := apiRoot + "api/author_url/" + name + "?key=" + key

	response := &Response{}
	getData(uri, response)

	return response.Author.Id
}

func GetAuthorInfoById(id, key string) JSON {
	uri := apiRoot + "author/show/" + id + "?format=xml&key=" + key

	response := &Response{}
	getData(uri, response)

	author := JSON{
		"name":       response.Author.Name,
		"worksCount": response.Author.WorksCount,
		"gender":     response.Author.Gender,
		"hometown":   response.Author.Hometown,
		"bookTitles": response.Author.BookTitles,
	}

	return author
}

func GetAuthorInfo(name string, key string) JSON {
	id := GetAuthorIDbyName(name, key)
	author := GetAuthorInfoById(id, key)

	return author
}

// Data Handling

func getData(uri string, i interface{}) {
	data := getRequest(uri)
	xmlUnmarshal(data, i)
}

func getRequest(uri string) []byte {
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	return body
}

func xmlUnmarshal(b []byte, i interface{}) {
	err := xml.Unmarshal(b, i)
	if err != nil {
		log.Fatal(err)
	}
}
