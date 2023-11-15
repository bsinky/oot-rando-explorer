package routes

import (
	"net/http"

	"github.com/bsinky/sohrando/authentication"
	"github.com/bsinky/sohrando/search"
	"github.com/bsinky/sohrando/util"
	"github.com/gin-gonic/gin"
)

func AddSearchRoutes(r *gin.Engine) {
	search := r.Group("/search")
	search.GET("/", searchPage)
	search.GET("/run", runSearch)
}

type SearchModel struct {
	util.ViewModel
	Filters map[string]*search.SearchFilter
}

func searchPage(c *gin.Context) {
	db := util.GetDatabase(c)
	user := authentication.GetCurrentUser(c)

	allFilters, err := search.AllFilters(db)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "search.html", SearchModel{
		ViewModel: util.ViewModel{
			User: user,
		},
		Filters: allFilters,
	})
}

type SearchResultModel struct {
	Result *search.Result
}

func runSearch(c *gin.Context) {
	db := util.GetDatabase(c)

	allFilters, err := search.AllFilters(db)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	reqFields := make(map[string]string)
	if err := c.BindQuery(reqFields); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var reqFilters []search.SearchFilterValue

	for k, v := range reqFields {
		// Remove invalid field names or blank values
		if fieldFilter, ok := allFilters[k]; !ok || v == "" {
			continue
		} else if !fieldFilter.IsValidOption(v) {
			continue
		}

		reqFilters = append(reqFilters, search.SearchFilterValue{
			FieldName: k,
			Value:     v,
		})
	}

	result, err := search.RunSearch(db, reqFilters)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "searchresult", SearchResultModel{
		Result: result,
	})
}
