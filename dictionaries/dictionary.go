package dictionaries

import (
	"sort"
)

type Dictionary[T any] struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	Data    []T    `json:"data"`
}

type ListConfig struct {
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

type Request struct {
	Limit  int           `json:"limit" validate:"omitempty,min=1,max=1000"`
	Search RequestSearch `json:"search" validate:"omitempty"`
}

type RequestSearch struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (d *Dictionary[T]) Search(value string, config SearchConfig[T]) []SearchResult[T] {
	value = normalize(value)

	if value == "" {
		return nil
	}

	limit := config.Limit

	if limit <= 0 {
		limit = 100
	}

	if limit > 100 {
		limit = 100
	}

	results := make([]SearchResult[T], 0)

	for _, item := range d.Data {
		best := 0.0

		for _, field := range config.Fields {
			fieldValue := getField(item, field)

			var current float64

			if config.Similarity {
				current = similarityScore(
					fieldValue,
					value,
				)
			} else {
				current = score(
					fieldValue,
					value,
				)
			}

			if current > best {
				best = current
			}
		}

		if best > 0 {
			results = append(results, SearchResult[T]{
				Item:  item,
				Score: best,
			})
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}

		if config.Priority != nil {
			return config.Priority(results[i].Item) >
				config.Priority(results[j].Item)
		}

		return false
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}
