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

Address book
------------

.. _permission.logistics.address-book.entry.create.any:

logistics.address-book.entry.create.any
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Allows creating address book entries, that can also be global or associated to foreign users.
This also includes associating entries with operations, the target user or the requesting client is not part of.

Options: `none`

.. _permission.logistics.address-book.entry.update.any:

logistics.address-book.entry.update.any
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Allows updating address book entries, that can also be global or associated to foreign users.
This also includes associating entries with operations, the target user or the requesting client is not part of.

Options: `none`

.. _permission.logistics.address-book.entry.delete.any:

logistics.address-book.entry.delete.any
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Allows deletion of address book entries, that can also be global or associated to foreign users.
This also includes associating entries with operations, the requesting client is not part of.

Options: `none`

.. _permission.logistics.address-book.entry.view.any:

logistics.address-book.entry.view.any
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Allows retrieval of all address book entries, including that can also be global or associated to foreign users.
This also includes associating entries with operations, the requesting client is not part of.

Options: `none`

Groups
------

.. _permission.group.create:

group.create
^^^^^^^^^^^^

Allows creating groups.

Options: `none`

.. _permission.group.update:

group.update
^^^^^^^^^^^^

Allows updating groups.

Options: `none`

.. _permission.group.delete:

group.delete
^^^^^^^^^^^^

Allows deleting groups.

Options: `none`

.. _permission.group.view:

group.view
^^^^^^^^^^

Allows retrieving of groups with details.

Options: `none`

Intelligence
------------

.. _permission.intelligence.intel.create:

intelligence.intel.create
^^^^^^^^^^^^^^^^^^^^^^^^^

Allows creating intel (if member of operation).

Options: `none`

.. _permission.intelligence.intel.invalidate:

intelligence.intel.invalidate
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Allows invalidating intel (if member of operation).

Options: `none`

.. _permission.intelligence.intel.view.any:

intelligence.intel.view.any
^^^^^^^^^^^^^^^^^^^^^^^^^^^

Allows viewing any intel (if member of operation), even if not assigned to.

Options: `none`

Operations
----------

.. _permission.operation.view.any:

operation.view.any
^^^^^^^^^^^^^^^^^^

Allows listing and viewing all registered operations. In contrast to that, usually, only viewing operations via id is allowed.

.. _permission.operation.create:

operation.create
^^^^^^^^^^^^^^^^

Allows creating operations.

.. _permission.operation.update:

operation.update
^^^^^^^^^^^^^^^^

Allows updating of operations. This also includes marking them as finished or archived.

.. _permission.operation.members.view:

operation.members.view
^^^^^^^^^^^^^^^^^^^^^^

Allows retrieving members for operations.

.. _permission.operation.members.update:

operation.members.update
^^^^^^^^^^^^^^^^^^^^^^^^

Allows (un)assigning members to operations.

Users
-----

.. _permission.user.create:

user.create
^^^^^^^^^^^

Allows creating users.

Options: `none`

.. _permission.user.set-active-state:

user.set-active-state
^^^^^^^^^^^^^^^^^^^^^

Allows setting the active-state for users.

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

Misc
----

.. _permission.core.search.rebuild-index:

core.search.rebuild-index
^^^^^^^^^^^^^^^^^^^^^^^^^

Allows rebuilding of search indices. Because of taking a long time, this permission is required.

Options: `none`