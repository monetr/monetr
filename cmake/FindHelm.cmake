include(GolangUtils)

set(HELM_EXECUTABLE ${GO_BIN_DIR}/helm CACHE INTERNAL "Path to the local helm tooling")
SET(HELM_VERSION "v3.15.0")
go_install(
  OUTPUT ${HELM_EXECUTABLE}
  PACKAGE "helm.sh/helm/v3/cmd/helm"
  VERSION ${HELM_VERSION}
)

