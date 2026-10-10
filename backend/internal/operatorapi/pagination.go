package operatorapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/starterapp/backend/internal/pagination"
)

// PageParams parses ?page= and ?per_page= for a list handler. On invalid input
// it writes the 422 validation error (naming the field) and returns false; the
// handler must return immediately in that case:
//
//	p, ok := operatorapi.PageParams(c)
//	if !ok {
//		return
//	}
//	rows, total := listThings(ctx, pool, p.Limit(), p.Offset())
//	c.JSON(http.StatusOK, pagination.New(rows, p, total))
func PageParams(c *gin.Context) (pagination.Params, bool) {
	p, err := pagination.Parse(c.Request.URL.Query())
	if err != nil {
		var perr *pagination.Error
		if errors.As(err, &perr) {
			writeFieldError(c, http.StatusUnprocessableEntity, "validation_error", perr.Message, perr.Field)
		} else {
			writeError(c, http.StatusBadRequest, "malformed_request", "invalid pagination parameters")
		}
		return pagination.Params{}, false
	}
	return p, true
}
