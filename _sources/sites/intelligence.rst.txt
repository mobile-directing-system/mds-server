Intelligence
############

The core of Mobile Directing System is intelligence or intel.
Intel can be any type of information flowing in the system.

Intel is always associated with an operation and creators and viewers must be member of that operation.

Intel types
===========

analog-radio-message
^^^^^^^^^^^^^^^^^^^^

Used for radio messages, received via analog radio.

Content:

.. code-block:: json

    {
        "channel": "<channel_over_which_received>",
        "callsign" "<sender_callsign>",
        "head": "<message_head>",
        "content": "<message_content>"
    }

The fields ``callsign`` and ``content`` must not be empty.

plaintext-message
^^^^^^^^^^^^^^^^^

Used for plaintext content.

Content:

.. code-block:: json

    {
        "text": "<content>"
    }

The ``text``-field must not be empty.

Create intel
============

In order to create intel, the :ref:`permission.intelligence.intel.create` permission is needed and the creator must be member of the operation.
Creating intel is done by calling:

`POST /intel`

.. code-block:: json

    {
        "operation": "<associated_operation_id>",
        "type": "<intel_type>",
        "content": {},
        "importance": 0,
        "initial_deliver_to": [
            "<address_book_entry_1>",
            "<address_book_entry_2>",
            "<address_book_entry_n>",
        ]
    }

Response (201):

.. code-block:: json

    {
        "id": "<assigned_id>",
        "created_at": "<creation_timestamp>",
        "created_by": "<creator_user_id>",
        "operation": "<associated_operation_id>",
        "type": "<intel_type>",
        "content": {},
        "search_text": "<search_text>",
        "importance": 0,
        "is_valid": true
    }

Invalidate intel
================

Intel cannot be deleted.
However, if marking is needed, invalidation is possible.

Invalidating intel requires the :ref:`permission.intelligence.intel.invalidate` permission as well as being member of the associated operation.
If provided, invalidation is done via:

`POST /intel/<intel_id>/invalidate`

Retrieve intel
==============

Single intel can be retrieved by any user which owns address book entries being recipients in deliveries for this intel.
Otherwise, the :ref:`permission.intelligence.intel.view.any` permission is required.

Retrieval is done via:

`GET /intel/<intel_id>`

Response (200):

.. code-block:: json

    {
        "id": "<assigned_id>",
        "created_at": "<creation_timestamp>",
        "created_by": "<creator_user_id>",
        "operation": "<associated_operation_id>",
        "type": "<intel_type>",
        "content": {},
        "search_text": "<search_text>",
        "importance": 0,
        "is_valid": true
    }

Retrieving intel as a :ref:`paginated <http-api.pagination>` list is possible via:

`GET /intel`

Entry payload:

.. code-block:: json

    {
        "id": "<intel_id>",
        "created_at": "<creation_timestamp>",
        "created_by": "<creator_user_id>",
        "operation": "<associated_operation_id>",
        "type": "<intel_type>",
        "content": {},
        "search_text": "<search_text>",
        "importance": 0,
        "is_valid": true
    }

Entries are ordered descending by creation timestamp.
Additionally, query parameters can be applied in order to filter intel:

- ``created_by=<user_id>``: Only include intel, created by the user with the given id.
- ``operation=<operation_id>``: Only include intel for the given operation.
- ``intel_type=<intel_type>``: Only include intel of the given type.
- ``min_importance=0``: Only include intel with the given minimum importance.
- ``include_invalid=false``: Include invalid intel as well.
- ``one_of_delivery_for_entries=["<entry_id_1>","<entry_id_n>"]``: Only include intel with deliveries for the address book entries with the given ids.
- ``one_of_delivered_to_entries=["<entry_id_1>","<entry_id_n>"]``: Only include intel with successful deliveries for the address book entries with the given ids.

:ref:`Search <http-api.search>` provides these filters as well and is available via:

`GET /intel/search`

Entry payload:

.. code-block:: json

    {
        "id": "<intel_id>",
        "created_at": "<creation_timestamp>",
        "created_by": "<creator_user_id>",
        "operation": "<associated_operation_id>",
        "type": "<intel_type>",
        "content": {},
        "search_text": "<search_text>",
        "importance": 0,
        "is_valid": true
    }

The search index can be rebuilt via:

`POST /intel/search/rebuild`

Intel-delivery
==============

Intel-delivery is managed by the MDS Server.
However, clients need to notify the server, when intel was successfully delivered (and read!).
When the source was an actual attempt, it should be confirmed via:

`POST /intel-delivery-attempts/<attempt_id>/delivered`

This makes it easier to see which channels were successful.

Otherwise, if no attempt is known, you can confirm via:

`POST /intel-deliveries/<delivery_id>/delivered`

Keep in mind, that you can only confirm deliveries that were assigned to you.
