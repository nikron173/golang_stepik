package main

import (
	"bufio"
	json "encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

var (
	r          = regexp.MustCompile("@")
	androidReg = regexp.MustCompile("Android")
	msieReg    = regexp.MustCompile("MSIE")

	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func main() {
	FastSearch(os.Stdout)
}

type User struct {
	ListBrowser []string `json:"browsers,"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
}

func FastSearch(out io.Writer) {

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	seenBrowsers := make(map[string]bool)
	uniqueBrowsers := 0
	var foundUsers strings.Builder
	i := -1

	fileContents := bufio.NewScanner(file)

	for fileContents.Scan() {
		line := fileContents.Text()
		user := &User{}
		// fmt.Printf("%v %v\n", err, line)
		err := easyjson.Unmarshal([]byte(line), user)
		if err != nil {
			panic(err)
		}
		i++
		isAndroid := false
		isMSIE := false
		// fmt.Println("User:", *user)

		for _, browser := range user.ListBrowser {
			if androidReg.MatchString(browser) {
				isAndroid = true
				_, notSeenBefore := seenBrowsers[browser]
				if !notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}

			if msieReg.MatchString(browser) {
				isMSIE = true
				_, notSeenBefore := seenBrowsers[browser]
				if !notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}
		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := r.ReplaceAllString(user.Email, " [at] ")
		foundUsers.WriteString(fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email))
	}
	// fmt.Println("Count users:", i)
	fmt.Fprintln(out, "found users:\n"+foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

// suppress unused package warning

func easyjson9e1087fdDecodeHw3User(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.ListBrowser = nil
			} else {
				in.Delim('[')
				if out.ListBrowser == nil {
					if !in.IsDelim(']') {
						out.ListBrowser = make([]string, 0, 4)
					} else {
						out.ListBrowser = []string{}
					}
				} else {
					out.ListBrowser = (out.ListBrowser)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.ListBrowser = append(out.ListBrowser, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9e1087fdEncodeHw3User(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"browsers\":"
		out.RawString(prefix[1:])
		if in.ListBrowser == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.ListBrowser {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeHw3User(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeHw3User(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeHw3User(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeHw3User(l, v)
}
