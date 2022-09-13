package store

import (
	"encoding/json"
	"fmt"
	"github.com/lefinal/meh"
	"github.com/mobile-directing-system/mds-server/services/go/shared/entityvalidation"
)

// IntelTypePlaintextMessage for simple plaintext messages.
const IntelTypePlaintextMessage IntelType = "plaintext-message"

// IntelTypePlaintextMessageContent is the content for intel with
// IntelTypePlaintextMessage.
type IntelTypePlaintextMessageContent struct {
	// Text is the actual text content.
	Text string `json:"text"`
}

// Validate assures that Text is not empty.
func (mc IntelTypePlaintextMessageContent) Validate() (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	if mc.Text == "" {
		report.AddError("text content must not be empty")
	}
	return report, nil
}

type intelContentValidator func(contentRaw json.RawMessage) (entityvalidation.Report, error)

func validateIntelContent[T entityvalidation.Validatable]() intelContentValidator {
	return func(contentRaw json.RawMessage) (entityvalidation.Report, error) {
		report := entityvalidation.NewReport()
		var content T
		err := json.Unmarshal(contentRaw, &content)
		if err != nil {
			report.AddError(fmt.Sprintf("invalid message content: %s", err.Error()))
		} else {
			subReport, err := content.Validate()
			if err != nil {
				return entityvalidation.Report{}, meh.Wrap(err, "validate content", nil)
			}
			report.Include(subReport)
		}
		return report, nil
	}
}

// validateCreateIntelTypeAndContent validates the given IntelType and content.
// Used in CreateIntel.Validate.
func validateCreateIntelTypeAndContent(intelType IntelType, contentRaw json.RawMessage) (entityvalidation.Report, error) {
	report := entityvalidation.NewReport()
	var contentValidator intelContentValidator
	switch intelType {
	case IntelTypePlaintextMessage:
		contentValidator = validateIntelContent[IntelTypePlaintextMessageContent]()
	default:
		report.AddError(fmt.Sprintf("unsupported intel-type: %v", intelType))
		return report, nil
	}
	if contentValidator == nil {
		return entityvalidation.Report{}, meh.NewInternalErr("no content validator set", nil)
	}
	subReport, err := contentValidator(contentRaw)
	if err != nil {
		return entityvalidation.Report{}, meh.Wrap(err, "content validator", meh.Details{"content": contentRaw})
	}
	report.Include(subReport)
	return report, nil
}
