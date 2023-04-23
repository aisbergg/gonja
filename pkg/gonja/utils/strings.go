package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
)

// Ellipsis truncates a string to a given length and appends an ellipsis sign
// (…).
func Ellipsis(text string, length int) string {
	runes := []rune(text)
	if len(runes) <= length {
		return text
	}
	return string(runes[:length-1]) + "…"
}

var htmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&quot;",
	`'`, "&#39;",
)

// HTMLEscape escapes HTML-unsafe characters with Unicode escapes. The resulting
// escaped string differs slightly from `html.EscapeString`, because the latter
// one uses `&#34;` for quotes instead of `&quot;`
func HTMLEscape(in string) string {
	return htmlEscaper.Replace(in)
}

const filterIRIChars = "/#%[]=:;$&()+,!?*@'~"

// IRIEncode converts an Internationalized Resource Identifier (IRI) to a string
// that is suitable for inclusion in a URL.
func IRIEncode(in string) string {
	var b bytes.Buffer

	for _, r := range in {
		if strings.ContainsRune(filterIRIChars, r) {
			b.WriteRune(r)
		} else {
			b.WriteString(url.QueryEscape(string(r)))
		}
	}

	return b.String()
}

//revive:disable:line-length-limit

const loremIpsumText = `Lorem ipsum dolor sit amet, consectetur adipisici elit, sed eiusmod tempor incidunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquid ex ea commodi consequat. Quis aute iure reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint obcaecat cupiditat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.
Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi.
Nam liber tempor cum soluta nobis eleifend option congue nihil imperdiet doming id quod mazim placerat facer possim assum. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat. Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat.
Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis.
At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, At accusam aliquyam diam diam dolore dolores duo eirmod eos erat, et nonumy sed tempor et et invidunt justo labore Stet clita ea et gubergren, kasd magna no rebum. sanctus sea sed takimata ut vero voluptua. est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat.
Consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.`

//revive:enable:line-length-limit

const loremIpsumDictRaw = `a ac accumsan ad adipiscing aenean aliquam aliquet amet
ante aptent arcu at auctor augue bibendum blandit class commodo condimentum
congue consectetuer consequat conubia convallis cras cubilia cum curabitur curae
cursus dapibus diam dictum dictumst dignissim dis dolor donec dui duis egestas
eget eleifend elementum elit enim erat eros est et etiam eu euismod facilisi
facilisis fames faucibus felis fermentum feugiat fringilla fusce gravida
habitant habitasse hac hendrerit hymenaeos iaculis id imperdiet in inceptos
integer interdum ipsum justo lacinia lacus laoreet lectus leo libero ligula
litora lobortis lorem luctus maecenas magna magnis malesuada massa mattis mauris
metus mi molestie mollis montes morbi mus nam nascetur natoque nec neque netus
nibh nisi nisl non nonummy nostra nulla nullam nunc odio orci ornare parturient
pede pellentesque penatibus per pharetra phasellus placerat platea porta
porttitor posuere potenti praesent pretium primis proin pulvinar purus quam quis
quisque rhoncus ridiculus risus rutrum sagittis sapien scelerisque sed sem
semper senectus sit sociis sociosqu sodales sollicitudin suscipit suspendisse
taciti tellus tempor tempus tincidunt torquent tortor tristique turpis
ullamcorper ultrices ultricies urna ut varius vehicula vel velit venenatis
vestibulum vitae vivamus viverra volutpat vulputate`

var (
	loremIpsumParagraphs = strings.Split(loremIpsumText, "\n")
	loremIpsumWords      = strings.Fields(loremIpsumText)
	loremIpsumDict       = strings.Fields(loremIpsumDictRaw)
)

// LoremIpsum returns a lorem ipsum text.
func LoremIpsum(count int, method string) (string, error) {
	var out strings.Builder
	switch method {
	case "b":
		for i := 0; i < count; i++ {
			if i > 0 {
				out.WriteString("\n")
			}
			par := loremIpsumParagraphs[i%len(loremIpsumParagraphs)]
			out.WriteString(par)
		}
	case "w":
		for i := 0; i < count; i++ {
			if i > 0 {
				out.WriteString(" ")
			}
			word := loremIpsumWords[i%len(loremIpsumWords)]
			out.WriteString(word)
		}
	case "p":
		for i := 0; i < count; i++ {
			if i > 0 {
				out.WriteString("\n")
			}
			out.WriteString("<p>")
			par := loremIpsumParagraphs[i%len(loremIpsumParagraphs)]
			out.WriteString(par)
			out.WriteString("</p>")

		}
	default:
		return "", fmt.Errorf("unsupported method: %s", method)
	}

	return out.String(), nil
}

// Lipsum generates `n` paragraphs of text from lorem ipsum words. Each
// paragraph will contain between `min` and `max` random words.
func Lipsum(n int, html bool, min int, max int) string {
	result := []string{}

	for i := 0; i < n; i++ {
		nextCapitalized := true
		lastComma, lastFullstop := 0, 0
		word := ""
		last := ""
		p := []string{}

		// each paragraph contains out of min to max words.
		for j := min; j < max; j++ {
			for {
				word = loremIpsumDict[rand.Intn(len(loremIpsumDict))]
				if word != last {
					last = word
					break
				}
			}

			if nextCapitalized {
				word = strings.Title(word)
				nextCapitalized = false
			}

			if j-(3+rand.Intn(5)) > lastComma {
				// Add comas
				lastComma = j
				lastFullstop += 2
				word += ","
			} else if j-(10+rand.Intn(10)) > lastFullstop {
				// Add end of sentences
				lastComma, lastFullstop = j, j
				word += "."
				nextCapitalized = true
			}

			p = append(p, word)
		}

		// # ensure that the paragraph ends with a dot.
		str := strings.Join(p, " ")
		if strings.HasSuffix(str, ",") {
			str = str[:len(str)-1] + "."
		} else if !strings.HasSuffix(str, ".") {
			str += "."
		}

		result = append(result, str)
	}

	if !html {
		return strings.Join(result, "\n\n")
	}
	htmlResult := []string{}
	for _, p := range result {
		htmlResult = append(htmlResult, fmt.Sprintf(`<p>%s<p>`, p))
	}
	return strings.Join(htmlResult, "\n")
}
