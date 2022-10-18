package permission

// DeliverAnyRadioDeliveryPermissionName for DeliverAnyRadioDelivery.
const DeliverAnyRadioDeliveryPermissionName Name = "radio-delivery.deliver.any"

// DeliverAnyRadioDelivery allows delivering any radio delivery for operations,
// the user is part of.
func DeliverAnyRadioDelivery() Matcher {
	return Matcher{
		Name: "deliver-any-radio-delivery",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[DeliverAnyRadioDeliveryPermissionName]
			return ok, nil
		},
	}
}

// ManageAnyRadioDeliveryPermissionName for ManageAnyRadioDelivery.
const ManageAnyRadioDeliveryPermissionName Name = "radio-delivery.manage.any"

// ManageAnyRadioDelivery allows releasing any radio delivery from operations
// being picked up by anybody.
func ManageAnyRadioDelivery() Matcher {
	return Matcher{
		Name: "manage-any-radio-delivery",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ManageAnyRadioDeliveryPermissionName]
			return ok, nil
		},
	}
}
