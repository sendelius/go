package infrastructure

type BulkResult struct {
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

type ListPagination struct {
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
	Last  int  `json:"last"`
	Next  *int `json:"next,omitempty"`
	Prev  *int `json:"prev,omitempty"`
}

type ListRequest struct {
	Limit   int               `json:"limit" validate:"omitempty,min=1,max=1000"`
	Page    int               `json:"page" validate:"omitempty,min=1"`
	Sort    ListSort          `json:"sort" validate:"omitempty"`
	Search  ListSearch        `json:"search" validate:"omitempty"`
	Filters map[string]string `json:"filters" validate:"omitempty"`
}

type ListSort struct {
	Key  string `json:"key"`
	Type string `json:"type" validate:"omitempty,oneof=asc desc"`
}

type ListSearch struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ListConfig struct {
	SearchFields     []string
	SearchSimilarity bool
	FilterFields     []string
	SortFields       []string
}

func defaultListConfig() *ListConfig {
	return &ListConfig{
		SearchFields:     []string{},
		SearchSimilarity: false,
		FilterFields:     []string{},
		SortFields:       []string{},
	}
}

func (r *ListRequest) SetQueryParam(key string, value any) {
	switch key {
	case "filters":
		if filters, ok := value.(map[string]string); ok {
			r.Filters = filters
		}
	case "search":
		if search, ok := value.(map[string]string); ok {
			r.Search.Key = search["key"]
			r.Search.Value = search["value"]
		}
	case "sort":
		if sort, ok := value.(map[string]string); ok {
			r.Sort.Key = sort["key"]
			r.Sort.Type = sort["type"]
		}
	}
}

func (r *ListRequest) SetFilter(key string, value string) {
	r.Filters[key] = value
}

func (r *ListRequest) SetSearch(key string, value string) {
	r.Search.Key = key
	r.Search.Value = value
}

func (r *ListRequest) SetSort(key string, value string) {
	r.Sort.Key = key
	r.Sort.Type = value
}
