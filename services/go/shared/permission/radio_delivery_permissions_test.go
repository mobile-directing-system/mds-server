package permission

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestDeliverAnyRadioDelivery(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "deliver-any-radio-delivery",
		Matcher:     DeliverAnyRadioDelivery(),
		Granted:     DeliverAnyRadioDeliveryPermissionName,
		Others: []Name{
			UpdateAnyAddressBookEntryPermissionName,
			DeleteAnyAddressBookEntryPermissionName,
			CreateGroupPermissionName,
			ManageAnyRadioDeliveryPermissionName,
		},
	})
}

func TestManageAnyRadioDelivery(t *testing.T) {
	suite.Run(t, &NameMatcherSuite{
		MatcherName: "manage-any-radio-delivery",
		Matcher:     ManageAnyRadioDelivery(),
		Granted:     ManageAnyRadioDeliveryPermissionName,
		Others: []Name{
			UpdateAnyAddressBookEntryPermissionName,
			DeleteAnyAddressBookEntryPermissionName,
			DeliverAnyRadioDeliveryPermissionName,
			ViewAnyAddressBookEntryPermissionName,
		},
	})
}
