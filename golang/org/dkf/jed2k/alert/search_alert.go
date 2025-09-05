package alert

import (
	"fmt"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// SearchResultAlert represents a search result notification
type SearchResultAlert struct {
	BaseAlert
	Results *protocol.SearchResult
	Query   string
}

// NewSearchResultAlert creates a new search result alert
func NewSearchResultAlert(results *protocol.SearchResult, query string) *SearchResultAlert {
	return &SearchResultAlert{
		BaseAlert: *NewBaseAlert(),
		Results:   results,
		Query:     query,
	}
}

// Severity returns the alert severity
func (sra *SearchResultAlert) Severity() Severity {
	return Info
}

// Category returns the alert category
func (sra *SearchResultAlert) Category() int {
	return int(StatusNotification)
}

// String returns string representation of the alert
func (sra *SearchResultAlert) String() string {
	return fmt.Sprintf("SearchResultAlert{query=%s, results=%d, more=%v}", 
		sra.Query, len(sra.Results.GetResults()), sra.Results.HasMoreResults())
}

// GetResults returns the search results
func (sra *SearchResultAlert) GetResults() []*protocol.SearchEntry {
	if sra.Results == nil {
		return nil
	}
	return sra.Results.GetResults()
}

// GetQuery returns the search query
func (sra *SearchResultAlert) GetQuery() string {
	return sra.Query
}

// HasMoreResults returns true if more results are available
func (sra *SearchResultAlert) HasMoreResults() bool {
	return sra.Results != nil && sra.Results.HasMoreResults()
}

// SearchFailedAlert represents a failed search notification
type SearchFailedAlert struct {
	BaseAlert
	Query string
	Error string
}

// NewSearchFailedAlert creates a new search failed alert
func NewSearchFailedAlert(query, error string) *SearchFailedAlert {
	return &SearchFailedAlert{
		BaseAlert: *NewBaseAlert(),
		Query:     query,
		Error:     error,
	}
}

// Severity returns the alert severity
func (sfa *SearchFailedAlert) Severity() Severity {
	return Warning
}

// Category returns the alert category
func (sfa *SearchFailedAlert) Category() int {
	return int(ErrorNotification)
}

// String returns string representation of the alert
func (sfa *SearchFailedAlert) String() string {
	return fmt.Sprintf("SearchFailedAlert{query=%s, error=%s}", sfa.Query, sfa.Error)
}

// GetQuery returns the search query
func (sfa *SearchFailedAlert) GetQuery() string {
	return sfa.Query
}

// GetError returns the error message
func (sfa *SearchFailedAlert) GetError() string {
	return sfa.Error
}