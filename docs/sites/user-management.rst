Users
=====

This section covers all user-related actions and API endpoints.

Create user
-----------

In order to create a user, the :ref:`permission.user.create` permission is needed.
Creating a user is done by calling:

`POST /users`

.. code-block:: json

    {
        "username": "<username_for_logging_in>",
        "first_name": "Max",
        "last_name": "Mustermann",
        "is_admin": false,
        "pass": "<password_in_plaintext>"
    }

Response:

.. code-block:: json

    {
        "id": "<assigned_id>",
        "username": "<username>",
        "first_name": "Max",
        "last_name": "Mustermann",
        "is_admin": false
    }

Keep in mind, that for creating a user with ``is_admin`` set to ``true``, the :ref:`permission.user.set-admin` permission is needed as well.

Update user
-----------

Updating a user requires the :ref:`permission.user.update` permission.
If provided, updating is done via:

`PUT /users/<user_id>`

.. code-block:: json

    {
        "id": "<user_id>"
        "username": "<username_for_logging_in>",
        "first_name": "Max",
        "last_name": "Mustermann",
        "is_admin": false,
    }

As with creating a user, changing the ``is_admin``-field requires the :ref:`permission.user.set-admin` permission, too.

Update user password
--------------------

Updating a user's password, not being the caller, requires the :ref:`permission.user.update-pass` permission and is done via:

`PUT /users/<user_id/pass`

.. code-block:: json

    {
        "user_id": "<user_id>",
        "new_pass": "<new_password_in_plaintext>"
    }

Delete user
-----------

Deleting a user requires the :ref:`permission.user.delete` permission and is done via:

`DELETE /users/<user_id>`

Retrieve users
--------------

Retrieving users, not being the caller, requires the :ref:`permission.user.view` permission.
A single user can be retrieved using:

`GET /users/<user_id>`

Response:

.. code-block:: json

    {
        "id": "<assigned_id>",
        "username": "<username>",
        "first_name": "Max",
        "last_name": "Mustermann",
        "is_admin": false
    }

:ref:`Paginated <http-api.pagination>` user lists can be retrieved via:

`GET /users`

Entry payload:

.. code-block:: json

    {
        "id": "<assigned_id>",
        "username": "<username>",
        "first_name": "Max",
        "last_name": "Mustermann",
        "is_admin": false
    }