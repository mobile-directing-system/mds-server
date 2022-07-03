Groups
######

Groups are used for grouping users globally but also for operations.
These groups can be used in the adress book.

Create group
============

In order to create a group, the :ref:`permission.group.create` permission is needed.
Creating a group is done by calling:

`POST /groups`

.. code-block:: json

    {
        "title": "<group_title>",
        "description": "<group_description>",
        "operation": "<optional_operation_id>",
        "members": [
            "<id_of_member_1>",
            "<id_of_member_2>",
            "<id_of_member_n>"
        ]
    }

Response:

.. code-block:: json

    {
        "id": "<assigned_id>",
        "title": "<group_title>",
        "description": "<group_description>",
        "operation": "<optional_operation_id>",
        "members": [
            "<id_of_member_1>",
            "<id_of_member_2>",
            "<id_of_member_n>"
        ]
    }

Update group
============

Updating a group requires the :ref:`permission.group.update` permission.
If provided, updating is done via:

`PUT /groups/<group_id>`

.. code-block:: json

    {
        "id": "<group_id>",
        "title": "<group_title>",
        "description": "<group_description>",
        "operation": "<optional_operation_id>",
        "members": [
            "<id_of_member_1>",
            "<id_of_member_2>",
            "<id_of_member_n>"
        ]
    }

Delete group
============

Deleting a group requires the :ref:`permission.group.delete` permission and is done via:

`DELETE /groups/<group_id>`

Retrieve groups
===============

Retrieving groups requires the :ref:`permission.group.view` permission.
A single group can be retrieved using:

`GET /groups/<group_id>`

Response:

.. code-block:: json

    {
        "id": "<group_id>",
        "title": "<group_title>",
        "description": "<group_description>",
        "operation": "<optional_operation_id>",
        "members": [
            "<id_of_member_1>",
            "<id_of_member_2>",
            "<id_of_member_n>"
        ]
    }

:ref:`Paginated <http-api.pagination>` group lists can be retrieved via:

`GET /groups`

Entry payload:

.. code-block:: json

    {
        "id": "<group_id>",
        "title": "<group_title>",
        "description": "<group_description>",
        "operation": "<optional_operation_id>",
        "members": [
            "<id_of_member_1>",
            "<id_of_member_2>",
            "<id_of_member_n>"
        ]
    }

The following fields can be used for ordering:

- ``title``
- ``description``

Additionally, query parameters can be applied in order to filter groups:

- ``by_user=<user_id>``: Only include groups, that the user with the given id is member of.
- ``for_operation=<operation_id>``: Only include groups for the operation with the given id or ones, that do not have an operation assigned (global groups).
- ``exclude_global=true``: Exclude groups with have no operation assigned.