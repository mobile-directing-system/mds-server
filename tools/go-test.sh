# Based on https://dev.to/stackdumper/setting-up-ci-for-microservices-in-monorepo-using-github-actions-5do2.

BASE_DIR=".."

GO_SERVICES=${BASE_DIR}/services/go

# Iterate over each go-service.
for SERVICE in $(ls ${GO_SERVICES}); do
  (cd "${GO_SERVICES}/${SERVICE}" && gotestsum --format dots ./...)
done
