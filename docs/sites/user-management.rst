Users
#####

This section covers all user-related actions and API endpoints.

A user consists of an unique id, a name, unique username and password for logging in, as well as an is-admin-state.
The is-admin-state grants all permissions and normally should not be granted to regular users.

When the system starts and no admin-user exists, an example admin user will be created.
The default credentials are::

    Username: admin
    Password: admin

**Change the password immediately after logging in for the first time!**

A valid user must have non-empty username, first name and last name.

Create user
===========

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
===========

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
====================

Updating a user's password, not being the caller, requires the :ref:`permission.user.update-pass` permission and is done via:

`PUT /users/<user_id/pass`

.. code-block:: json

    {
        "user_id": "<user_id>",
        "new_pass": "<new_password_in_plaintext>"
    }

Delete user
===========

Deleting a user requires the :ref:`permission.user.delete` permission and is done via:

`DELETE /users/<user_id>`

Retrieve users
==============

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

Available query parameters:

- ``only_ongoing``: Limit only to operations, currently running (default is ``false``).
- ``include_archived``: Include archived operations in results (default is ``false``).
- ``for_user``: Only show operations, the given user is member of.

The following fields can be used for ordering:

- ``username``
- ``first_name``
- ``last_name``
- ``is_admin``

:ref:`Search <http-api.search>` is available via:

`GET /users/search`

Entry payload:

.. code-block:: json

    {
        "id": "<assigned_id>",
        "username": "<username>",
        "first_name": "Max",
        "last_name": "Mustermann",
        "is_admin": false
    }

The search index can be rebuilt via:

`POST /users/search/rebuild`
