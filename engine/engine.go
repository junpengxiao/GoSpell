package engine

import (
	"appengine"
	"github.com/tonyshaw/GoSpell/database"
	"github.com/russross/blackfriday"
	"time"
	"unicode/utf8"
	"log"
	"bytes"
	"strings"
	"strconv"
)

type PageStruct struct {
	Title, Content, Kind, Key string
	Date time.Time
}

const page2db = 4 //translate page into database in following order Title, MiniContent, FullContent, RawText

/*Save the page into db. if Key is "" then save page into database. page.Content and page.Kind should be set.
  otherwise, update the original context via key
  func will save the content, i.e., markdown text, in data[3]. add <h1> </h1> into title and save in data[0]
  save data[0] + html generated by markdown in data[2]. extract an abstract from data[2] save in data[1] */
func Save(page *PageStruct, context appengine.Context) string {
	//check if page is null
	if len(page.Content) == 0 || len(page.Kind) == 0 {
		log.Println("engine.go save func, page content or kind is empty")
	}
	data := make([]string, page2db)
	data[3] = page.Content//RawText
	//Escape mathjax and d3 from original text
	var buffer bytes.Buffer
	buffer.Grow(len(page.Content))
	escapedtext := EscapeFromMarkdown(&buffer, page.Content)
	//Raw -> Markdown
	newbuf := bytes.NewBuffer(blackfriday.MarkdownCommon(buffer.Bytes()))
	data[2] = RecoverEscape(newbuf,escapedtext, page.Title)
	data[0] = data[2][:len("<h1></h1>\n")+len(page.Title)]
	//Extract abstract
	data[1] = data[2][len(data[0]):AwesomeCutoff(data[2])]
	log.Println(data)
	if (page.Key == "") {
		key, err := database.PutData(data, page.Kind, context)
		if err != nil { return "" } else { return key }
	} else {
		key, err := database.UpdateData(data, page.Key, context)
		if err != nil { return "" } else { return key }
	}
}

//Get Content via Key
func Get(key string, context appengine.Context) PageStruct {
	var ret PageStruct
	
	data, _ := database.GetData(key, context)
	ret.Content = data.Content
	ret.Date = data.Date
	ret.Key = data.NextKey
	
	return ret
}

//List 'number' of  1st level 'kind' contents. 'skip' is the number of contents u want to skip. 
func Query(kind string, skip, number int, context appengine.Context) []PageStruct {
	data, err := database.QueryData(kind, 0, skip, number, context)
	if err != nil {
		return nil
	}
	ret := make([]PageStruct, len(data))
	
	for i, tmp := range data {
		ret[i].Title = tmp.Content
		ret[i].Date = tmp.Date
		ttmp, _ := database.GetData(tmp.NextKey, context)
		ret[i].Key = ttmp.NextKey
		ret[i].Content = ttmp.Content
	}
	return ret
}

//This function designed for cutting off the html so that it can be displayed in nav page as abstract.
func AwesomeCutoff(str string) int {
	count := 0
	for index, runeVal := range str {
		count++
		if count > 150 && (runeVal == '.' || runeVal == '。' || runeVal =='，') {
			if (runeVal == '.') {
				for i:=index-1; i>1; i-- {
					if index - i > 5 { break }
					if str[i-1] == 0 && str[i]==' ' { continue }
				}
			}
			_, width := utf8.DecodeRuneInString(str[index:])
			return index + width
		}
	}
	return len(str)
}

/*
  keywords that will be used as escape identifer. kwbegin[i]<->kwend[i]
  0 : mathjax, begin{} end{}
  1 : d3js figure
  2 : mathjax, ref
  others: mathjax
*/
var kwbegin = []string{"\\begin","\\d3beg","\\ref{","$$","\\[","\\("}
var kwend =   []string{"\\end"  ,"\\d3end","}"    ,"$$","\\]","\\)"}
//original text will be replaced by pattern
var patternOri = "{{.}}"
var patternMar = "\\{\\{\\.\\}\\}"
func EscapeFromMarkdown(buff *bytes.Buffer, str string) []string {
	var index int
	var beginCount int //count the \begin keywords
	ret := make([]string,0,2048)//2048 escaped text in each page on average. this number can be seted optional 
	for i, flag := 0, -1; i < len(str); i++{
		if str[i] == '\\' || str[i] == '$' || (flag != 0 && str[i]=='}') {//\,$ are identifier, } only be useful when flag != 0
			if flag == -1 {
				for j, tmp := range kwbegin {
					if strings.HasPrefix(str[i:],tmp) {
						flag = j
						index = i
						i+=len(tmp)-1
						buff.WriteString(patternMar)
						beginCount = 0
						if flag == 0 {//begin with \\begin
							beginCount++
						}
						break
					}
				}
				if flag != -1 { continue }
			} else {
				//begin can be nest with others. to handle this, we use beginCount
				if strings.HasPrefix(str[i:], kwbegin[0]) {
					beginCount++
				}
				if strings.HasPrefix(str[i:], kwend[0]) {
					beginCount--
				}
				if beginCount == 0 && strings.HasPrefix(str[i:], kwend[flag]) {
					i += len(kwend[flag])-1
					if (flag == 0) {//handle \\end{...}
						for str[i] != '}' && i < len(str) { i++ }
					}
					if i == len(str) { i-- }
					flag = -1
					ret = append(ret, str[index:i+1])
					continue
				}
			}
		}
		if flag == -1 {	buff.WriteByte(str[i]) }
	}
	return ret
}

func skipBlanks(str string, reverse bool, startPoint int) int {
	i:=startPoint
	if reverse {
		for i>0 && (str[i-1]==' '||str[i-1]=='\t'||str[i-1]=='\n'||str[i-1]=='\r') { i-- }
	} else {
		for i<len(str) && (str[i]==' '||str[i]=='\t'||str[i]=='\n'||str[i]=='\r'){ i++ }
	}
	return i
}

func BuildD3JSReference(list []string) map[string]int {
	ret := make(map[string]int)
	count := 1
	for _, str := range list {
		if !strings.HasPrefix(str, kwbegin[1]) { continue }
		i := skipBlanks(str, false, len(kwbegin[1]))
		if len(str)-i < len("label") { continue }
		if strings.HasPrefix(strings.ToLower(str[i:i+5]),"label") {
			i := skipBlanks(str, false, i+len("label"))
			j := i
			for j < len(str) && str[j]!='\r' && str[j]!='\n' { j++ }
			if j>len(str) {
				log.Println("\nengine.go BuildD3JSReference, figure label is wrong" + str + "\n")
				continue
			}
			j = skipBlanks(str, true, j)
			ret[str[i:j]] = count
			count = count + 1
		}
	}
	return ret
}			

func RecoverEscape(buff *bytes.Buffer, strlist []string, title string) string{
	var ret bytes.Buffer
	ret.Grow(buff.Len()*2);
	plist := 0
	tmpbuf := make([]byte, len(patternOri))
	ret.WriteString("<h1>") //add title
	ret.WriteString(title)
	ret.WriteString("</h1>\n")
	ref := BuildD3JSReference(strlist)//Build D3js reference maps
	for {
		if buff.Len() <len(patternOri) {
			ret.Write(buff.Bytes())
			break
		}
		ch, _ := buff.ReadByte()
		if ch == '{' {
			index := 0
			tmpbuf[index] = '{'
			for index = 1; index != len(patternOri); index++ {
				tmpbuf[index], _ = buff.ReadByte()
				if tmpbuf[index] != patternOri[index] { break }
			}
			if index == len(patternOri) {
				if strings.HasPrefix(strlist[plist], kwbegin[1]) {//d3js
					ret.WriteString(D3JSFormat(strlist[plist], &ref))
				} else {//mathjax
					ret.WriteString(strlist[plist])
				}
				plist=plist+1
			} else {
				ret.Write(tmpbuf[:index+1])
			}
		}  else {
			ret.WriteByte(ch)
		}
		if buff.Len() == 0 { break }
	}
	if plist != len(strlist) { log.Println("engine, RecoverEscape, strlist isn't match") }
	return ret.String()
}

func D3JSFormat(str string, ref *map[string]int) string {
	index := skipBlanks(str, false, len(kwbegin[1]))
	name := ""
	if strings.HasPrefix(strings.ToLower(str[index:index+len("ref")]), "ref") {
		//reference
		index = skipBlanks(str, false, index + len("ref"))
		name = str[index:skipBlanks(str, true, len(str)-len(kwend[1]))]
		num := (*ref)[name]
		if num == 0 {
			return " Bad Reference, Name : " + name + " "
		} else {
			return "<a href=\"#Figure"+strconv.Itoa(num)+"\">Figure" + strconv.Itoa(num) + "</a>"
		}
	} else {
		//figures
		if strings.HasPrefix(strings.ToLower(str[index:index+len("label")]), "label") {//label
			index = skipBlanks(str, false, index+len("label"))
			j:=index
			for j<len(str) && str[j]!='\n' && str[j]!='\r' { j++ }
			if j>len(str) {
				log.Println("\nengine.go D3JSFormat, data format is wrong" + str + "\n")
				return " Figure Label is wrong: " + str + " "
			}
			j = skipBlanks(str,true,j)
			name = str[index:j]
			index = skipBlanks(str, false, j)
		}
		if strings.HasPrefix(strings.ToLower(str[index:index+len("type")]), "type") {//figure type
			index = skipBlanks(str, false, index+len("type"))
			j := index
			for j<len(str) && 'a'<=str[j] && str[j]<='z' { j++ } //type should be defined by 'a'-'z'
			if j>len(str) {
				return " Figure type is wrong: " + str + " "
			}
			figureType := str[index:j]
			//data should be in the newline right after type
			//figure data
			for j<len(str) && str[j] != '\n' { j++ }
			if j>len(str) {
				return " Figure data should be in a newline right after type "
			}
			index = skipBlanks(str, false, j)
			j = skipBlanks(str, true, len(str)-len(kwend[1]))
			refNum := "";
			if name != "" {
				refNum = "Figure"+strconv.Itoa((*ref)[name])
				refNum =`<div class="d3jsref" name="`+refNum+`" id="`+refNum+`">`+refNum+"</div>"
			}
			return `
<div class="d3jsfig">
    <div class="d3jscanvas">
        <div class="` + figureType + `" hidden>
            `+ str[index:j] + `
        </div>
    </div> 
`+ refNum +`
</p>
`			
		} else {
			return " Figure Type is wrong, can't find \"type\" keyword after \\d3beg or label "
		}
	}
	return " engine.go D3JSFormat internal error "
}	
