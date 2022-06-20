Permissions
===========

Permissions always have a unique name like `user.create` and optional additional options.

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