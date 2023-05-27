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
	"strings"
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

	// multiple values support
	q = strings.Replace(q, " ", "|", -1)

	// case-insensitive by default
	rgx := "(?i)(%s)"
	if caseSensitiveParam := query.Get("cs"); caseSensitiveParam == "on" {
		rgx = "(%s)"
	}

	// match only whole words
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

	results := []string{}
	str := []string{}

	var previousToIdx int

	for _, idx := range idxs {

		if idx[1] < previousToIdx+charLimit {

			// if the value is the first one
			if previousToIdx == 0 {
				// To avoid runtime error slice bounds out of range in case that the match is in the first or last words
				if idx[0]-charLimit < 0 {
					previousToIdx = 0
				}
			}

			// si es menor incluyo todo el texto hasta el find de este valor

			prevStr := s.CompleteWorks[previousToIdx:idx[0]]

			// if it is the first block, first word must be complete
			if len(str) == 0 {
				prevStrArray := strings.Split(prevStr, " ")
				prevStr = strings.Join(prevStrArray[1:], " ")
			}

			str = append(str, prevStr, "<mark>", s.CompleteWorks[idx[0]:idx[1]], "</mark>")
			previousToIdx = idx[1]

		} else {

			// si el desde no esta incluido en el anterior, cierro el parrafo, lo agrego a result
			// y limpio la variable str

			// To avoid runtime error slice bounds out of range in case that the match is in the first or last words
			toIdx := previousToIdx + charLimit
			if toIdx > len(s.CompleteWorks)-1 {
				toIdx = len(s.CompleteWorks) - 1
			}

			postStr := s.CompleteWorks[previousToIdx:toIdx]
			postStrArray := strings.Split(postStr, " ")
			str = append(str, strings.Join(postStrArray[:len(postStrArray)-1], " "))

			results = append(results, strings.Join(str, ""))

			// cleans str buffer
			str = []string{}

			// start new block with the new-found value
			fromIdx := idx[0] - charLimit
			prevStr := s.CompleteWorks[fromIdx:idx[0]]
			prevStrArray := strings.Split(prevStr, " ")

			str = append(str, strings.Join(prevStrArray[1:], " "), "<mark>", s.CompleteWorks[idx[0]:idx[1]], "</mark>")

			previousToIdx = idx[1]
		}

	}

	return results
}
