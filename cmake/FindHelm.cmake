set(HELM_VERSION "v3.15.1")
if (WIN32) 
  set(HELM_OS "windows")
  set(HELM_ARCH "amd64") # arm not supported on windows
  set(HELM_HASH "8ebe6d353f0fbc7e51861a676ba1c14af9efb3443ae2c78eb91946a756b93a9a")
elseif (APPLE)
  set(HELM_OS "darwin")
  if("${CMAKE_HOST_SYSTEM_PROCESSOR}" STREQUAL "x86_64")
    set(HELM_ARCH "amd64")
    set(HELM_HASH "5fdc60e090d183113f9fa0ae9dd9d12f0c1462b9ded286370f84e340f84bd676")
  elseif("${CMAKE_HOST_SYSTEM_PROCESSOR}" STREQUAL "arm64")
    set(HELM_ARCH "arm64")
    set(HELM_HASH "4b04ede5ab9bb226c9b198c94ce12818f0b0e302193defd66970b45fc341f6e7")
  else()
    message(FATAL "Unrecognized system architecture for MacOS: ${CMAKE_HOST_SYSTEM_PROCESSOR}")
  endif()
elseif (UNIX)
  set(HELM_OS "linux")
  if("${CMAKE_HOST_SYSTEM_PROCESSOR}" STREQUAL "x86_64")
    set(HELM_ARCH "amd64")
    set(HELM_HASH "7b20e7791c04ea71e7fe0cbe11f1a8be4a55a692898b57d9db28f3b0c1d52f11")
  elseif("${CMAKE_HOST_SYSTEM_PROCESSOR}" STREQUAL "arm64")
    set(HELM_ARCH "arm64")
    set(HELM_HASH "b4c5519b18f01dd2441f5e09497913dc1da1a1eec209033ae792a8d45b9e0e86")
  else()
    message(FATAL "Unrecognized system architecture for Unix: ${CMAKE_HOST_SYSTEM_PROCESSOR}")
  endif()
endif()

set(HELM_TEMP_DIR ${CMAKE_BINARY_DIR}/tools/temp/helm)
set(HELM_TEMP_ARCHIVE ${HELM_TEMP_DIR}/helm.tar.gz)
file(MAKE_DIRECTORY ${HELM_TEMP_DIR})
if(NOT EXISTS ${HELM_TEMP_ARCHIVE})
  message(STATUS "Downloading helm ${HELM_VERSION} (${HELM_OS}/${HELM_ARCH})")
endif()
file(DOWNLOAD "https://get.helm.sh/helm-${HELM_VERSION}-${HELM_OS}-${HELM_ARCH}.tar.gz" ${HELM_TEMP_ARCHIVE}
  TIMEOUT 30
  TLS_VERIFY ON
  EXPECTED_HASH "SHA256=${HELM_HASH}"
)
file(ARCHIVE_EXTRACT INPUT ${HELM_TEMP_ARCHIVE}
  DESTINATION ${HELM_TEMP_DIR}
)
set(HELM_EXECUTABLE "${HELM_TEMP_DIR}/${HELM_OS}-${HELM_ARCH}/helm" CACHE INTERNAL "Path to the local helm tooling")
