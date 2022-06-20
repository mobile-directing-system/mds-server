General HTTP API
################

This chapter covers some general aspects regarding the HTTP API.

Authentication and permissions
==============================

For most endpoints, authentication is required.
This is explained in the :ref:`chapter.api-gateway` chapter.

If an endpoint requires authentication, a `401` code is returned.
If a caller lacks required permissions, a `403` code is returned.

.. _http-api.pagination:

Pagination
==========

Most endpoints with list data use pagination in order to limit results and data exchange.
The following query parameters are available:

.. list-table::
    :header-rows: 1

    *   - Parameter
        - Description
    *   - ``limit``
        - Maximum amount of entries to retrieve. Defaults to 20.
    *   - ``offset``
        - Offset for retrieved entries.
    *   - ``order_by``
        - Field to order entries by.
    *   - ``order_dir``
        - Direction for ordering. Supported values are ``asc`` and ``desc`` with the first being default.

Paginated responses will look like the following:

.. code-block:: json

    {
        "total": 0,
        "limit": 0,
        "offset": 0,
        "ordered_by": "<field_name>",
        "order_dir": "<order_direction>",
        "retrieved": 0,
        "entries": []
    }

.. list-table::
    :header-rows: 1

    *   - Field
        - Description
    *   - ``total``
        - Total amounts of available entries.
    *   - ``limit``
        - Applied limit for retrieving entries.
    *   - ``offset``
        - Applied offset for retrieved entries.
    *   - ``ordered_by``
        - Field name the entries were ordered by.
    *   - ``order_dir``
        - Applied direction for ordering entries. As with query parameters, ``asc`` and ``desc`` are possible values.
    *   - ``retrieved``
        - Amount of entries in the ``entries``-field.
    *   - ``entries``
        - The actual entries. Structure depends on the retrieved data.