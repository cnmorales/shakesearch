# ShakeSearch

## Introduction

Hi Pulley team! Thanks for sharing with me this challenge, it was really interesting and fun to do.

For this solution I took a more backend oriented approach with little changes in the front to make it easier
to use and read.

Here, I will outline the main changes, a bit of context behind each one and a list of this I would love to improve in
the future.

> Hosted in https://cnmorales-shakesearch.onrender.com


## Main changes

For my solution I took advantage of some features that https://pkg.go.dev/index/suffixarray offers.
Instead of using `Lookup` function for every word, I use `FindAllIndex` with some really cool regexes that help me a lot
in the search process, always prioritizing the computational complexity with only one scan.

## Features

### Case-sensitive parameter in form
This parameter enables a case-sensitive search, considering capital letters.

### Whole word parameter
This value offers the possibility to search the exact given word, excluding results that match only a few characters.

### Matching word highlight
Matching values are highlighted to assist the user in better understanding the search results.
This came really usefull during development, I was using the integrated search ctrl+f every time.

### Validate if a result already exists inside other result
Each paragraph can have more than one result, this erases similar blocks and highlight all currencies of the word
inside.

### Multiword search
Search more than one word is possible separating each one with spaces.

### Predictive search
Predictive search is now compatible, with a minimum of 3 characters in the text input, search words with that prefix in
all the document and return a list of 10 suggestions as maximum.

## Things I would like to improve or add in the future

- Paging Support (Using pages and offset values to retrieve only a part of the info.
- "Load more" button (ability to offer more text before and after each result, similar to what Github does in PullRequests).
- Quantity of results (Really simple, show the number of matches in all document).
- Error handling and validations on server side can be improved.
- Change project architecture to package oriented design.
- Highlight results in frontend instead of backend. (I really don't like inserting html marks in the result).
- Order multisearch results by best matches (if a multiword search found all given words together should be the first result the user see)
- Misspelling.

---