//go:build sdkcodegen
// +build sdkcodegen

package main

import (
	"errors"
	"log"
	"strings"

	"github.com/russross/blackfriday/v2"
)

var (
	errToplevelTopicNotH1 = errors.New("the top-level topic must be h1")
	errTopicChildNotH2    = errors.New("the children sections of a topic must be h2")
	errUnknownTopicChild  = errors.New("unknown child section of topic")
	errModelDefNotH3      = errors.New("model definition header must be h3")

	errMultipleModelTables    = errors.New("only one table allowed per model")
	errUnknownFieldTableTitle = errors.New("unknown column title of field table")

	errUnknownAPICallTableTitle = errors.New("unknown column title of api call table")
	errInvalidAPICallURLSpec    = errors.New("invalid API call URL spec")
	errUnknownAPICallHTTPMethod = errors.New("unknown HTTP method for API call")

	errUnknownBooleanSpec = errors.New("unknown text for boolean value")
)

func analyzeDocument(doc *mdTocNode) (result hir, err error) {

	for _, n := range doc.TocChildren {
		var topic topic
		topic, err = analyzeH1(n)
		if err != nil {
			return
		}
		result.topics = append(result.topics, topic)
	}

	return
}

func analyzeH1(doc *mdTocNode) (topic, error) {
	empty := topic{}

	if doc.Level != 1 {
		return empty, errToplevelTopicNotH1
	}

	result := topic{}
	for _, n := range doc.TocChildren {
		// log.Printf("H2: %s\n", n.ThisText())
		title := n.ThisInnerText()
		title = strings.TrimSpace(title)

		switch title {
		case "Imports":
			imports, err := analyzeH2Imports(n)
			if err != nil {
				return empty, err
			}
			result.imports = imports

		case "Models":
			models, err := analyzeH2Models(n)
			if err != nil {
				return empty, err
			}
			result.models = models

		case "API calls":
			calls, err := analyzeH2Calls(n)
			if err != nil {
				return empty, err
			}
			result.calls = calls

		default:
			return empty, errUnknownTopicChild
		}
	}
	return result, nil
}

func analyzeH2Imports(doc *mdTocNode) ([]string, error) {
	if doc.Level != 2 {
		return nil, errTopicChildNotH2
	}
	log.Printf("imports: %s lvl: %d, %+v, %+v\n", doc.ThisType(), doc.Level, doc.TocChildren, doc.Content)
	var names []string
	for _, sub := range doc.Content {
		log.Printf("sub: %s c%+v tc%+v\n", sub.ThisType(), sub.Content, sub.ThisContent)
		for _, item := range sub.ThisContent {
			log.Printf("item: %s c%+v tc%+v\n", item.ThisType(), item.Content, item.ThisContent)
			name := item.ThisInnerText()
			names = append(names, name)
		}
	}
	return names, nil
}

func analyzeH2Models(doc *mdTocNode) ([]apiModel, error) {
	if doc.Level != 2 {
		return nil, errTopicChildNotH2
	}

	result := make([]apiModel, 0)
	for _, n := range doc.TocChildren {
		model, err := analyzeH3Model(n)
		if err != nil {
			return nil, err
		}
		result = append(result, model)
	}
	return result, nil
}

func analyzeH3Model(doc *mdTocNode) (apiModel, error) {
	empty := apiModel{}

	if doc.Level != 3 {
		return empty, errModelDefNotH3
	}

	result := apiModel{
		vis:              visibilityPublic,
		preCodeSections:  make(map[string][]string),
		postCodeSections: make(map[string][]string),
	}

	// model metadata
	var docSb strings.Builder
	for _, n := range doc.ThisContent {
		switch n.ThisType() {
		case blackfriday.Code:
			// ident
			result.ident = n.ThisLit()

		case blackfriday.Text:
			docSb.WriteString(n.ThisLit())

		default:
			// ignore
		}
	}
	result.doc = strings.TrimSpace(docSb.String())

	// model fields
	// currently only one table is allowed
	seenTable := false
	for _, n := range doc.Content {
		switch n.ThisType() {
		case blackfriday.Table:
			if seenTable {
				return empty, errMultipleModelTables
			}
			seenTable = true

			fields, err := analyzeModelFieldTable(n)
			if err != nil {
				return empty, err
			}
			result.fields = fields

		case blackfriday.CodeBlock:
			lang := string(n.This.CodeBlockData.Info)
			if !seenTable {
				result.preCodeSections[lang] = append(result.preCodeSections[lang], string(n.This.Literal))
			} else {
				result.postCodeSections[lang] = append(result.postCodeSections[lang], string(n.This.Literal))
			}

		default:
			// ignore for now
			// TODO: allow collecting paragraphs into doc comments
		}
	}
	return result, nil
}

func analyzeModelFieldTable(tbl *mdContentNode) ([]apiModelField, error) {
	// initially mark the columns as non-existent
	idxIdent := -1
	idxType := -1
	idxDesc := -1
	idxTagBSON := -1
	idxTagJSON := -1
	idxTagXML := -1

	result := make([]apiModelField, 0)

	// TODO: disallow multiple header rows
	for _, n := range tbl.ThisContent {
		switch n.ThisType() {
		case blackfriday.TableHead:
			// only look at the first row
			tr := n.ThisContent[0]

			// parse out the column titles
			for i, td := range tr.ThisContent {
				colTitle := strings.ToLower(td.ThisInnerText())
				switch colTitle {
				case "name":
					idxIdent = i
				case "type":
					idxType = i
				case "doc":
					idxDesc = i
				case "bson":
					idxTagBSON = i
				case "json":
					idxTagJSON = i
				case "xml":
					idxTagXML = i
				default:
					return nil, errUnknownFieldTableTitle
				}
			}

		case blackfriday.TableBody:
			// parse the fields
			for _, tr := range n.ThisContent {
				field := apiModelField{
					vis: visibilityPublic,
				}

				isTODO := false
				for i, td := range tr.ThisContent {
					// skip any row that contains a TODO cell
					if td.ThisInnerText() == "TODO" {
						isTODO = true
						break
					}

					if i == idxIdent {
						if lit, ok := td.FirstCodeLit(); ok {
							field.ident = lit
						}
					}

					if i == idxType {
						if lit, ok := td.FirstCodeLit(); ok {
							field.typ = lit
						}
					}

					if i == idxDesc {
						field.doc = td.ThisInnerText()
					}

					if i == idxTagBSON {
						if lit, ok := td.FirstCodeLit(); ok {
							field.tags = append(field.tags, fieldTag{"bson", lit})
						}
					}

					if i == idxTagJSON {
						if lit, ok := td.FirstCodeLit(); ok {
							field.tags = append(field.tags, fieldTag{"json", lit})
						}
					}

					if i == idxTagXML {
						if lit, ok := td.FirstCodeLit(); ok {
							field.tags = append(field.tags, fieldTag{"xml", lit})
						}
					}
				}

				if isTODO {
					continue
				}

				result = append(result, field)
			}
		}
	}

	return result, nil
}

func analyzeH2Calls(doc *mdTocNode) ([]apiCall, error) {
	if doc.Level != 2 {
		return nil, errTopicChildNotH2
	}

	result := []apiCall{}
	for _, n := range doc.Content {
		switch n.ThisType() {
		case blackfriday.Table:
			calls, err := analyzeAPICallsTable(n)
			if err != nil {
				return nil, err
			}
			result = append(result, calls...)

		default:
			// ignored
			// TODO: allow inline code snippets here too
		}
	}

	return result, nil
}

func analyzeAPICallsTable(tbl *mdContentNode) ([]apiCall, error) {
	// initially mark the columns as non-existent
	idxIdent := -1
	idxReqType := -1
	idxRespType := -1
	idxURL := -1
	idxAK := -1
	idxDoc := -1

	result := make([]apiCall, 0)

	// TODO: disallow multiple header rows
	for _, n := range tbl.ThisContent {
		switch n.ThisType() {
		case blackfriday.TableHead:
			// only look at the first row
			tr := n.ThisContent[0]

			// parse out the column titles
			for i, td := range tr.ThisContent {
				colTitle := strings.ToLower(td.ThisInnerText())
				switch colTitle {
				case "name":
					idxIdent = i
				case "request type":
					idxReqType = i
				case "response type":
					idxRespType = i
				case "url":
					idxURL = i
				case "access token":
					idxAK = i
				case "doc":
					idxDoc = i
				default:
					return nil, errUnknownAPICallTableTitle
				}
			}

		case blackfriday.TableBody:
			for _, tr := range n.ThisContent {
				row := apiCallRow{}

				isTODO := false
				for i, td := range tr.ThisContent {
					// skip any row that contains a TODO cell
					if td.ThisInnerText() == "TODO" {
						isTODO = true
						break
					}

					if i == idxIdent {
						if lit, ok := td.FirstCodeLit(); ok {
							row.ident = lit
						}
					}

					if i == idxReqType {
						if lit, ok := td.FirstCodeLit(); ok {
							row.reqType = lit
						}
					}

					if i == idxRespType {
						if lit, ok := td.FirstCodeLit(); ok {
							row.respType = lit
						}
					}

					if i == idxURL {
						if lit, ok := td.FirstCodeLit(); ok {
							row.urlSpec = lit
						}
					}

					if i == idxAK {
						row.akSpec = td.ThisInnerText()
					}

					if i == idxDoc {
						row.doc = td.ThisInnerText()
					}
				}

				if isTODO {
					continue
				}

				call, err := parseAPICallRow(row)
				if err != nil {
					return nil, err
				}

				result = append(result, call)
			}
		}
	}

	return result, nil
}

type apiCallRow struct {
	ident    string
	doc      string
	reqType  string
	respType string
	urlSpec  string
	akSpec   string
}

func parseAPICallRow(x apiCallRow) (apiCall, error) {
	empty := apiCall{}

	urlSpecParts := strings.Split(x.urlSpec, " ")
	if len(urlSpecParts) != 2 {
		return empty, errInvalidAPICallURLSpec
	}

	httpMeth := urlSpecParts[0]
	url := urlSpecParts[1]

	meth, err := parseAPIMethod(httpMeth)
	if err != nil {
		return empty, err
	}

	ak, err := parseBool(x.akSpec)
	if err != nil {
		return empty, err
	}

	return apiCall{
		ident: x.ident,
		doc:   x.doc,
		vis:   visibilityPrivate,

		reqType:  x.reqType,
		respType: x.respType,

		needsAccessToken: ak,

		method:  meth,
		httpURI: url,
	}, nil
}

func parseBool(x string) (bool, error) {
	switch strings.ToLower(x) {
	case "y", "yes", "+":
		return true, nil
	case "n", "no", "-":
		return false, nil
	default:
		return false, errUnknownBooleanSpec
	}
}

func parseAPIMethod(x string) (apiMethod, error) {
	switch x {
	case "GET":
		return apiMethodGET, nil
	case "POST":
		return apiMethodPOSTJSON, nil
	case "POST(media)":
		return apiMethodPOSTMedia, nil
	default:
		return apiMethodUnknown, errUnknownAPICallHTTPMethod
	}
}
