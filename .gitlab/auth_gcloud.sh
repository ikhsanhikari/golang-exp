#! /bin/bash

echo "Getting service key..."
echo $GCLOUD_SERVICE_KEY > ${HOME}/gcloud-service-key.json

echo "Authenticating service account..."
gcloud auth activate-service-account --key-file=${HOME}/gcloud-service-key.json

echo "Setting Google Project..."
gcloud --quiet config set project ${GCLOUD_PROJECT_ID}
