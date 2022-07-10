package entityvalidation

import (
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"net/http"
)

// publicReport is the public representation of Report.
type publicReport struct {
	Errors []string `json:"errors"`
}

// toPublic creates the public representation of the Report.
func (report Report) toPublic() publicReport {
	return publicReport{
		Errors: report.Errors,
	}
}

// ValidateInRequest handles validation of a Validatable in a gin.Context. The
// first return value describes whether validation was ok. If not, the given
// error needs to returned as we expect it to be called in the context of
// httpendpoints.HandlerFunc, where the returned error also alters response
// behavior. This means, that if validation returns an error that led it fail,
// the error will be returned directly. If validation was not okay, validation
// errors will be sent to the client using the given gin.Context and a nil-error
// will be returned.
func ValidateInRequest(c *gin.Context, v Validatable) (bool, error) {
	report, err := v.Validate()
	if err != nil {
		return false, meh.Wrap(err, "validate", nil)
	}
	if report.IsOK() {
		return true, nil
	}
	// Respond details to client.
	c.JSON(http.StatusBadRequest, report.toPublic())
	return false, nil
}
