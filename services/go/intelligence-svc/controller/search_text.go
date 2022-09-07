package controller

import (
	"encoding/json"
	"github.com/lefinal/meh"
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
)

type searchTextGenerator func(contentRaw json.RawMessage) (nulls.String, error)

// newSearchTextGeneratorWithUnmarshal returns a searchTextGenerator that
// unmarshals the raw content and calls the given generator function on it.
func newSearchTextGeneratorWithUnmarshal[T any](gen func(content T) (nulls.String, error)) searchTextGenerator {
	return func(contentRaw json.RawMessage) (nulls.String, error) {
		var content T
		err := json.Unmarshal(contentRaw, &content)
		if err != nil {
			return nulls.String{}, meh.NewBadInputErrFromErr(err, "parse content", meh.Details{"raw": contentRaw})
		}
		return gen(content)
	}
}

// genSearchText generates the search-text for the given store.CreateIntel.
func (c *Controller) genSearchText(create store.CreateIntel) (nulls.String, error) {
	var generator searchTextGenerator
	switch create.Type {
	case store.IntelTypePlaintextMessage:
		generator = newSearchTextGeneratorWithUnmarshal(genSearchTextForPlaintextMessage)
	default:
		return nulls.String{}, meh.NewInternalErr("unsupported intel-type", meh.Details{"type": create.Type})
	}
	if generator == nil {
		return nulls.String{}, meh.NewInternalErr("no generator set", nil)
	}
	searchText, err := generator(create.Content)
	if err != nil {
		return nulls.String{}, meh.Wrap(err, "generator", meh.Details{"content_raw": create.Content})
	}
	return searchText, nil
}

// genSearchTextForPlaintextMessage generates the search-text for intel with
// store.IntelTypePlaintextMessage.
func genSearchTextForPlaintextMessage(content store.IntelTypePlaintextMessageContent) (nulls.String, error) {
	return nulls.NewString(content.Text), nil
}
