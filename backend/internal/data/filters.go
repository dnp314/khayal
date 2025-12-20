package data

import (
	"khayal/internal/validator"
	"math"
	"strings"
)

type MetaData struct {
	CurrentPage  int `json:"current_page, omitempty"`
	PageSize     int `json:"page_size, omitempty"`
	FirstPage    int `json:"first_page, omitempty"`
	LastPage     int `json:"last_page, omitempty"`
	TotalRecords int `json:"total_records, omitempty"`
}

type Filters struct {
	Page         int
	Pagesize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, f Filters) {

	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 10000000, "pagesize", "must be less than or equal to 10000000")
	v.Check(f.Pagesize > 0, "pagesize", "must be greater than 0")
	v.Check(f.Pagesize <= 100, "pagesize", "must be less than or equal to 100")

	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")

}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter:" + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.Pagesize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.Pagesize
}

func calculateMetaData(totalRecords, page, pageSize int) MetaData {

	if totalRecords == 0 {
		return MetaData{}
	}

	return MetaData{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}

}
