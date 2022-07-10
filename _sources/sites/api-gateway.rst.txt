.. _chapter.api-gateway:

API Gateway
###########

MDS Server uses a concept of passing authentication information via HTTP headers.
The public Ingress forwards HTTP requests to the API Gateway.
This one manages sessions and handles login as well as logout.
After having logged in, the client is supplied with a public token, that it uses for all request.
For each request, the API Gateway uses the public token for checking if the client is logged in.
If he seems to be, user details are retrieved from the database regarding the logged in user and an internal signed JWT token is generated.
The HTTP request is forwarded with the generated "internal" token to the internal Ingress, that forwards to internal services.
Each service can then parse and validate the token, containing all relevant user information like username, permissions, etc.
Therefore, each service can check permissions, if needed, but session management only needs to be handled by the API Gateway.

Logging in
==========

Login with username and password via:

`POST /login`

.. code-block:: json

    {
        "username": "<username>",
        "pass": "<password>"
    }

Response `200`:

.. code-block:: json

    {
        "user_id": "<the_user_id>",
        "access_token": "<the-access-token>",
        "token_type": "Bearer"
    }

Making requests
===============

Each request should be made with the HTTP header:

.. code-block::

    Authorization: Bearer <the-access-token>

Logging out
===========

Logging out is done via:

`POST /logout`

This returns `200`, if logging out was successful.