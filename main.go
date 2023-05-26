package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		q := query.Get("q")
		if q == "" || len(q) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		rgx := "(?i)(%s)"
		if caseSensitiveParam := query.Get("cs"); caseSensitiveParam == "on" {
			rgx = "(%s)"
		}

		if wholeWordParam := query.Get("ww"); wholeWordParam == "on" {
			rgx = fmt.Sprintf("\\b%s\\b", rgx)
		}

		rgxExpr, err := regexp.Compile(fmt.Sprintf(rgx, q))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}

		results := searcher.Search(rgxExpr)
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)

		// set false encoder option escape html to highlight results
		// enc.SetEscapeHTML(false)

		err = enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	return nil
}

func (s *Searcher) Search(rgxExpr *regexp.Regexp) []string {

	idxs := s.SuffixArray.FindAllIndex(rgxExpr, -1)

	results := []string{}
	for _, idx := range idxs {

		// To avoid runtime error slice bounds out of range in case that the match is in the first or last words
		fromIdx := idx[0] - 250
		if fromIdx < 0 {
			fromIdx = 0
		}

		toIdx := idx[1] + 250
		if toIdx > len(s.CompleteWorks)-1 {
			toIdx = len(s.CompleteWorks) - 1
		}

		str := []string{s.CompleteWorks[fromIdx:idx[0]], "<mark>", s.CompleteWorks[idx[0]:idx[1]], "</mark>", s.CompleteWorks[idx[1]:toIdx]}

		results = append(results, strings.Join(str, ""))
	}

	return results
}
