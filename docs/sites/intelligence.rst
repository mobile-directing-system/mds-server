Intelligence
############

The core of Mobile Directing System is intelligence or intel.
Intel can be any type of information flowing in the system.

Intel is always associated with an operation and creators and viewers must be member of that operation.

Intel types
===========

plaintext-message
^^^^^^^^^^^^^^^^^

Used for plaintext content.

Options:

.. code-block:: json

    {
        "text": "<content>"
    }

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
        "assignments": [
            {
                "to": "<assigned_address_book_entry_id>"
            }
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
        "is_valid": true,
        "assignments": [
            {
                "id": "<assignment_id>",
                "to": "<assigned_address_book_entry_id>"
            }
        ]
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

Retrieving intel (if member of the operation) does not require any permission.
However, retrieving intel, not being associated with the requester, requires the :ref:`permission.intelligence.intel.view.any` permission.

Usually, intel-retrieval should be done via the mailbox and therefore no batch-retrieval or search is provided.
Only intels by id can be retrieved via:

`GET /intel/<intel_id>`

Response:

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
        "is_valid": true,
        "assignments": [
            {
                "id": "<assignment_id>",
                "to": "<assigned_address_book_entry_id>"
            }
        ]
    }

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
