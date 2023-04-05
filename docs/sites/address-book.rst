Address Book
############

The address book is part of logistics.
It allows addressing entities (users, groups, etc.) in order to forward intel to recipients.
Each address book entry internally uses so called channels, that represent a way of delivering information.
Channels hold priorities and minimum importance barriers in order to forward information as fast as possible
but still do not cause too many interruptions.

Address book entries can optionally be assigned to operations as well users.
This allows a single user to have multiple entries for different operations.
An example use-case might be varying radio callsigns.
Not associating entries with users allows addressing artifical entities.
Here, there might be an email address for leaders@test.com.
Now, there is no need for creating a user account only in order to allow forwarding to this email address.

As already mentioned, each user can have multiple address book entries.
Keep in mind, that per user, only one entry per operation is allowed (plus a global one).

Create entry
============

In order to create an entry, not being associated with the requesting user,
the :ref:`permission.logistics.address-book.entry.create.any` permission is needed.
Creating an entry is done by calling:

`POST /address-book/entries`

.. code-block:: json

    {
        "label": "<public_entry_label>",
        "description": "<additional_information>",
        "operation": "<optional_operation_id>",
        "user": "<optional_user_id>"
    }

Response (201):

.. code-block:: json

    {
        "id": "<assigned_entry_id>"
        "label": "<public_entry_label>",
        "description": "<additional_information>",
        "operation": "<optional_operation_id>",
        "user": "<optional_user_id>"
        "user_details": {
            "id": "<associated_user_id>",
            "username": "<associated_username>",
            "first_name": "<associated_user_first_name>",
            "last_name": "<associated_user_last_name>"
            "is_active": true
        }
    }

The ``user_details``-field is ``null``, if no user is associated with the entry (``user``-field is ``null``).

Update entry
============

Updating an entry, not being associated with the requesting user, requires the :ref:`permission.logistics.address-book.entry.update.any` permission.
If needed and provided, updating is done via:

`PUT /address-book/entries/<entry_id>`

.. code-block:: json

    {
        "id": "<entry_id>"
        "label": "<public_entry_label>",
        "description": "<additional_information>",
        "operation": "<optional_operation_id>",
        "user": "<optional_user_id>"
    }

Delete entry
============

Deleting an entry, not being associated with the requesting user, requires the :ref:`permission.logistics.address-book.entry.delete.any` permission.
If needed and provided, updating is done via:

`DELETE /address-book/entries/<entry_id>`

Retrieve entries
================

Retrieval every entry is allowed, as long as it is visible by the requester.
This means, that if the entry is associated with a user, the requesting user must share an operation with him.
This does not apply to global entries (no associated user).
If the limit for associated users should be removed, the :ref:`permission.logistics.address-book.entry.view.any` permission is required.

`GET /address-book/entries/<entry_id>`

Response:

.. code-block:: json

    {
        "id": "<entry_id>"
        "label": "<public_entry_label>",
        "description": "<additional_information>",
        "operation": "<optional_operation_id>",
        "user": "<optional_user_id>"
        "user_details": {
            "id": "<associated_user_id>",
            "username": "<associated_username>",
            "first_name": "<associated_user_first_name>",
            "last_name": "<associated_user_last_name>",
            "is_active": true
        }
    }

Retrieving multiple entries in a :ref:`paginated <http-api.pagination>` list is a bit more complicated because of various use-cases.
A user might want to inspect his own entries or he might want to request all entries, visible to him.
Regarding entry visibility, the same regulations apply as for retrieving single ones.

`GET /address-book/entries`

Entry payload:

.. code-block:: json

    {
        "id": "<entry_id>"
        "label": "<public_entry_label>",
        "description": "<additional_information>",
        "operation": "<optional_operation_id>",
        "user": "<optional_user_id>"
        "user_details": {
            "id": "<associated_user_id>",
            "username": "<associated_username>",
            "first_name": "<associated_user_first_name>",
            "last_name": "<associated_user_last_name>",
            "is_active": true
        }
    }

The following fields can be used for ordering:

- ``label``
- ``description``

Additionally, query parameters can be applied in order to filter entries:

- ``by_user=<user_id>``: Only include entries, being associated with the user with the given id.
- ``for_operation=<operation_id>``: Only include entries for the operation with the given id or global ones.
- ``exclude_global=true``: Exclude entries with have no operation assigned.
- ``visible_by=<user_id>``: Only include entries, being visible to the user with the given id. If the :ref:`permission.logistics.address-book.entry.view.any` permission is not granted, this will have no effect, as the requesting users id is used here by default.
- ``include_for_inactive_users=false``: Includes entries, associated with inactive users.
- ``auto_delivery_enabled=false``: Whether to filter entries by having auto-delivery being enabled. This parameter is not allowed with search!

:ref:`Search <http-api.search>` allows using these filters as well and is available via:

`GET /address-book/entries/search`

Entry payload:

.. code-block:: json

    {
        "id": "<entry_id>"
        "label": "<public_entry_label>",
        "description": "<additional_information>",
        "operation": "<optional_operation_id>",
        "user": "<optional_user_id>"
        "user_details": {
            "id": "<associated_user_id>",
            "username": "<associated_username>",
            "first_name": "<associated_user_first_name>",
            "last_name": "<associated_user_last_name>",
            "is_active": true
        }
    }

The search index can be rebuilt via:

`POST /address-book/entries/search/rebuild`

Channels in General
===================

Channels are ways of delivering intel to recipients.
For example, an email channel is used for sending an email containing the intel to a target email address.
A radio channel might forward intel to a radio operator, that calls the recipient.
Each channel has a unique priority, timeout and minimum importance for intel.

The is-active-Flag of a channel describes whether the channel is available for intel-delivery.

Currently, the following channel types are supported, but not all implemented:

- (**Coming soon** |:rocket:|) **Direct** (`direct`): Use, if the recipient can be contacted directly, for example by talking.
- (**Coming soon** |:rocket:|) **Email** (`email`): Send an email and await a response.
- (**Coming soon** |:rocket:|) **Forward to Group** (`forward-to-group`): Forward intel to members of a group. This will use the first available address book entry for each member.
- (**Coming soon** |:rocket:|) **Forward to User** (`forward-to-user`): Forward intel to a user. This will use the first available address book entry for the user.
- **In-App Notification** (`in-app-notification`): Send an in-app notification via the MDS application and await it being read.
- (**Coming soon** |:rocket:|) **Phone Call** (`phone-call`): Call the recipient.
- (**Coming soon** |:rocket:|) **Radio** (`radio`): Forward to a radio operator, that transmits the intel over radio.

Each channel holds additional details, based on the type.

For **Direct** channel:

.. code-block:: json

    {
        "info": "<plain_text>"
    }

For **Email** channel:

.. code-block:: json

    {
        "email": "<target_email_address>"
    }

For **Forward to Group** channel:

.. code-block:: json

    {
        "forward_to_group": ["<target_group_id>"]
    }

If the referenced group is deleted, this channel will automatically be deleted as well.

For **Forward to User** channel:

.. code-block:: json

    {
        "forward_to_user": ["<target_user_id>"]
    }

If the referenced user is deleted, this channel will automatically be deleted as well.

For **Phone Call** channel:

.. code-block:: json

    {
        "phone": "<phone_number>"
    }

The phone number is expected to be in :e-164:`E.164 <>` format.

For **In-App Notification** channel:

.. code-block:: json

    {}

For **Radio** channel:

.. code-block:: json

    {
        "info": "<plain_text>"
    }

Set channels
============

Setting channels for global entries or ones, being associated to other users than the caller or different operations, the :ref:`permission.logistics.address-book.entry.update.any` permission is required.

`PUT /address-book/entries/<entry_id>/channels`

.. code-block:: json

    [
        {
            "entry": "<entry_id>",
            "is_active": false,
            "label": "<label>",
            "type": "<channel_type>",
            "priority": 20,
            "min_importance": 10,
            "details": {},
            "timeout": 8000
        }
    ]

This is a list of channels, that will be set.
Keep in mind that updating channels will restart all ongoing deliveries.
So if delivery was already tried over an old channel and failed or timed out, it will be tried again.

Retrieving channels
===================

Viewing channels for global entries or ones, being associated to other users than the caller or different operations, the :ref:`permission.logistics.address-book.entry.view.any` permission is required.

`GET /address-book/entries/<entry_id>/channel`

Response:

.. code-block:: json

    [
        {
            "id": "<channel_id>",
            "entry": "<entry_id>",
            "is_active": false,
            "label": "<label>",
            "type": "<channel_type>",
            "priority": 20,
            "min_importance": 10,
            "details": {},
            "timeout": 8000
        }
    ]
