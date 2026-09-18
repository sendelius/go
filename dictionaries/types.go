package dictionaries

type GetListConfig struct {
	SearchFields     []string
	SearchSimilarity bool
}

type SearchResult[T any] struct {
	Item  T
	Score float64
}

type SearchConfig[T any] struct {
	Fields     []string
	Similarity bool
	Limit      int
	Priority   func(T) float64
}

type DictionaryRequestGetList struct {
	Limit  int                            `json:"limit" validate:"omitempty,min=1,max=1000"`
	Search DictionaryRequestGetListSearch `json:"search" validate:"omitempty"`
}

type DictionaryRequestGetListSearch struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
