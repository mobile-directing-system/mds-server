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

Retrieving a :ref:`paginated <http-api.pagination>` operation list requires the :ref:`permission.operation.view.any` permission, if all operations should be retrieved.
Otherwise, only operations are returned that the requesting client is member of.

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

:ref:`Search <http-api.search>` is available via:

`GET /operations/search`

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

The search index can be rebuilt via:

`POST /operations/search/rebuild`

Update operation members
========================

Updating operation members requires the :ref:`permission.operation.members.update` permission:

`PUT /operations/<operation_id>/members`

.. code-block:: json

    [
        "<user_id_1>",
        "<user_id_2>",
        "<user_id_n>"
    ]

Retrieve operation members
==========================

Retrieving a :ref:`paginated <http-api.pagination>` member list requires the :ref:`permission.operation.members.view` permission and is done via:

`GET /operations/<operation_id>/members`

Entry payload:

.. code-block:: json

    {
        "id": "<user_id>",
        "username": "<username>",
        "first_name": "<first_name>",
        "last_name": "<last_name>"
    }

The following fields can be used for ordering:

- ``username``
- ``first_name``
- ``last_name``