package main

// These are standard Go libraries:
import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	// In this function HTML data is proccessed using a stream
	html, err := parseHTML(data)
	if err != nil {
		log.Fatalf("error parsing HTML: %v", err)
	}

	// Returning an output with number of <li> elements:
	fmt.Printf("List with the most <li> children has %d items.\n", countItems(html))
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

func parseHTML(html []byte) (*Node, error) {

	// making a stream
	parser := NewParser(bytes.NewReader(html))
	root := &Node{Tag: "root"}
	stack := []*Node{root} // stack to store elements

	///
	inside := false
	_ = inside
	count := 0

	///

	// s := *parser

	// fmt.Printf("s.r: %v\n", s.r)

	// looping through HTML elements until we reach the end of a file
	for {
		token, err := parser.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// fmt.Printf("current: %v\n", token)

		// Checking if token is start/end tag to keep track of nested tags inside a html
		// If it's start tag: we append this tag to stack as a node of previous node
		// If the ending tag matches with current first element on stack it pops it out
		switch t := token.(type) {
		case StartTagToken:
			if t.TagName == "ul" {
				inside = true
				// fmt.Printf("start token: %v\n", token)
			} else if t.TagName == "li" && inside {
				// fmt.Printf("\"start li \": %v\n", "start li ")
				count++
			}
			newNode := &Node{Tag: t.TagName}
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, newNode)
			stack = append(stack, newNode)
		case EndTagToken:
			if t.TagName == "ul" {
				inside = false
				fmt.Fprintf(os.Stdout, "xxx				count: %v\n", []any{count}...)
				// fmt.Printf("end token: %v\n", token)
				count = 0
			}

			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		}

	}

	// Returning first node from the function so we can traverse it later and count <li> elements
	return root, nil
}

// This function count <li> tags inside <ul> tags recursively
func countItems(html *Node) int {

	maxItems := 0
	var countChildren func(node *Node)

	countChildren = func(node *Node) {

		if node.Tag == "ul" {
			count := 0

			for _, child := range node.Children {
				if child.Tag == "li" {
					count++
				}
			}

			fmt.Printf("Found <ul> with %d <li> elements\n", count) // Debug log
			if count > maxItems {
				maxItems = count
			}
		}
		for _, child := range node.Children {
			countChildren(child)
		}
	}
	countChildren(html)

	return maxItems
}

// Tree Node object: Tag is representing html elements ("divs, ul, li etc.")
type Node struct {
	Tag      string
	Children []*Node
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

				// fmt.Printf("name: %v\n", name)

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
