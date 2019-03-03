package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var dsn = "web:password@/poemcollection?parseTime=true"

type application struct {
	DB *sql.DB
}

type result struct {
	id      int
	title   string
	content string
	created time.Time
	expired time.Time
}

func main() {
	db, err := openDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &application{DB: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/home", home)
	mux.HandleFunc("/search", app.searchKeyword)

	port := ":8080"
	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func home(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("homepage.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}

	t.Execute(w, nil)
}

func (app *application) searchKeyword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	results, err := app.fullTextSearch(r.FormValue("keywords"))
	if err != nil {
		log.Println(err)
		return
	}

	if len(results) == 0 {
		w.Write([]byte("Sorry. No results were found."))
		return
	}

	for _, v := range results {
		fmt.Fprintf(w, "Title: %s\n", v.title)
		fmt.Fprintf(w, "%s\n\n\n", v.content)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (app *application) fullTextSearch(keyword string) ([]result, error) {
	stmt := `SELECT * FROM poems WHERE MATCH (title,content) AGAINST (? IN NATURAL LANGUAGE MODE)`

	rows, err := app.DB.Query(stmt, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []result{}

	for rows.Next() {
		r := result{}
		err = rows.Scan(&r.id, &r.title, &r.content, &r.created, &r.expired)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
