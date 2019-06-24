#! /bin/bash

REPOSITORY_ADDRESS="${REPOSITORY_HOST}/${GCLOUD_PROJECT_ID}/${CI_PROJECT_NAMESPACE}/${CI_PROJECT_NAME}"
docker build \
  --tag "${REPOSITORY_ADDRESS}:${CI_PIPELINE_IID}" \
  --build-arg sshkey="${SSH_KEY}" \
  --build-arg CI_PROJECT_NAME \
  .
gcloud auth configure-docker --quiet
docker push ${REPOSITORY_ADDRESS}:${CI_PIPELINE_IID}