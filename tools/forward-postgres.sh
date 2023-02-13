#!/bin/bash

trap terminate SIGINT
terminate() {
    pkill -SIGINT -P $$
    exit
}

START_PORT=$((4000))
lastPort=$START_PORT
messages=()

forwardPostgreSQLInBackground() {
  newPort=$((lastPort+1))
  lastPort=$newPort
  messages+=("PORT $newPort: $1")
  kubectl port-forward "$1" $((newPort)):5432 &
}

# Do not change order. If new services are to be added, always append them!
forwardPostgreSQLInBackground "service/mds-api-gateway-svc-postgres-service"
forwardPostgreSQLInBackground "service/mds-user-svc-postgres-service"
forwardPostgreSQLInBackground "service/mds-permission-svc-postgres-service"
forwardPostgreSQLInBackground "service/mds-group-svc-postgres-service"
forwardPostgreSQLInBackground "service/mds-operation-svc-postgres-service"
forwardPostgreSQLInBackground "service/mds-logistics-svc-postgres-service"
forwardPostgreSQLInBackground "service/mds-in-app-notifier-svc-postgres-service"
forwardPostgreSQLInBackground "service/mds-radio-delivery-svc-postgres-service"

sleep 1
echo "*********************************"
for message in "${messages[@]}"; do
  echo "$message"
done
echo "*********************************"
wait
