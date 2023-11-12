
################################################################################
# Helm stuff and deployments                                                   #
################################################################################
set(DEPLOY_NAMESPACE)
if(NOT DEPLOY_NAMESPACE)
  if("$ENV{DEPLOY_NAMESPACE}" STREQUAL "")
    message(FATAL_ERROR "DEPLOY_NAMESPACE must be defined in order to generate and deploy to Kubernetes. Add -DDEPLOY_NAMESPACE=$NAMESPACE to your CMake configure command or set the DEPLOY_NAMESPACE environment variable in your env.")
  endif()

  set(DEPLOY_NAMESPACE "$ENV{DEPLOY_NAMESPACE}")
endif()

set(DEPLOY_VERSION)
if(NOT DEPLOY_VERSION)
  if("$ENV{DEPLOY_VERSION}" STREQUAL "")
    message(FATAL_ERROR "DEPLOY_VERSION must be defined in order to generate and deploy to Kubernetes. Add -DDEPLOY_VERSION=$VERSION to your CMake configure command or set the DEPLOY_VERSION environment variable in your env.")
  endif()

  set(DEPLOY_VERSION "$ENV{DEPLOY_VERSION}")
endif()

set(ENVIRONMENT)
if(NOT ENVIRONMENT)
  if("$ENV{ENVIRONMENT}" STREQUAL "")
    message(FATAL_ERROR "ENVIRONMENT must be defined in order to generate and deploy to Kubernetes. Add -DENVIRONMENT=$ENV to your CMake configure command or set the ENVIRONMENT environment variable in your env.")
  endif()

  set(ENVIRONMENT "$ENV{ENVIRONMENT}")
endif()

message(STATUS "======================================================================")
message(STATUS "Deployment configuration:")
message(STATUS "  Environment: ${ENVIRONMENT}")
message(STATUS "  Namespace:   ${DEPLOY_NAMESPACE}")
message(STATUS "  Version:     ${DEPLOY_VERSION}")
message(STATUS "======================================================================")

set(KUBECTL_MIN_VERSION "1.21.0")
find_package(Kubectl REQUIRED)
find_package(Helm REQUIRED)


string(TOLOWER "${ENVIRONMENT}" ENVIRONMENT_LOWER)
set(GENERATED_YAML "${CMAKE_BINARY_DIR}/generated/${ENVIRONMENT_LOWER}")
file(GLOB HELM_FILES
  "${CMAKE_SOURCE_DIR}/Chart.yaml"
  "${CMAKE_SOURCE_DIR}/values.yaml"
  "${CMAKE_SOURCE_DIR}/values.${ENVIRONMENT_LOWER}.yaml"
  "${CMAKE_SOURCE_DIR}/templates/*"
)

set(KUBERNETES_SPLIT_YAML_EXECUTABLE ${GO_BIN_DIR}/kubernetes-split-yaml)
go_install(
  OUTPUT ${KUBERNETES_SPLIT_YAML_EXECUTABLE}
  PACKAGE "github.com/elliotcourant/kubernetes-split-yaml"
  VERSION "3c77b924132b7ac914dc156eeea2e1db47541bb0"
)

string(REPLACE "v" "" DEPLOY_VERSION "${DEPLOY_VERSION}")
add_custom_command(
  OUTPUT ${GENERATED_YAML}
  BYPRODUCTS ${GENERATED_YAML}
  COMMAND ${CMAKE_COMMAND} -E remove_directory "${GENERATED_YAML}"
  COMMAND ${HELM_EXECUTABLE} template monetr ${CMAKE_SOURCE_DIR} --dry-run --set image.tag=${DEPLOY_VERSION} --values=${CMAKE_SOURCE_DIR}/values.${ENVIRONMENT_LOWER}.yaml | ${KUBERNETES_SPLIT_YAML_EXECUTABLE} --outdir ${GENERATED_YAML} -
  COMMAND ${CMAKE_COMMAND} -E echo "Finished generating deployment yaml's for ${ENVIRONMENT_LOWER}, they are available at: ${GENERATED_YAML}"
  COMMENT "Generating deployment yamls for environment ${ENVIRONMENT_LOWER}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  VERBATIM
  DEPENDS
    ${HELM_FILES}
    ${HELM_EXECUTABLE}
    ${KUBERNETES_SPLIT_YAML_EXECUTABLE}
)

add_custom_target(
  build.yaml
  DEPENDS ${GENERATED_YAML}
)

add_custom_target(
  deploy.dry
  COMMAND ${KUBECTL_EXECUTABLE} apply -f ${GENERATED_YAML} -n ${DEPLOY_NAMESPACE} --dry-run=server
  DEPENDS ${GENERATED_YAML}
)

add_custom_target(
  deploy.apply
  COMMAND ${KUBECTL_EXECUTABLE} apply -f ${GENERATED_YAML} -n ${DEPLOY_NAMESPACE}
  COMMAND ${KUBECTL_EXECUTABLE} rollout status deploy/monetr -n ${DEPLOY_NAMESPACE} --timeout=600s
  DEPENDS ${GENERATED_YAML}
)
