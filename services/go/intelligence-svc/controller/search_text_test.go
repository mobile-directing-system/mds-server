package controller

import (
	"github.com/lefinal/nulls"
	"github.com/mobile-directing-system/mds-server/services/go/intelligence-svc/store"
	"github.com/mobile-directing-system/mds-server/services/go/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestController_genSearchText(t *testing.T) {
	c := NewMockController()
	unsupportedIntelTypeErrMessage := "unsupported intel-type"
	t.Run("TestAssureUnsupportedIntelError", func(t *testing.T) {
		unknownIntelType := store.IntelType(testutil.NewUUIDV4().String())
		_, err := c.Ctrl.genSearchText(store.CreateIntel{Type: unknownIntelType})
		require.Error(t, err, "should fail")
		assert.Contains(t, err.Error(), unsupportedIntelTypeErrMessage, "should return correct error message")
	})
	testutil.TestMapperWithConstExtraction(t, func(from store.IntelType) (string, error) {
		_, err := c.Ctrl.genSearchText(store.CreateIntel{Type: from})
		if err == nil {
			return "", nil
		}
		if !strings.Contains(err.Error(), unsupportedIntelTypeErrMessage) {
			return "", nil
		}
		return "", err
	}, "../store/intel_content.go", nulls.String{})
}

func Test_genSearchTextForPlaintextMessage(t *testing.T) {
	content := store.IntelTypePlaintextMessageContent{
		Text: "bone",
	}
	searchText, err := genSearchTextForPlaintextMessage(content)
	require.NoError(t, err, "should not fail")
	require.Truef(t, searchText.Valid, "should return search text")
	assert.Contains(t, searchText.String, content.Text, "should contain the text")
}

func Test_genSearchTextForAnalogRadioMessage(t *testing.T) {
	content := store.IntelTypeAnalogRadioMessageContent{
		Channel:  "person",
		Callsign: "chance",
		Head:     "cottage",
		Content:  "ahead",
	}
	searchText, err := genSearchTextForAnalogRadioMessage(content)
	require.NoError(t, err, "should not fail")
	require.Truef(t, searchText.Valid, "should return search text")
	assert.Contains(t, searchText.String, content.Callsign, "should contain callsign")
	assert.Contains(t, searchText.String, content.Content, "should contain content")
}
