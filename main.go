package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Book struct {
	Id      int      `json:"id"`
	Title   string   `json:"title"`
	Authors []string `json:"authors"`
	Year    int      `json:"year"`
}

var Books = []Book{
	{
		Id:      1,
		Title:   "Go на практике",
		Authors: []string{"Мэтт Батчер", "Мэтт Фарина"},
		Year:    2016,
	},
	{
		Id:      2,
		Title:   "Чистый код",
		Authors: []string{"Роберт Мартин"},
		Year:    2019,
	},
	{
		Id:      3,
		Title:   "Алгоритмы",
		Authors: []string{"Томас Кормен", "Чарльз Эрик Лейзерсон"},
		Year:    1989,
	},
	{
		Id:      4,
		Title:   "Чистая архитектура",
		Authors: []string{"Роберт Мартин"},
		Year:    2018,
	},
}

func GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idstr := r.URL.Query().Get("id")
	idint, err := strconv.Atoi(idstr)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	var foundBook Book
	for _, book := range Books {
		if book.Id == idint {
			foundBook = book
			break
		}
	}

	if foundBook.Id == 0 {
		handleError(w, http.StatusNotFound, fmt.Errorf("book with id %d not found", idint))
		return
	}

	data, err := json.Marshal(foundBook)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
	w.Write(data)
}

func AddBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newBook Book
	err := json.NewDecoder(r.Body).Decode(&newBook)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	newBook.Id = len(Books) + 1
	Books = append(Books, newBook)

	data, err := json.Marshal(newBook)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	w.Write(data)
}

func AddBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newBooks []Book
	err := json.NewDecoder(r.Body).Decode(&newBooks)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	for i := range newBooks {
		newBooks[i].Id = len(Books) + 1
		Books = append(Books, newBooks[i])
	}

	data, err := json.Marshal(newBooks)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	w.Write(data)
}

func AllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем квери параметры из URL
	query := r.URL.Query()
	limit := query.Get("limit")
	titleFilter := query.Get("title")
	sort := query.Get("sort")

	// Применяем фильтрацию по полю Title, если параметр titleFilter задан
	var filteredBooks []Book
	if titleFilter != "" {
		for _, book := range Books {
			if strings.Contains(book.Title, titleFilter) {
				filteredBooks = append(filteredBooks, book)
			}
		}
	} else {
		filteredBooks = Books
	}

	// Применяем сортировку по полю Id, если параметр sort равен "asc" или "desc"
	if sort == "asc" {
		sortBooksByIdAsc(filteredBooks)
	} else if sort == "desc" {
		sortBooksByIdDesc(filteredBooks)
	}

	// Применяем ограничение количества книг, если параметр limit задан
	if limit != "" {
		limitNum, err := strconv.Atoi(limit)
		if err != nil {
			handleError(w, http.StatusBadRequest, errors.New("invalid limit parameter"))
			return
		}

		// Проверяем, если параметр limit больше количества книг, то устанавливаем его равным количеству книг
		if limitNum > len(filteredBooks) {
			limitNum = len(filteredBooks)
		}

		filteredBooks = filteredBooks[:limitNum]
	}

	data, err := json.Marshal(filteredBooks)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}
	w.Write(data)

}

// Функция для сортировки книг по возрастанию Id
func sortBooksByIdAsc(books []Book) {
	sort.SliceStable(books, func(i, j int) bool {
		return books[i].Id < books[j].Id
	})
}

// Функция для сортировки книг по убыванию Id
func sortBooksByIdDesc(books []Book) {
	sort.SliceStable(books, func(i, j int) bool {
		return books[i].Id > books[j].Id
	})
}

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err)
		return
	}

	var book Book
	err = json.Unmarshal(data, &book)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	Books[book.Id] = book
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idstr := r.URL.Query().Get("id")
	idint, err := strconv.Atoi(idstr)
	if err != nil {
		handleError(w, http.StatusBadRequest, err)
		return
	}

	// Ищем индекс книги в слайсе Books
	index := -1
	for i, book := range Books {
		if book.Id == idint {
			index = i
			break
		}
	}

	if index == -1 {
		handleError(w, http.StatusNotFound, errors.New("book not found"))
		return
	}

	// Удаляем книгу из слайса
	Books = append(Books[:index], Books[index+1:]...)

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/book", GetBook).Methods(http.MethodGet)
	r.HandleFunc("/book", AddBook).Methods(http.MethodPost)
	r.HandleFunc("/book", DeleteBook).Methods(http.MethodDelete)
	r.HandleFunc("/book", UpdateBook).Methods(http.MethodPut)
	r.HandleFunc("/books", AllBooks).Methods(http.MethodGet)
	r.HandleFunc("/books", AddBooks).Methods(http.MethodPost)

	http.ListenAndServe("127.0.0.1:8080", r)
}

func handleError(w http.ResponseWriter, status int, err error) {
	result := map[string]interface{}{
		"error":  err.Error(),
		"status": http.StatusText(status),
	}

	data, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	w.WriteHeader(status)
	_, err = w.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
}
