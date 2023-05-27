package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

const charLimit = 250

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

		rgxExpr, err := buildRegexExprWithQuery(query)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		results := searcher.Search(rgxExpr)
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)

		// set false encoder option escape html
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

func buildRegexExprWithQuery(query url.Values) (*regexp.Regexp, error) {
	q := query.Get("q")
	if q == "" || len(q) < 1 {
		return nil, fmt.Errorf("missing search query in URL params")
	}

	rgx := "(?i)(%s)"
	if caseSensitiveParam := query.Get("cs"); caseSensitiveParam == "on" {
		rgx = "(%s)"
	}

	if wholeWordParam := query.Get("ww"); wholeWordParam == "on" {
		rgx = fmt.Sprintf("\\b%s\\b", rgx)
	}

	rgxExpr, _ := regexp.Compile(fmt.Sprintf(rgx, q))

	return rgxExpr, nil
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

// TODO add go doc
func (s *Searcher) Search(rgxExpr *regexp.Regexp) []string {

	idxs := s.SuffixArray.FindAllIndex(rgxExpr, -1)

	var results []string

	var fromIdx int
	var toIdx int

	// si es el primero de un parrafo
	if len(idxs) > 0 {
		fromIdx = idxs[0][0] - charLimit
		toIdx = idxs[0][1] + charLimit
	}

	for _, idx := range idxs {

		// si esta contenido en el anterior, cambio el to y sigo
		if idx[1] < toIdx {
			toIdx = idx[1] + charLimit

		} else {

			if fromIdx < 0 {
				fromIdx = 0
			}

			if toIdx > len(s.CompleteWorks)-1 {
				toIdx = len(s.CompleteWorks) - 1
			}

			// si no lo esta, armo el mensaje y limpio los indices
			results = append(results, s.CompleteWorks[fromIdx:toIdx])

			fromIdx = idx[0] - charLimit
			toIdx = idx[1] + charLimit

		}

	}

	// si quedo algo lo agrego
	if fromIdx != 0 && toIdx != 0 {

		if fromIdx < 0 {
			fromIdx = 0
		}

		if toIdx > len(s.CompleteWorks)-1 {
			toIdx = len(s.CompleteWorks) - 1
		}

		results = append(results, s.CompleteWorks[fromIdx:toIdx])
	}

	return results
}
