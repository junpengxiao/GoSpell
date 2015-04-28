package engine
import (
	"testing"
	"appengine/aetest"
	"log"
	"bytes"
	"github.com/russross/blackfriday"
)

func TestStoreData(t *testing.T) {
	context, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Close()

	var test PageStruct
	test.Title = "title"
	test.Content = "123"
	test.Kind = "art"
	test.Key = ""
	key := Save(&test, context)
	result := Query(test.Kind, 0, 1, context)
	fetch := result[0]
	
	//Check fetch
	if fetch.Title != "<h1>"+test.Title+"</h1>\n" {
		t.Errorf("content title is not correct " + fetch.Title)
	}
	if fetch.Content != "<p>"+test.Content+"</p>\n" {
		t.Errorf("content is not correct " +fetch.Content)
	}
	fetch = Get(fetch.Key, context)
	fetch = Get(fetch.Key, context)
	if fetch.Content != test.Content {
		t.Errorf("original content is not correct ", fetch.Content)
	}
	
	if fetch.Key != key {
		t.Errorf("Database structure should be circule")
	}

	//check update
	test.Content = "abc"
	test.Key = key
	test.Title = "tt"
	newkey := Save(&test, context)
	if newkey != key {
		t.Errorf("Update error, key is wrong")
	}
	fetch = Get(newkey, context)
	if fetch.Content != "<h1>tt</h1>\n" {
		t.Fatal()
	}
	fetch = Get(fetch.Key, context)
	if fetch.Content != "<p>abc</p>\n" {t.Errorf(fetch.Content)}
	fetch = Get(fetch.Key, context)
	if fetch.Content != "<h1>tt</h1>\n<p>abc</p>\n" {t.Errorf(fetch.Content)}
}
var content = `
## This is head
Hello \(f(x) = 3\) is a function. 

\begin{align}
  \begin{array}
  123
  \end{array}
  333
\end{align}

\[ f(x) = 4 \]$$z(y)=6$$

We can refer a d3js figure like this \d3beg ref figure name \d3end. And draw the figure. figure can be defined before reference or after it.
\d3beg
label figure name
type pie
item1 item2 item3
0.3   0.2   0.5
\d3end`

//Due to the details are hard to handle to achive the auto checking. we require to check it manually
func TestEscape(t *testing.T) {
	var buf  bytes.Buffer
	ret := EscapeFromMarkdown(&buf, content)
	log.Println("\n-----> Pls Check Escape Manually")
	log.Println("----->Original Text:")
	log.Println(content)
	log.Println("----->Escaped Text: ")
	log.Println(buf.String())
	log.Println("----->Escaped List: ")
	log.Println(ret)
}

func TestSkipBlanks(t *testing.T) {
	str:="\t\n\r hello \t\n\r"
	i:=skipBlanks(str, false, 0)
	j:=skipBlanks(str, true, len(str))
	if str[i:j] != "hello" {
		t.Errorf("skip blanks wrong " + str[i:j])
	}
}

func TestMarkdown(t *testing.T) {
	var buf bytes.Buffer
	EscapeFromMarkdown(&buf, content)
	str := string(blackfriday.MarkdownCommon(buf.Bytes()))
	log.Println("\n------->Test Markdown")
	log.Println(str)
}

func TestRecover(t *testing.T) {
	var buf bytes.Buffer
	ret := EscapeFromMarkdown(&buf, content)
	newbuf := bytes.NewBuffer(blackfriday.MarkdownCommon(buf.Bytes()))
	log.Println("\n------>Recovery: \n" + RecoverEscape(newbuf, ret, "Test") + "\n") //Check Manually
}

