package permission

// CreateAnyAddressBookEntryPermissionName for CreateAnyAddressBookEntry.
const CreateAnyAddressBookEntryPermissionName Name = "logistics.address-book.entry.create.any"

// CreateAnyAddressBookEntry allows creation of address book entries, even if
// not for associated with the issuer.
func CreateAnyAddressBookEntry() Matcher {
	return Matcher{
		Name: "create-any-address-book-entry",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[CreateAnyAddressBookEntryPermissionName]
			return ok, nil
		},
	}
}

// UpdateAnyAddressBookEntryPermissionName for UpdateAnyAddressBookEntry.
const UpdateAnyAddressBookEntryPermissionName Name = "logistics.address-book.entry.update.any"

// UpdateAnyAddressBookEntry allows creation of address book entries, even if
// not for associated with the issuer.
func UpdateAnyAddressBookEntry() Matcher {
	return Matcher{
		Name: "update-any-address-book-entry",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[UpdateAnyAddressBookEntryPermissionName]
			return ok, nil
		},
	}
}

// DeleteAnyAddressBookEntryPermissionName for DeleteAnyAddressBookEntry.
const DeleteAnyAddressBookEntryPermissionName Name = "logistics.address-book.entry.delete.any"

// DeleteAnyAddressBookEntry allows creation of address book entries, even if
// not for associated with the issuer.
func DeleteAnyAddressBookEntry() Matcher {
	return Matcher{
		Name: "delete-any-address-book-entry",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[DeleteAnyAddressBookEntryPermissionName]
			return ok, nil
		},
	}
}

// ViewAnyAddressBookEntryPermissionName for ViewAnyAddressBookEntry.
const ViewAnyAddressBookEntryPermissionName Name = "logistics.address-book.entry.view.any"

// ViewAnyAddressBookEntry allows creation of address book entries, even if not
// for associated with the issuer.
func ViewAnyAddressBookEntry() Matcher {
	return Matcher{
		Name: "view-any-address-book-entry",
		MatchFn: func(granted map[Name]Permission) (bool, error) {
			_, ok := granted[ViewAnyAddressBookEntryPermissionName]
			return ok, nil
		},
	}
}
