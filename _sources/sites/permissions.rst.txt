Permissions
###########

Permissions always have a unique name like `user.create` and optional additional options.

Setting permissions
===================

Setting permissions requires the :ref:`permission.permissions.update` permission and is accomplished by calling:

`PUT /permissions/user/<user_id>`

.. code-block:: json

    [
        {
            "name": "<name_of_first_permission>",
            "options": null
        },
        {
            "name": "<name_of_second_permission>",
            "options": {
                "hello": "world"
            }
        }
    ]

The ``options``-field contains the options available for the certain permission.

Retrieving permissions
======================

Retrieving permissions for users requires the :ref:`permission.permissions.view` permission, if not retrieving for the caller.
Retrieval is done via:

`GET /permissions/user/<user_id>`

Response:

.. code-block:: json

    [
        {
            "name": "<name_of_first_permission>",
            "options": null
        },
        {
            "name": "<name_of_second_permission>",
            "options": {
                "hello": "world"
            }
        }
    ]

Permission list
===============

This is a list of all available permissions.

Permissions
-----------

Permissions regarding permissions themselves like updating or retrieving.

.. _permission.permissions.update:

permissions.update
^^^^^^^^^^^^^^^^^^

Allows setting permissions for users.

Options: `none`

.. _permission.permissions.view:

permissions.view
^^^^^^^^^^^^^^^^^^

Allows retrieving permissions of users.

Options: `none`

Users
-----

.. _permission.user.create:

user.create
^^^^^^^^^^^

Allows creating users.

Options: `none`

.. _permission.user.delete:

user.delete
^^^^^^^^^^^

Allows deleting users.

Options: `none`

.. _permission.user.set-admin:

user.set-admin
^^^^^^^^^^^^^^

Allows setting the is-admin-state of users.

Options: `none`

.. _permission.user.update:

user.update
^^^^^^^^^^^

Allows updating a user. If the is-admin-state is wanted to be changed, the :ref:`permission.user.set-admin` permission is required, too.

Options: `none`

.. _permission.user.update-pass:

user.update-pass
^^^^^^^^^^^^^^^^

Allows setting the password of other users.

Options: `none`

.. _permission.user.view:

user.view
^^^^^^^^^

Allows retrieving information of other users.

Options: `none`