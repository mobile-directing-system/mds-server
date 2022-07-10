Operations
##########

This section cover creating, retrieving and updating operations via API endpoints.

Deleting operations is not allowed.
They can only be marked as archived.

Each operation must have an non-empty title.
The end-timestamp must not be before the start-timestamp.

Create operation
================

In order to create an operation, the :ref:`permission.operation.create` permission is needed.
Creating an operation is done by calling:

`POST /operations`

.. code-block:: json

    {
        "title": "<operation_title>",
        "description": "<description>"
        "start": "2006-01-02T15:04:05Z07:00",
        "end": "2006-01-02T15:04:05Z07:00",
        "is_archived": false
    }

The ``end``-field is optional.

Response (201):

.. code-block:: json

    {
        "id": "<assigned_id>",
        "title": "<operation_title>",
        "description": "<description>"
        "start": "2006-01-02T15:04:05Z07:00",
        "end": "2006-01-02T15:04:05Z07:00",
        "is_archived": false
    }

Update operation
================

Updating an operation requires the :ref:`permission.operation.update` permission.
If provided, updating is done via:

`PUT /operations/<operation_id>`

.. code-block:: json

    {
        "id": "<operation_id",
        "title": "<operation_title>",
        "description": "<description>"
        "start": "2006-01-02T15:04:05Z07:00",
        "end": "2006-01-02T15:04:05Z07:00",
        "is_archived": false
    }

Retrieve operations
===================

Retrieving an operation by its id is possible via:

`GET /operations/<operation_id>`

Response (200):

.. code-block:: json

    {
        "id": "<operation_id",
        "title": "<operation_title>",
        "description": "<description>"
        "start": "2006-01-02T15:04:05Z07:00",
        "end": "2006-01-02T15:04:05Z07:00",
        "is_archived": false
    }

Retrieving a :ref:`paginated <http-api.pagination>` operation list requires the :ref:`permission.operation.view.any` permission:

`GET /operations`

Entry payload:

.. code-block:: json

    {
        "id": "<operation_id",
        "title": "<operation_title>",
        "description": "<description>"
        "start": "2006-01-02T15:04:05Z07:00",
        "end": "2006-01-02T15:04:05Z07:00",
        "is_archived": false
    }

The following fields can be used for ordering:

- ``title``
- ``description``
- ``start``
- ``end``
- ``is_archived``
