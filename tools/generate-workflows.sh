# Based on https://dev.to/stackdumper/setting-up-ci-for-microservices-in-monorepo-using-github-actions-5do2.

BASE_DIR=".."
# Prefix for deleting and generating workflow files.
GEN_PREFIX="_gen_"
# Directory for GitHub workflows.
WORKFLOW_DIR=${BASE_DIR}/.github/workflows
GO_SERVICES=${BASE_DIR}/services/go

# Remove all previously generated workflows.
rm ${WORKFLOW_DIR}/${GEN_PREFIX}*

# Read the go workflow template.
GO_WORKFLOW_TEMPLATE=$(cat ${BASE_DIR}/tools/workflow-template-go.yaml)

# Iterate over each go-service.
for SERVICE in $(ls ${GO_SERVICES}); do
  echo "generating workflow for go-service ${SERVICE}..."
  # Replace workflow name.
  WORKFLOW=$(echo "${GO_WORKFLOW_TEMPLATE}" | sed "s/{{SERVICE}}/${SERVICE}/g")
  # Save workflow.
  echo "${WORKFLOW}" > ${WORKFLOW_DIR}/${GEN_PREFIX}go_${SERVICE}.yaml
done
