Development
###########

This chapter provides instructions in order to explain the development setup with MDS Server.
As MDS Server is meant to be deployed on a Kubernetes cluster, local development requires some adjustments.
Here, we describe usage with :docker-homepage:`Docker <>`, :minikube-homepage:`minikube <>`, :skaffold-homepage:`Skaffold <>` and :intellij-cloud-code-plugin-homepage:`IntelliJ Cloud Code Plugin <>`.

Setup Project
=============

MDS Server source code is stored in a :github-repo:`GitHub Repository <>`.
If you want to contribute, create a fork and checkout the repository locally using :git-homepage:`Git <>`.
If you only want to run the server, you can clone the regular repository with:

.. code-block:: sh

    git clone https://github.com/mobile-directing-system/mds-server.git

From now on, we expect you to use the repository root as your working directory.

Setup Docker
============

Docker provides needed containerization for services.
For installation instructions refer to the Docker documentation :docker-install:`here <>`.

Setup minikube
==============

First, you need to install minikube.
Installation instructions for your operating system are available :minikube-install:`here <>`.
Then we need to start a cluster:

.. code-block:: sh

    minikube start

You can also launch the Kubernetes dashboard by running:

.. code-block:: sh

    minikube dashboard

Prepare deployment
==================

Kubernetes uses YAML-files for deploying configurations.
Skaffold simplifies working with these files as well as building Docker images, etc.
When developing in JetBrains IDEs like :goland-homepage:`GoLand <>`, you can use the :intellij-cloud-code-plugin-homepage:`IntelliJ Cloud Code Plugin <>`.
Detailed usage instructions are available :intellij-cloud-code-plugin-install:`here <>`.
However, we will explain the approach with using Skaffold manually.

First make sure that Skaffold is installed. Installation instructions can be found :skaffold-install:`here <>`.

The Skaffold configuration provides two relevant profiles: `prepare` and `mds`.
Due to Skaffold's nature, we can only apply all referenced YAML-files at once.
However, CRDs for example need to be applied and then processed by Kubernetes until it can understand the files, that rely on them.
This also applies to NGINX configurations.
Therefore, the `prepare`-profile provides basic setup for resources as well as NGINX deployments and the `mds`-profile includes the rest of the actual server application.
Normally, the `prepare`-profile needs to be deployed exactly once after a new Kubernetes was created with minikube.
For deploying the profile, execute the following command:

.. code-block:: sh

    skaffold deploy --profile=prepare

Now everything is setup for the actual deployment of MDS Server.

Deploy for development using Skaffold
=====================================

For actually deploying MDS Server, we need to build required Docker images as well as deploy everything to Kubernetes.
This job is handled by Skaffold.
Run from the working directory:

.. code-block:: sh

    skaffold dev --profile=mds

This will build all Docker images and deploy to the Kubernetes cluster.
Skaffold will keep running.
For stopping MDS Server, simply press ``CTRL+C``.
This will make Skaffold remove all applied configurations and make Kubernetes shut down all components.

We expect you to be running minikube as Docker container.
In order to access services provided by MDS Server, you need to find the minikube IP address as well as the port number.
This can be achieved by running the following command:

.. code-block:: sh

    minikube service list

This will return something like the following:

.. code-block:: sh

    |------------------------|---------------------------------------------|--------------|---------------------------|
    |       NAMESPACE        |                    NAME                     | TARGET PORT  |            URL            |
    |------------------------|---------------------------------------------|--------------|---------------------------|
    | default                | kubernetes                                  | No node port |
    ...
    | kafka                  | kafka-cluster-zookeeper-nodes               | No node port |
    | kafka                  | kafka-ui-service                            | http/30010   | http://192.168.49.2:31847 |
    | kube-system            | kube-dns                                    | No node port |
    | public-ingress-nginx   | public-ingress-nginx-controller             | http/30080   | http://192.168.49.2:30080 |
    |                        |                                             | https/30443  | http://192.168.49.2:30443 |
    | public-ingress-nginx   | public-ingress-nginx-controller-admission   | No node port |
    |------------------------|---------------------------------------------|--------------|---------------------------|

From the information provided, we can access the servers HTTP service at `http://192.168.49.2:30080`.
The IP address is also shown when running:

.. code-block:: sh

    minikube ip

If you want to, you can create a new entry in `/etc/hosts` file with your minikube IP address:

.. code-block::

    192.168.49.2 minikube

This allows you accessing the HTTP service via `http://minikube:30080`.

If something goes wrong
=======================

Due to some bugs, sometimes resources might not get cleaned up or the cluster is in an unexpected state.
In this case, a hard reset can be performed by running:

.. code-block:: sh

    minikube delete

Keep in mind, that in order to deploy again, you need to deploy the `prepare`-profile again as well as described in this chapter.

Viewing logs
============

Logs can be either viewed through the Kubernetes dashboard by navigating to the desired pod and it's logs or through Kibana.
Part of the deployment is :fluent-bit-homepage:`Fluent Bit<>`, which allows centralized logging.
It forwards processed log entries to Elasticsearch.
Results can be viewed in Kibana, accessible via `http://minikube:30090`.
Upon first start do the following:

- Open the sidebar to the left.
- Under `Analytics`, click `Discover`.
- Click `Create data view`.
- Click `Create data view`.

You can now view logs using the `Discover`-button on the left toolbar (the compass-icon).
If you only want to see logs from MDS pods, type ``kubernetes.pod_name : mds*`` in the search bar.