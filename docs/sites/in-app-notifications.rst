In-App Notifications
####################

Similar to push-notifications, MDS Server allows in-app notifications as a channel-type.
The notifier is available via :ref:`websocket` and uses the gate `desktop-app` with message channel `in-app-notifier`.

The payload for a notification has the following content:

.. code-block:: json

    {
        "type": "intel-notification",
        "payload": {
            "intel_to_deliver": {
                "attempt": "<attempt_id>",
                "id": "<intel_id>",
                "created_at": "<intel_created_at_timestamp>",
                "created_by": "<intel_creator_user_id>",
                "operation": "<intel_operation_id>",
                "type": "<intel_type>",
                "content": {},
                "importance": 0
            },
            "delivery_attempt": {
                "id": "<attempt_id>",
                "assigned_to": "<assigned_address_book_entry_id>",
                "assigned_to_label": "<assigned_address_book_entry_label>",
                "assigned_to_user": "<optionally_assigned_user_from_address_book_entry>",
                "delivery": "<delivery_id>",
                "channel": "<used_channel_id>",
                "created_at": "<attempt_created_at_timestamp>",
                "is_active": true,
                "status_ts": "<status_updated_timestamp>",
                "note": "<optional_note>",
                "accepted_at": "<attempt_accepted_by_notifier_timestamp>"
            },
            "channel": {
                "id": "<used_channel_id>",
                "entry": "<assigned_address_book_entry_id>",
                "label": "<used_channel_label>",
                "timeout": 0
            },
            "creator_details": {
                "id": "<creator_user_id>",
                "username": "<creator_username>",
                "first_name": "<creator_first_name>",
                "last_name": "<creator_last_name>",
                "is_active": true
            },
            "recipient_details": {
                "id": "<recipient_user_id>",
                "username": "<recipient_username>",
                "first_name": "<recipient_first_name>",
                "last_name": "<recipient_last_name>",
                "is_active": true
            }
        }
    }

The ``recipient_details``-field is optional as the assigned address book entry may not have an assigned user.
