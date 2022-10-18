Radio Delivery
##############

Radio deliveries are for radio channels.
Delivery requests are added to queues and assigned to radio operators, if they request one.
Prioritization is either highest importance or oldest timestamp first.

Clients receive updates via :ref:`websocket` (gate `desktop-app` with channel `radio-delivery`) which informs them about possible new delivery requests.
Payload has the following content:

.. code-block:: json

    {
        "type": "new-radio-deliveries-available",
        "payload": {
            "operation": "<operation_with_possible_new_requests>"
        }
    }

Retrieve next radio delivery to deliver
=======================================

In order to pick up requests, the :ref:`permission.radio-delivery.deliver.any` permission is required.
Picking up a request is done by calling:

`GET /radio-deliveries/operations/<operation_id>/next`

Response if none available (204): None

Response if picked up (200):

.. code-block:: json

    {
        "id": "<attempt_id>",
        "intel": "<intel_id>",
        "intel_operation": "<operation_id_of_intel>",
        "intel_importance": 0,
        "assigned_to": "<assigned_to_address_book_entry>",
        "assigned_to_label": "<name_of_assigned_address_book_entry>",
        "delivery": "<delivery_id>",
        "channel": "<channel_id>",
        "created_at": "<attempt_creation_ts>",
        "status_ts": "<radio_delivery_status_update_ts>",
        "note": "<optional_radio_delivery_note>",
        "accepted_at": "<attempt_acception_ts>"
    }

Release picked up radio delivery
================================

In order to release a picked up radio delivery, the :ref:`permission.radio-delivery.deliver.any` permission is required.
If a delivery is to be released, not being assigned to the requesting client, the :ref:`permission.radio-delivery.manage.any` permission is needed.
Releasing is done by calling:

`POST /radio-deliveries/<attempt_id>/release`

Finish picked up radio delivery
===============================

In order to finish a picked up radio delivery, the :ref:`permission.radio-delivery.deliver.any` permission is required.
If a delivery is to be finished, not being assigned to the requesting client, the :ref:`permission.radio-delivery.manage.any` permission is needed.
Finishing is done by calling:

`POST /radio-deliveries/<attempt_id>/finish`

.. code-block:: json

    {
        "success": true,
        "note": "<note>"
    }

The ``success``-Field can be ``true`` or ``false`` according to if delivery was successful.
Keep in mind, that ``false`` will result in radio delivery being considered unsuccessful.
If you simply want to not deliver anymore, release it instead.
