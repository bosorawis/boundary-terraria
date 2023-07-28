#!/bin/sh

TASKID=$(curl -s "$ECS_CONTAINER_METADATA_URI_V4/task" | jq -r ".TaskARN" | cut -d "/" -f 3)

sed -i "s/%%CLUSTERID%%/$HCP_BOUNDARY_CLUSTER_ID/g" worker.hcl
sed -i "s/%%TASKID%%/$TASKID/g" worker.hcl

exec "$@"
