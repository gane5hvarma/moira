package bleve

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/blevesearch/bleve"

	"github.com/moira-alert/moira"
	"github.com/moira-alert/moira/index/mapping"
)

// Search gets search params and returns triggerIDs in following order:
// TriggerCheck.Score (desc)
// Relevance (asc)
// Trigger.Name (asc)
func (index *TriggerIndex) Search(filterTags []string, searchString string, onlyErrors bool, page int64, size int64) (triggerSearchResults moira.TriggerSearchResults, total int64, err error) {
	if size < 0 {
		page = 0
		docs, _ := index.index.DocCount()
		size = int64(docs)
	}

	req := buildSearchRequest(filterTags, searchString, onlyErrors, int(page), int(size))

	searchResult, err := index.index.Search(req)
	if err != nil {
		return
	}
	total = int64(searchResult.Total)
	if searchResult.Hits.Len() == 0 {
		return
	}

	results := make([]moira.TriggerSearchResult, 0, total)

	for _, result := range searchResult.Hits {
		triggerRes := moira.TriggerSearchResult{
			TriggerID: result.ID,
		}
		if nameHighlights, ok := result.Fragments[mapping.TriggerName.String()]; ok {
			for _, fragment := range nameHighlights {
				triggerRes.HighLights.Name += fragment
			}
		}
		if descHighlights, ok := result.Fragments[mapping.TriggerDesc.String()]; ok {
			for _, fragment := range descHighlights {
				triggerRes.HighLights.Desc += fragment
			}
		}
		results = append(results, triggerRes)
	}

	return results, total, err
}

func buildSearchRequest(filterTags []string, searchString string, onlyErrors bool, page, size int) *bleve.SearchRequest {

	searchTerms := splitStringToTerms(searchString)
	searchQuery := buildSearchQuery(filterTags, searchTerms, onlyErrors)

	from := page * size
	req := bleve.NewSearchRequestOptions(searchQuery, size, from, false)
	// sorting order:
	// TriggerCheck.Score (desc)
	// Relevance (asc)
	// Trigger.Name (asc)
	req.SortBy([]string{fmt.Sprintf("-%s", mapping.TriggerLastCheckScore.String()), "_score", mapping.TriggerName.String()})
	req.Highlight = bleve.NewHighlight()

	return req
}

func splitStringToTerms(searchString string) (searchTerms []string) {
	searchString = escapeString(searchString)

	return strings.Fields(searchString)
}

func escapeString(original string) (escaped string) {
	return regexp.MustCompile(`[|+\-=&<>!(){}\[\]^"'~*?\\/.,:;_@]`).ReplaceAllString(original, " ")
}
