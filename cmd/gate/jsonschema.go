package main

// A reader for the JSON Schema documents this repository holds.
//
// The component record schema says what a record must carry, and until this
// existed nothing read it. A record with a missing field, a value outside its
// stated bound or a spectrum with a band short would have been caught by whoever
// reviewed the diff and by nothing else, which is the state issue #38 exists to
// leave.
//
// Why a reader here rather than a library. The check has to run from a clean
// checkout with no network, and the module graph of this repository is empty:
//
//	go list -m all
//
// Taking a dependency is a licence question this repository cannot answer yet,
// because it has no licence to judge compatibility against. Writing the reader
// costs what is below and it is bounded by the schema it has to read, which now
// exists, so the keywords are read off a document rather than guessed at.
//
// The property that makes that bound safe is the refusal. A keyword this reader
// does not implement is a defect in the reader and is reported as one, rather
// than skipped. A validator that ignores what it does not understand passes
// every record the ignored keyword would have refused, and it passes them
// silently, which is worse than having no validator at all: the run is green and
// the schema reads as enforced.
//
// What it does not do, said here because a green run says nothing about any of
// it. It resolves no reference outside the schema document, so `$schema` and
// `$id` are read as annotations and nothing is fetched. It judges no format
// beyond what `pattern` states. It compares numbers as float64, so a bound and a
// value that differ beyond that precision compare equal.

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// A problem is one thing wrong, with where it was found.
//
// Instance is the path to the field inside the document being judged, so a
// message can name the field rather than a byte offset into a file somebody has
// to count into. The empty path is the document itself.
//
// Written without a leading separator, which is a departure from the JSON
// pointer this would otherwise be. A leading slash makes every one of these a
// string an absolute path, and the suite rule that refuses a fixture path
// leaving its package reads them as exactly that. The field path is what a
// person correcting a record wants to read anyway.
type problem struct {
	Instance string
	Message  string
}

func (p problem) String() string {
	where := p.Instance
	if where == "" {
		where = "the record"
	}
	return where + " " + p.Message
}

// keywords is every keyword this reader implements, with what it does with it.
//
// The map is written out rather than derived because it is the reader's own
// declaration of what it covers, and it is what schemaDefects compares a schema
// against. A keyword absent from here is refused by name when a schema uses it.
var keywords = map[string]string{
	// Read and ignored. Each says something to a person and nothing to a
	// validator, and `unit` is this repository's own annotation carrying the SI
	// unit of a numeric field.
	"$schema":     "annotation",
	"$id":         "annotation",
	"$comment":    "annotation",
	"title":       "annotation",
	"description": "annotation",
	"unit":        "annotation",
	"$defs":       "annotation",

	// Applied to the instance or to part of it.
	"$ref":                  "applicator",
	"allOf":                 "applicator",
	"anyOf":                 "applicator",
	"oneOf":                 "applicator",
	"if":                    "applicator",
	"then":                  "applicator",
	"else":                  "applicator",
	"properties":            "applicator",
	"items":                 "applicator",
	"additionalProperties":  "applicator",
	"propertyNames":         "applicator",
	"unevaluatedProperties": "applicator",

	// Judged directly.
	"type":              "assertion",
	"enum":              "assertion",
	"const":             "assertion",
	"required":          "assertion",
	"dependentRequired": "assertion",
	"minLength":         "assertion",
	"pattern":           "assertion",
	"minimum":           "assertion",
	"maximum":           "assertion",
	"exclusiveMinimum":  "assertion",
	"minItems":          "assertion",
	"minProperties":     "assertion",
}

// maxDepth stops a schema that refers to itself from taking the run down with
// it. The schemas here are acyclic and this is the guard for the one that is
// not, so a bad schema is a failure with a message rather than a hang.
const maxDepth = 64

// A schema document, ready to judge instances against.
type schema struct {
	// Path is where the document came from, for a message.
	Path string
	// root is the whole document, so a reference can be resolved inside it.
	root any
}

// readSchema parses a schema document and refuses one this reader cannot judge
// against, before any record is read.
//
// The refusal comes first deliberately. A schema keyword this reader skips would
// otherwise be discovered by a record that should have been refused and was not,
// which is the failure with no symptom.
func readSchema(path string, src []byte) (*schema, error) {
	var root any
	if err := json.Unmarshal(src, &root); err != nil {
		return nil, fmt.Errorf("%s is not JSON: %w", path, err)
	}
	s := &schema{Path: path, root: root}
	if defects := s.defects(); len(defects) > 0 {
		var b strings.Builder
		for _, d := range defects {
			fmt.Fprintf(&b, "\n  %s", d)
		}
		return nil, fmt.Errorf("%s uses %d thing(s) this repository's reader does not implement, so a record it would refuse would pass:%s", path, len(defects), b.String())
	}
	return s, nil
}

// defects is every place the schema asks for something this reader does not do.
func (s *schema) defects() []problem {
	var found []problem
	s.walk(s.root, "", &found)
	sort.Slice(found, func(i, j int) bool { return found[i].Instance < found[j].Instance })
	return found
}

// walk descends a schema document looking for unimplemented keywords.
//
// It descends through `$defs` and through every applicator's subschemas. It does
// not try to tell a schema from a plain object inside one, because everywhere it
// descends is a place a schema may be, and a false report here is a loud failure
// somebody reads rather than a quiet pass.
func (s *schema) walk(node any, at string, found *[]problem) {
	obj, ok := node.(map[string]any)
	if !ok {
		return
	}
	for key, sub := range obj {
		here := at + "/" + key
		if _, known := keywords[key]; !known {
			*found = append(*found, problem{Instance: at, Message: fmt.Sprintf("uses the keyword %q, which this reader does not implement", key)})
			continue
		}
		switch key {
		case "$defs", "properties", "dependentRequired":
			// A map of names to subschemas, or in the last case to lists of
			// names. Descending into a list finds nothing and costs nothing.
			if m, ok := sub.(map[string]any); ok {
				for name, child := range m {
					s.walk(child, here+"/"+name, found)
				}
			}
		case "allOf", "anyOf", "oneOf":
			if list, ok := sub.([]any); ok {
				for i, child := range list {
					s.walk(child, here+"/"+strconv.Itoa(i), found)
				}
			}
		case "$ref":
			ref, ok := sub.(string)
			if !ok {
				*found = append(*found, problem{Instance: here, Message: "is not a string"})
				continue
			}
			if _, err := s.resolve(ref); err != nil {
				*found = append(*found, problem{Instance: here, Message: err.Error()})
			}
		case "if", "then", "else", "items", "additionalProperties",
			"propertyNames", "unevaluatedProperties":
			s.walk(sub, here, found)
		case "pattern":
			p, ok := sub.(string)
			if !ok {
				*found = append(*found, problem{Instance: here, Message: "is not a string"})
				continue
			}
			// Go's regular expressions are RE2 and a schema's are ECMA-262. A
			// pattern using something RE2 has no form of is refused here rather
			// than evaluated to whatever Go makes of it.
			if _, err := regexp.Compile(p); err != nil {
				*found = append(*found, problem{Instance: here, Message: fmt.Sprintf("is a pattern this reader cannot compile: %v", err)})
			}
		}
	}
}

// resolve follows a reference inside this document.
//
// Only a reference into `$defs` of the same document, and the document itself.
// Anything else names a place outside this file, and following it would mean
// either a fetch, which the no-network condition refuses, or a second document
// this reader was not given.
func (s *schema) resolve(ref string) (any, error) {
	if ref == "#" {
		return s.root, nil
	}
	const prefix = "#/$defs/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, fmt.Errorf("points at %q, and this reader follows only a reference into this document's own $defs", ref)
	}
	name := strings.TrimPrefix(ref, prefix)
	root, ok := s.root.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot be resolved, because the schema document is not an object")
	}
	defs, ok := root["$defs"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("points at %q and this document has no $defs", ref)
	}
	target, ok := defs[name]
	if !ok {
		return nil, fmt.Errorf("points at %q, which this document does not define", ref)
	}
	return target, nil
}

// Validate judges one document against this schema and returns everything wrong
// with it.
//
// Every problem rather than the first, because a record is corrected by a person
// reading a list, and a validator that stops at the first field sends them
// through the loop once per field.
func (s *schema) Validate(src []byte) ([]problem, error) {
	var instance any
	if err := json.Unmarshal(src, &instance); err != nil {
		return nil, fmt.Errorf("is not JSON: %w", err)
	}
	found, _ := s.apply(s.root, instance, "", 0)
	sort.SliceStable(found, func(i, j int) bool { return found[i].Instance < found[j].Instance })
	return found, nil
}

// apply judges one instance against one subschema.
//
// It returns what is wrong and, where nothing is, which of the instance's
// properties this subschema and everything it applied reached. That second
// return is what `unevaluatedProperties` is decided from, and it is why the
// applicators below merge a subschema's answer rather than only its verdict.
//
// A subschema that failed contributes no reached properties. That is the rule
// the specification states and it also keeps the messages honest: a branch that
// did not match cannot be the reason a property counts as covered.
func (s *schema) apply(sub any, instance any, at string, depth int) ([]problem, map[string]bool) {
	reached := map[string]bool{}
	if depth > maxDepth {
		return []problem{{Instance: at, Message: fmt.Sprintf("is nested deeper than %d schemas, which is a schema that refers to itself", maxDepth)}}, reached
	}

	switch shape := sub.(type) {
	case bool:
		if shape {
			return nil, reached
		}
		return []problem{{Instance: at, Message: "is present, and the schema allows nothing here"}}, reached
	case map[string]any:
		return s.applyObject(shape, instance, at, depth, reached)
	default:
		return []problem{{Instance: at, Message: "is judged against something that is neither a schema object nor true or false"}}, reached
	}
}

func (s *schema) applyObject(sub map[string]any, instance any, at string, depth int, reached map[string]bool) ([]problem, map[string]bool) {
	var found []problem

	merge := func(from map[string]bool) {
		for k := range from {
			reached[k] = true
		}
	}
	// branch applies a subschema and merges what it reached only where it
	// passed.
	branch := func(node any, inst any, where string) []problem {
		probs, got := s.apply(node, inst, where, depth+1)
		if len(probs) == 0 {
			merge(got)
		}
		return probs
	}

	if ref, ok := sub["$ref"].(string); ok {
		target, err := s.resolve(ref)
		if err != nil {
			found = append(found, problem{Instance: at, Message: "is judged against a reference that " + err.Error()})
		} else {
			found = append(found, branch(target, instance, at)...)
		}
	}

	found = append(found, s.assertions(sub, instance, at)...)

	for _, key := range []string{"allOf"} {
		if list, ok := sub[key].([]any); ok {
			for _, node := range list {
				found = append(found, branch(node, instance, at)...)
			}
		}
	}

	if list, ok := sub["anyOf"].([]any); ok {
		matched := 0
		var why []string
		for i, node := range list {
			if probs := branch(node, instance, at); len(probs) == 0 {
				matched++
			} else {
				why = append(why, fmt.Sprintf("alternative %d: %s", i+1, probs[0]))
			}
		}
		if matched == 0 {
			found = append(found, problem{Instance: at, Message: "matches none of the alternatives the schema allows: " + strings.Join(why, "; ")})
		}
	}

	if list, ok := sub["oneOf"].([]any); ok {
		matched := 0
		var why []string
		var passed map[string]bool
		for i, node := range list {
			probs, got := s.apply(node, instance, at, depth+1)
			if len(probs) == 0 {
				matched++
				passed = got
			} else {
				why = append(why, fmt.Sprintf("alternative %d: %s", i+1, probs[0]))
			}
		}
		switch {
		case matched == 1:
			merge(passed)
		case matched == 0:
			found = append(found, problem{Instance: at, Message: "matches none of the alternatives the schema allows: " + strings.Join(why, "; ")})
		default:
			found = append(found, problem{Instance: at, Message: fmt.Sprintf("matches %d of the alternatives the schema allows, and exactly one is allowed", matched)})
		}
	}

	if cond, ok := sub["if"]; ok {
		probs, got := s.apply(cond, instance, at, depth+1)
		if len(probs) == 0 {
			merge(got)
			if then, ok := sub["then"]; ok {
				found = append(found, branch(then, instance, at)...)
			}
		} else if otherwise, ok := sub["else"]; ok {
			found = append(found, branch(otherwise, instance, at)...)
		}
	}

	obj, isObject := instance.(map[string]any)
	if isObject {
		declared := map[string]bool{}
		if props, ok := sub["properties"].(map[string]any); ok {
			for name, node := range props {
				declared[name] = true
				value, present := obj[name]
				if !present {
					continue
				}
				probs := s.applyProperty(node, value, at, name, depth)
				found = append(found, probs...)
				if len(probs) == 0 {
					reached[name] = true
				}
			}
		}
		if more, ok := sub["additionalProperties"]; ok {
			for _, name := range sortedKeys(obj) {
				if declared[name] {
					continue
				}
				probs := s.applyProperty(more, obj[name], at, name, depth)
				found = append(found, probs...)
				if len(probs) == 0 {
					reached[name] = true
				}
			}
		}
		if names, ok := sub["propertyNames"]; ok {
			for _, name := range sortedKeys(obj) {
				if probs, _ := s.apply(names, name, below(at, name), depth+1); len(probs) > 0 {
					found = append(found, problem{Instance: at, Message: fmt.Sprintf("carries the name %q, which the schema does not allow here", name)})
				}
			}
		}
		// Last, because it is decided from everything the applicators above
		// reached, including the ones inside `$ref`, `allOf` and `then`.
		if unevaluated, ok := sub["unevaluatedProperties"]; ok && len(found) == 0 {
			for _, name := range sortedKeys(obj) {
				if reached[name] {
					continue
				}
				probs := s.applyProperty(unevaluated, obj[name], at, name, depth)
				found = append(found, probs...)
				if len(probs) == 0 {
					reached[name] = true
				}
			}
		}
	}

	if list, ok := instance.([]any); ok {
		if node, ok := sub["items"]; ok {
			for i, item := range list {
				probs, _ := s.apply(node, item, below(at, strconv.Itoa(i)), depth+1)
				found = append(found, probs...)
			}
		}
	}

	return found, reached
}

// applyProperty judges one property of an object and reports a `false` subschema
// as the sentence a reader needs, which is that this field may not be here at
// all rather than that something inside it is wrong.
func (s *schema) applyProperty(node any, value any, at, name string, depth int) []problem {
	where := below(at, name)
	if allowed, ok := node.(bool); ok && !allowed {
		return []problem{{Instance: at, Message: fmt.Sprintf("carries %q, which the schema does not allow on a record of this shape", name)}}
	}
	probs, _ := s.apply(node, value, where, depth+1)
	return probs
}

// assertions judges the keywords that look at the instance directly.
func (s *schema) assertions(sub map[string]any, instance any, at string) []problem {
	var found []problem
	add := func(format string, args ...any) {
		found = append(found, problem{Instance: at, Message: fmt.Sprintf(format, args...)})
	}

	if want, ok := sub["type"].(string); ok && !hasType(instance, want) {
		add("is %s, and the schema asks for %s", nameOfType(instance), want)
	}
	if want, ok := sub["const"]; ok && !sameJSON(instance, want) {
		add("is %s, and the schema allows only %s", asText(instance), asText(want))
	}
	if list, ok := sub["enum"].([]any); ok {
		matched := false
		var allowed []string
		for _, want := range list {
			allowed = append(allowed, asText(want))
			if sameJSON(instance, want) {
				matched = true
			}
		}
		if !matched {
			add("is %s, and the schema allows only %s", asText(instance), strings.Join(allowed, ", "))
		}
	}

	obj, isObject := instance.(map[string]any)
	if isObject {
		if list, ok := sub["required"].([]any); ok {
			for _, name := range list {
				field, ok := name.(string)
				if !ok {
					continue
				}
				if _, present := obj[field]; !present {
					add("is missing the required field %q", field)
				}
			}
		}
		if deps, ok := sub["dependentRequired"].(map[string]any); ok {
			for _, trigger := range sortedKeys(deps) {
				if _, present := obj[trigger]; !present {
					continue
				}
				list, ok := deps[trigger].([]any)
				if !ok {
					continue
				}
				for _, name := range list {
					field, ok := name.(string)
					if !ok {
						continue
					}
					if _, present := obj[field]; !present {
						add("carries %q, which the schema says requires %q, and that is missing", trigger, field)
					}
				}
			}
		}
		if want, ok := asFloat(sub["minProperties"]); ok && float64(len(obj)) < want {
			add("has %d field(s), and the schema asks for at least %s", len(obj), asNumber(want))
		}
	}

	if text, isText := instance.(string); isText {
		if want, ok := asFloat(sub["minLength"]); ok && float64(len([]rune(text))) < want {
			add("is %d character(s) long, and the schema asks for at least %s", len([]rune(text)), asNumber(want))
		}
		if p, ok := sub["pattern"].(string); ok {
			re, err := regexp.Compile(p)
			switch {
			case err != nil:
				add("is judged against a pattern this reader cannot compile: %v", err)
			case !re.MatchString(text):
				add("is %q, which does not match the pattern %s the schema asks for", text, p)
			}
		}
	}

	if number, isNumber := asFloat(instance); isNumber {
		if want, ok := asFloat(sub["minimum"]); ok && number < want {
			add("is %s, and the schema asks for at least %s", asNumber(number), asNumber(want))
		}
		if want, ok := asFloat(sub["maximum"]); ok && number > want {
			add("is %s, and the schema asks for at most %s", asNumber(number), asNumber(want))
		}
		if want, ok := asFloat(sub["exclusiveMinimum"]); ok && number <= want {
			add("is %s, and the schema asks for more than %s", asNumber(number), asNumber(want))
		}
	}

	if list, isList := instance.([]any); isList {
		if want, ok := asFloat(sub["minItems"]); ok && float64(len(list)) < want {
			add("has %d entry/entries, and the schema asks for at least %s", len(list), asNumber(want))
		}
	}

	return found
}

// hasType is the type test, with the one case that is not a Go type switch:
// a number with no fractional part is an integer, which is what the schema means
// by the word and is not what Go's decoder distinguishes.
func hasType(instance any, want string) bool {
	switch want {
	case "object":
		_, ok := instance.(map[string]any)
		return ok
	case "array":
		_, ok := instance.([]any)
		return ok
	case "string":
		_, ok := instance.(string)
		return ok
	case "boolean":
		_, ok := instance.(bool)
		return ok
	case "null":
		return instance == nil
	case "number":
		_, ok := instance.(float64)
		return ok
	case "integer":
		n, ok := instance.(float64)
		return ok && n == math.Trunc(n)
	default:
		return false
	}
}

func nameOfType(instance any) string {
	switch v := instance.(type) {
	case map[string]any:
		return "an object"
	case []any:
		return "an array"
	case string:
		return "a string"
	case bool:
		return "a boolean"
	case nil:
		return "null"
	case float64:
		if v == math.Trunc(v) {
			return "a whole number"
		}
		return "a number"
	default:
		return "of a shape this reader does not know"
	}
}

// sameJSON compares an instance against a schema's literal.
//
// Through the encoder rather than by reflection, so two objects with their keys
// written in different orders are not two different values, which is what a
// `const` on an object would otherwise mean.
func sameJSON(a, b any) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(left) == string(right)
}

func asFloat(value any) (float64, bool) {
	n, ok := value.(float64)
	return n, ok
}

// asNumber prints a bound the way the schema wrote it rather than in Go's
// default float spelling, so a message quoting a limit of 5000 does not say
// 5000.000000.
func asNumber(n float64) string {
	return strconv.FormatFloat(n, 'g', -1, 64)
}

func asText(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(b)
}

func sortedKeys(obj map[string]any) []string {
	names := make([]string, 0, len(obj))
	for name := range obj {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// below is the path to one field inside another, and the escaping that keeps it
// readable as one step. A field whose own name carries a separator would
// otherwise read as two levels of a record that has only one.
func below(at, name string) string {
	name = strings.ReplaceAll(name, "~", "~0")
	name = strings.ReplaceAll(name, "/", "~1")
	if at == "" {
		return name
	}
	return at + "/" + name
}
