package main

// These are standard Go libraries:
import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const defaultURL string = "http://en.wikipedia.org/wiki/List_of_lists_of_lists"

func main() {
	fmt.Print("running html parser...\n")
	HTMLParser()
}

func HTMLParser() {

	// Getting an input using 'url' flag.
	url := getInput()
	if len(url) == 0 {
		log.Fatal("invalid or no data provided")
	}

	// There is a function for getting html from provided url
	data, err := getHTML(url)
	if err != nil {
		log.Fatalf("failed to fetch data. %v", err)
	}

	biggestListElement, err := countElements(&data)
	if err != nil {
		log.Fatal("internal server error")
	}

	// Returning an output with number of <li> elements:
	fmt.Printf("List with the most <li> children has %d items.\n", biggestListElement)
}

func getInput() string {

	// -h flag description
	urlFlagPtr := flag.String("url", "", "URL of the webpage to parse")
	defaultUrlPtr := flag.Bool("d", true, "default url")
	flag.Parse()

	inputURL := *urlFlagPtr
	if len(inputURL) == 0 {

		if *defaultUrlPtr {
			return defaultURL
		}

		fmt.Println("usage: go run main.go -url <URL>")
		return ""
	}

	return inputURL
}

func getHTML(url string) ([]byte, error) {

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("html provider responded with status:  %d", res.StatusCode)
	}

	// Reading response using io.Reader interface
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func countElements(html *[]byte) (int, error) {

	parser := NewParser(bytes.NewReader(*html))

	inside := false
	count := 0
	biggest := 0

	max := func(v int) int {
		if v > biggest {
			return v
		}
		return biggest
	}

	// looping through HTML elements
	for {
		token, err := parser.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		switch t := token.(type) {
		case StartTagToken:
			if t.TagName == "ul" {
				inside = true
			} else if t.TagName == "li" && inside {
				count++
			}
		case EndTagToken:
			if t.TagName == "ul" {
				inside = false
				biggest = max(count)
				count = 0
			}
		}

	}

	return biggest, nil
}

// Html parser object
type Parser struct {
	r     *strings.Reader
	buf   bytes.Buffer // bytes buffer
	inTag bool         // This one is to know if parser is inside of a tag
	isEnd bool         // This one is telling that parser has reached the end of the tag
}

type StartTagToken struct {
	TagName string
}

type EndTagToken struct {
	TagName string
}

// Constructor for a html parser
func NewParser(r io.Reader) *Parser {
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return &Parser{r: strings.NewReader(buf.String())}
}

// Note: interface{} type in Golang is 'any' type so we can use it here to return anything from a function
func (d *Parser) NextToken() (interface{}, error) {

	// iterating through chars (Note: in Golang char type is called 'rune')
	for {
		ch, _, err := d.r.ReadRune()
		if err != nil {
			return nil, err
		}

		if d.inTag {
			if ch == '>' {
				d.inTag = false
				name := strings.Fields(d.buf.String())[0]
				d.buf.Reset()

				if d.isEnd {
					d.isEnd = false
					return EndTagToken{TagName: name}, nil
				}

				return StartTagToken{TagName: name}, nil
			}

			if ch == '/' && d.buf.Len() == 0 {
				d.isEnd = true
			} else {
				d.buf.WriteRune(ch)
			}
		} else {
			if ch == '<' {
				d.inTag = true
			}
		}
	}
}
