Manual Intel Delivery
#####################

MDS Server supports automatic intel delivery for selected address book entries.
Manual intel delivery is possible through usage of HTTP endpoints while also being notified about open intel deliveries via WebSocket.

Listen for open intel deliveries
================================

Connect via :ref:`websocket` and use the gate `desktop-app` with message channel `open-intel-delivery-notifier`.
Subscribing to notifications for open intel deliveries requires the :ref:`permission.logistics.intel-delivery.manage` permission.

In order to subscribe to notifications for an operation, send the following message:

.. code-block:: json

    {
        "type": "subscribe-open-intel-deliveries",
        "payload": {
            "operation": "<operation_id>"
        }
    }

You will receive a confirmation as you are sent the following message:

.. code-block:: json

    {
        "type": "subscribed-open-intel-deliveries",
        "payload": {
            "operations": ["<subscribed_operation_1>", "<subscribed_operation_2>"]
        }
    }

Open intel deliveries for subscribed operations are then received on updates in the following format:

.. code-block:: json

    {
        "type": "open-intel-deliveries",
        "payload": {
            "operation": "<operation_id>",
            "open_intel_deliveries": [
                {
                    "delivery": {
                        "id": "<delivery_id>",
                        "intel": "<intel_id>",
                        "to": "<recipient_address_book_entry_id>",
                        "note": "<optional_note>"
                    },
                    "intel": {
                        "id": "<intel_id>",
                        "created_at": "<timestamp_of_creation>",
                        "created_by": "<user_id_of_creator>",
                        "operation": "<referenced_operation_id>",
                        "importance": 0,
                        "is_valid": true
                    }
                }
            ]
        }
    }

Of course, subscribing to multiple operations at the same time is possible.

If you want to unsubscribe from an operation, send the following message:

.. code-block:: json

    {
        "type": "unsubscribe-open-intel-deliveries",
        "payload": {
            "operation": "<operation_id>"
        }
    }

Retrieve delivery attempts
==========================

All delivery attempts for a delivery can be retrieved using the :ref:`permission.logistics.intel-delivery.manage` permission via:

`GET /intel-deliveries/<delivery_id>/attempts`

Response (200):

.. code-block:: json

    [
        {
            "id": "<attempt_id>",
            "delivery": "<delivery_id>",
            "channel": "<channel_id>",
            "created_at": "<attempt_creation_timestamp>",
            "is_active": false,
            "status": "<attempt_status>",
            "status_ts": "<status_timestamp>",
            "note": "<optional_note>"
        }
    ]

With ``status`` being one of the following:

- ``open``
- ``awaiting-delivery``
- ``delivering``
- ``awaiting-ack``
- ``delivered``
- ``timeout``
- ``canceled``
- ``failed``

Manually schedule a delivery attempt
====================================

Scheduling a delivery attempt requires the :ref:`permission.logistics.intel-delivery.manage` permission and can be done via:

`POST /intel-deliveries/<delivery_id>/deliver/channel/<channel_id>`

Response (201)

Enable auto intel delivery for address book entries
===================================================

Managing intel auto delivery requires the :ref:`permission.logistics.intel-delivery.manage` permission.
Updating the list of address book entries with auto delivery enabled can be done via:

`PUT /address-book/entries-with-auto-intel-delivery`

.. code-block:: json

    [
        "<entry_id_1>",
        "<entry_id_2>"
    ]

Response (200)

Cancel active intel delivery
============================

Cancelling an active intel delivery requires the :ref:`permission.logistics.intel-delivery.manage` permission and can be done via:

`POST /intel-deliveries/<delivery_id>/cancel`

.. code-block:: json

    {
        "success": false,
        "note": "<option_note>"
    }

Set ``success`` to ``true`` if the delivery should be marked as successful.
This means that the intel to be delivered via this delivery as confirmed to be delivered and read by the recipient.
Otherwise, set this to ``false``.
It is good practise to provide a note as well, describing why the delivery was cancelled.

Note: All active delivery attempts will be cancelled as well.
