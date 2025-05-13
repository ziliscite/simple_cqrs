package product

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"strings"
)

type Search struct {
	name     string // match a query on name
	category string // term filter on category

	minPrice *float64 // range filter
	maxPrice *float64

	page     int // pagination: page number (1-based)
	pageSize int // pagination: items per page

	sortField string // e.g. "price"
	sortAsc   bool   // true = asc, false = desc
}

func NewSearch() *Search {
	return &Search{
		page:     1,  // default page
		pageSize: 20, // default size
	}
}

func (s *Search) WithName(name string) *Search {
	s.name = name
	return s
}

func (s *Search) WithCategory(category string) *Search {
	s.category = category
	return s
}

func (s *Search) WithMinPrice(minPrice float64) *Search {
	s.minPrice = &minPrice
	return s
}

func (s *Search) WithMaxPrice(maxPrice float64) *Search {
	s.maxPrice = &maxPrice
	return s
}

func (s *Search) WithPage(page int) *Search {
	s.page = page
	return s
}

func (s *Search) WithPageSize(pageSize int) *Search {
	s.pageSize = pageSize
	return s
}

func (s *Search) WithSortField(sort string) *Search {
	s.sortField = sort
	return s
}

func (s *Search) WithSortAsc(asc bool) *Search {
	s.sortAsc = asc
	return s
}

func (s *Search) Name() string {
	return s.name
}

func (s *Search) Category() string {
	return s.category
}

// PriceRange returns (minPrice, maxPrice)
func (s *Search) PriceRange() (*float64, *float64) {
	return s.minPrice, s.maxPrice
}

func (s *Search) Page() int {
	return s.page
}

func (s *Search) PageSize() int {
	return s.pageSize
}

// Pagination returns (page, pageSize)
func (s *Search) Pagination() (int, int) {
	return s.Page(), s.PageSize()
}

// Offset returns offset
func (s *Search) Offset() int {
	return (s.page - 1) * s.pageSize
}

// SortBy returns (sortField, sortDir)
func (s *Search) SortBy() (string, string) {
	sortDir := "asc"
	if !s.sortAsc {
		sortDir = "desc"
	}

	return s.sortField, sortDir
}

// Key builds a consistent Redis key for a product search and return tags that can be used to invalidate the cache.
// E.g. "products:all|name=foo|cat=bar|sort=name-asc|page=1|size=20|min=10.00|max=20.00"
func (s *Search) Key() (string, []string) {
	var tags []string
	parts := []string{"products:all"}

	// normalized filters
	if name := normalize(s.Name()); name != "" {
		tags = append(tags, "tag:name:"+name)
		parts = append(parts, "name="+name)
	}
	if cat := normalize(s.Category()); cat != "" {
		tags = append(tags, "tag:category:"+cat)
		parts = append(parts, "cat="+cat)
	}
	if field, order := s.SortBy(); field != "" {
		tags = append(tags, "tag:sort:"+field+"-"+order) // e.g. "tag:sort:price-asc"
		parts = append(parts, fmt.Sprintf("sort=%s-%s", field, order))
	}

	// pagination
	page, size := s.Pagination()
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	tags = append(tags, "tag:paging:"+fmt.Sprintf("%d-%d", page, size)) // e.g. "tag:paging:1-20"
	parts = append(parts, fmt.Sprintf("page=%d", page), fmt.Sprintf("size=%d", size))

	// price range filters
	minPrice, maxPrice := s.PriceRange()
	if minPrice != nil {
		tags = append(tags, "tag:min:"+fmt.Sprintf("%.2f", *minPrice)) // e.g. "tag:min:10.00"
		parts = append(parts, fmt.Sprintf("min=%.2f", *minPrice))
	}
	if maxPrice != nil {
		tags = append(tags, "tag:max:"+fmt.Sprintf("%.2f", *maxPrice)) // e.g. "tag:max:20.00"
		parts = append(parts, fmt.Sprintf("max=%.2f", *maxPrice))
	}

	// join with '|'
	raw := strings.Join(parts, "|")
	if len(raw) > 128 {
		sum := sha1.Sum([]byte(raw))
		raw = fmt.Sprintf("%x", sum)
	}
	return raw, tags
}

func normalize(val string) string {
	v := strings.TrimSpace(strings.ToLower(val))
	return url.QueryEscape(v)
}
