package template

import (
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

type namingSchemeFn func(string) string

var knownAcronyms = map[string]string{}

var convertKeyFuncs = map[string]namingSchemeFn{
	"snake": strcase.ToSnake,
	"kebab": strcase.ToKebab,
	"camel": CamelCaseWithInitialisms,
	"none":  func(s string) string { return s },
}

var namifyPerScheme = map[string]namingSchemeFn{
	"camel": CamelCaseWithInitialisms,
	"none":  DefaultNamifier,
}

var convertKey = convertKeyFuncs["none"]
var namify = namifyPerScheme["none"]

// SetKnownAcronyms configures the list of acronyms/initialisms that should be
// preserved when generating CamelCase identifiers.
func SetKnownAcronyms(acronyms []string) {
	knownAcronyms = make(map[string]string, len(acronyms))
	for _, acronym := range acronyms {
		acronym = strings.TrimSpace(acronym)
		if acronym == "" {
			continue
		}
		knownAcronyms[strings.ToLower(acronym)] = acronym
	}
}

// CamelCaseWithInitialisms converts a string to CamelCase while preserving any
// configured acronyms/initialisms as-is.
func CamelCaseWithInitialisms(sentence string) string {
	if len(knownAcronyms) == 0 {
		return strcase.ToCamel(sentence)
	}

	words := splitIdentifierWords(sentence)
	if len(words) == 0 {
		return ""
	}

	var out strings.Builder
	for _, word := range words {
		if acronym, ok := knownAcronyms[strings.ToLower(word)]; ok {
			out.WriteString(acronym)
			continue
		}

		out.WriteString(strcase.ToCamel(strings.ToLower(word)))
	}

	return out.String()
}

func splitIdentifierWords(sentence string) []string {
	runes := []rune(strings.TrimSpace(sentence))
	if len(runes) == 0 {
		return nil
	}

	words := make([]string, 0)
	current := make([]rune, 0, len(runes))
	flush := func() {
		if len(current) == 0 {
			return
		}
		words = append(words, string(current))
		current = current[:0]
	}

	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}

		if len(current) > 0 {
			prev := current[len(current)-1]
			var next rune
			hasNext := false
			for j := i + 1; j < len(runes); j++ {
				if unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) {
					next = runes[j]
					hasNext = true
					break
				}
				if !unicode.IsLetter(runes[j]) && !unicode.IsDigit(runes[j]) {
					break
				}
			}

			if shouldSplitIdentifier(prev, r, next, hasNext) {
				flush()
			}
		}

		current = append(current, r)
	}

	flush()
	return words
}

func shouldSplitIdentifier(prev, current, next rune, hasNext bool) bool {
	if unicode.IsDigit(prev) && unicode.IsLetter(current) {
		return true
	}

	if unicode.IsLower(prev) && unicode.IsUpper(current) {
		return true
	}

	return unicode.IsUpper(prev) && unicode.IsUpper(current) && hasNext && unicode.IsLower(next)
}

// NamifyWithoutParams will convert a sentence to a golang conventional type name.
// and will remove all parameters that can appear between '{' and '}'.
func NamifyWithoutParams(sentence string) string {
	// Remove parameters
	re := regexp.MustCompile("{[^()]*}")
	sentence = string(re.ReplaceAll([]byte(sentence), []byte("_")))

	return namify(sentence)
}

// DefaultNamifier will convert a sentence to a golang conventional type name.
func DefaultNamifier(sentence string) string {
	// Check if empty
	if len(sentence) == 0 {
		return sentence
	}

	// Upper letters that are preceded with an underscore
	previous := '_'
	for i, r := range sentence {
		if !unicode.IsLetter(previous) && !unicode.IsDigit(previous) {
			sentence = sentence[:i] + strings.ToUpper(string(r)) + sentence[i+1:]
		}
		previous = r
	}

	// Remove everything except alphanumerics
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	sentence = string(re.ReplaceAll([]byte(sentence), []byte("")))

	// Remove leading numbers
	re = regexp.MustCompile("^[0-9]+")
	sentence = string(re.ReplaceAll([]byte(sentence), []byte("")))
	if len(sentence) == 0 {
		return sentence
	}

	// Upper first letter
	sentence = strings.ToUpper(sentence[:1]) + sentence[1:]

	return sentence
}

// ConvertKey is used in template to convert schema property key name
// according to chosen strategy.
func ConvertKey(sentence string) string {
	return convertKey(sentence)
}

// Namify is used in template to generated golang structs names
// according to chosen strategy.
func Namify(sentence string) string {
	return namify(sentence)
}

// SetConvertKeyFn sets the function used to convert schema property key names.
func SetConvertKeyFn(name string) error {
	fn, ok := convertKeyFuncs[name]
	if !ok {
		return fmt.Errorf("unknown convert key function %s, supported values: snake, kebab, camel, none", name)
	}

	convertKey = fn

	return nil
}

// SetNamifyFn sets the function used to generate golang struct names.
func SetNamifyFn(name string) error {
	fn, ok := namifyPerScheme[name]
	if !ok {
		return fmt.Errorf("unknown namify function %s, supported values: camel, none", name)
	}

	namify = fn

	return nil
}

// HasField will check if a struct has a field with the given name.
func HasField(v any, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}

// DescribeStruct will describe a struct in a human-readable way using `%+v`
// format from the standard library.
func DescribeStruct(st any) string {
	return MultiLineComment(fmt.Sprintf("%+v", st))
}

// MultiLineComment will prefix each line of a comment with "// " in order to
// make it a valid multiline golang comment.
func MultiLineComment(comment string) string {
	comment = strings.TrimSuffix(comment, "\n")
	return strings.ReplaceAll(comment, "\n", "\n// ")
}

// Args is a function used to pass arguments to templates.
func Args(vs ...any) []any {
	return vs
}

// CutSuffix is a function used to remove a suffix to a string.
func CutSuffix(s, suffix string) string {
	s, _ = strings.CutSuffix(s, suffix)
	s, _ = strings.CutSuffix(s, "_"+suffix)
	return s
}

var isDateOrDateTimeGenerated = func(format string) bool {
	return format == "date" || format == "date-time"
}

// DisableDateOrTimeGeneration is used to disable the generation of date/date-time formats within types.
func DisableDateOrTimeGeneration() {
	isDateOrDateTimeGenerated = func(_ string) bool { return false }
}

// HelpersFunctions returns the functions that can be used as helpers
// in a golang template.
func HelpersFunctions() template.FuncMap {
	return template.FuncMap{
		"namifyWithoutParam":        NamifyWithoutParams,
		"namify":                    Namify,
		"isDateOrDateTimeGenerated": isDateOrDateTimeGenerated,
		"convertKey":                ConvertKey,
		"snakeCase":                 strcase.ToSnake,
		"hasField":                  HasField,
		"describeStruct":            DescribeStruct,
		"multiLineComment":          MultiLineComment,
		"cutSuffix":                 CutSuffix,
		"args":                      Args,
	}
}
