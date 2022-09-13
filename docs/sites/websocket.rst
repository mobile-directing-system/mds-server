.. _websocket:

WebSocket
#########

This section covers general WebSocket communication with the MDS Server.
Due to its microservice-architecture and communicating with many services, adjustments in WebSocket communication were needed.
Therefore, an approach was chosen that allows communicating with multiple ones over the same WebSocket connection.
When initiating a connection, it is made for a certain so called **gate**.
Internally, a gate is configured to use a certain number of services.
The WebSocket-Hub holds the connection to the client and negotiates between these services.
This is done through **channels**.
Each message is of the following format:

.. code-block:: json

    {
        "channel": "<channel_name>",
        "payload": {}
    }

Based on the channel, the hub will forward to the correct service and for messages from a service, the channel will be set accordingly.
