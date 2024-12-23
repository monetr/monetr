cmake_minimum_required(VERSION 3.23)

option(BUILD_THIRD_PARTY_NOTICE "Generate a third party notice as part of the build stage, requires Ruby." OFF)
option(BUILD_SIMPLE_ICONS "Embed simple-icons into the final binary, used for providing icons for transactions." ON)
option(BUILD_NOSIMD "Disable SIMD support at compile time." OFF)
option(BUILD_CONTAINER_TAG_LATEST "Tag containers with latest when building them." ON)

option(TEST_RACE "Go tests should be run with the -race flag" ON)
option(TEST_COVERAGE "Coverage data should be collected from tests" ON)
option(TEST_USE_GOTESTSUM "Wrap test execution in gotestsum" OFF)
option(TEST_SHORT "Exclude long running tests" ON)

if (CMAKE_BUILD_TYPE STREQUAL "Release")
  message(STATUS "Mode: Release")
else()
  message(STATUS "Mode: Debug")
endif()

# This makes it possible to support some other languages.
# You'll find here we have the FindX files as well as some other ones.
# This will make it so cmake looks here for some Find files.
list(PREPEND CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")
project(monetr Go)

# Get the hostname for the build.
if (WIN32)
  set(HOME $ENV{USERPROFILE})
  execute_process(COMMAND hostname OUTPUT_VARIABLE HOSTNAME OUTPUT_STRIP_TRAILING_WHITESPACE)
else()
  set(HOME $ENV{HOME})
  execute_process(COMMAND uname -n OUTPUT_VARIABLE HOSTNAME OUTPUT_STRIP_TRAILING_WHITESPACE)
endif()

# monetr requires cgo in order to be built at this time, so if gcc is not present on the system then show a warning.
find_package(Gcc REQUIRED)
find_package(Git REQUIRED)

set(DOCKER_MIN_VERSION "26.0.0")
find_package(Docker QUIET)

set(MONETR_PKG_ROOT github.com/monetr/monetr CACHE INTERNAL "Base go package for monetr.")

# The UI tasks are split out into their own file.
# Import this _after_ we have found all the packages above.
include(NodeModules)
include(GolangUtils)
add_subdirectory("${CMAKE_SOURCE_DIR}/images")
add_subdirectory("${CMAKE_SOURCE_DIR}/interface")
add_subdirectory("${CMAKE_SOURCE_DIR}/stories")
add_subdirectory("${CMAKE_SOURCE_DIR}/emails")

if(BUILD_DEPLOYMENT_TARGETS)
  include("${CMAKE_SOURCE_DIR}/cmake/deploy.cmake")
endif()

add_custom_target(
  dependencies
  DEPENDS dependencies.node_modules ${GO_MODULES}
)

if(DOCKER_EXECUTABLE)
  message(STATUS "Container engine: Docker")
  message(STATUS "  Container build targets will be enabled!")
  include("${CMAKE_SOURCE_DIR}/cmake/development.cmake")

  if(NOT "$ENV{CONTAINER_REGISTRY}" STREQUAL "")
    set(CONTAINER_REGISTRY "$ENV{CONTAINER_REGISTRY}")
  else()
    set(CONTAINER_REGISTRY "ghcr.io/monetr/monetr" "docker.io/monetr/monetr")
  endif()
  set(CONTAINER_TAGS "$(CONTAINER_VERSION)")
  if(BUILD_CONTAINER_TAG_LATEST)
    list(APPEND CONTAINER_TAGS "latest")
  endif()

  set(CONTAINER_ARGS)
  foreach(REGISTRY ${CONTAINER_REGISTRY})
    foreach(TAG ${CONTAINER_TAGS})
      list(APPEND CONTAINER_ARGS "-t" "${REGISTRY}:${TAG}")
    endforeach()
  endforeach()

  set(DOCKER_CONTAINER ${CMAKE_BINARY_DIR}/docker-container-marker.txt)
  add_custom_command(
    OUTPUT ${DOCKER_CONTAINER}
    BYPRODUCTS ${DOCKER_CONTAINER}
    COMMENT "Building monetr container image"
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    DEPENDS
      ${APP_GO_FILES}
      ${APP_UI_FILES}
      ${CMAKE_SOURCE_DIR}/Dockerfile
      ${CMAKE_SOURCE_DIR}/package.json
      ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
      ${PUBLIC_FILES}
      ${UI_CONFIG_FILES}
    COMMAND ${CMAKE_COMMAND} -E echo "-- Building monetr container: ${CONTAINER_REGISTRY} - ${CONTAINER_TAGS}"
    COMMAND ${DOCKER_EXECUTABLE} buildx build
        --iidfile ${DOCKER_CONTAINER}
        -f ${CMAKE_SOURCE_DIR}/Dockerfile
        --platform linux/amd64,linux/arm64
        --build-arg GOFLAGS=-tags=icons,simple_icons
        --build-arg RELEASE=$(RELEASE_VERSION)
        --build-arg REVISION=$(RELEASE_REVISION)
        --build-arg BUILD_HOST=${HOSTNAME}
        ${CONTAINER_ARGS}
        ${CMAKE_SOURCE_DIR}
    VERBATIM
  )
  add_custom_target(
    build.container.docker
    DEPENDS
      ${DOCKER_CONTAINER}
  )

  add_custom_target(
    build.container.docker.push
    DEPENDS
      build.container.docker
  )

  foreach(REGISTRY ${CONTAINER_REGISTRY})
    foreach(TAG ${CONTAINER_TAGS})
      # Create a target for each tag and make it a dependency of the overall push target.
      # TODO Move the configuration of tags into the cmakepresets.
      string(SHA256 PUSH_TARGET_NAME "${REGISTRY}:${TAG}")
      add_custom_target(
        push.${PUSH_TARGET_NAME}
        COMMAND ${DOCKER_EXECUTABLE} push ${REGISTRY}:${TAG}
        DEPENDS
          build.container.docker
      )
      add_dependencies(build.container.docker.push push.${PUSH_TARGET_NAME})
    endforeach()
  endforeach()
endif()

################################################################################
#                                     TESTING                                  #
################################################################################

enable_testing()
option(BUILD_TESTING "Build the testing tree." OFF)

include(GolangTestUtils)

if(BUILD_TESTING)
  if(CMAKE_Go_COMPILER)
    execute_process(
      COMMAND ${CMAKE_Go_COMPILER} run ${CMAKE_SOURCE_DIR}/cmake/find_tests.go ${CMAKE_SOURCE_DIR}/server
      OUTPUT_VARIABLE GOLANG_TEST_LIST
      OUTPUT_STRIP_TRAILING_WHITESPACE
      WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    )
    set(GOLANG_TEST_LIST "${GOLANG_TEST_LIST}" CACHE STRING "All Golang unit tests in json format." FORCE)

    add_test(
      NAME migrate.db
      COMMAND ${CMAKE_COMMAND} --build ${CMAKE_BINARY_DIR} -t development.migrate
      WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    )
    set_tests_properties(migrate.db PROPERTIES FIXTURES_SETUP DB)
    set_property(
      TEST migrate.db
      PROPERTY LABELS "bootstrap"
    )

    # If this test fails that means you need to run `go mod tidy`.
    add_test(
      NAME go.mod.tidy
      COMMAND ${CMAKE_Go_COMPILER} mod tidy --diff
      WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    )
    set_property(
      TEST go.mod.tidy
      PROPERTY LABELS "bootstrap"
    )
  else()
    message(STATUS "Go tests will not be run since Go installation could not be found.")
  endif()

  add_test(
    NAME ui.dependencies
    COMMAND ${CMAKE_COMMAND} --build ${CMAKE_BINARY_DIR} -t dependencies.node_modules
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  )
  set_tests_properties(ui.dependencies PROPERTIES FIXTURES_SETUP node_modules DEPENDS ${PNPM_EXECUTABLE})
  set_property(
    TEST ui.dependencies
    PROPERTY LABELS "bootstrap"
  )
endif()

file(GLOB_RECURSE SUB_PACKAGES RELATIVE ${CMAKE_SOURCE_DIR} "${CMAKE_SOURCE_DIR}/server/*CMakeLists.txt")
# # Loop through the files and get their directories
foreach(SUB_PACKAGE IN LISTS SUB_PACKAGES)
  get_filename_component(DIR ${SUB_PACKAGE} DIRECTORY)
  add_subdirectory(${DIR})
endforeach()

# Recursively add all of the directories in the interface package that have CMakeLists of their own.
file(
  GLOB_RECURSE SUB_PACKAGES
  RELATIVE ${CMAKE_SOURCE_DIR} "${CMAKE_SOURCE_DIR}/interface/src/*CMakeLists.txt"
)
# # Loop through the files and get their directories
foreach(SUB_PACKAGE IN LISTS SUB_PACKAGES)
  get_filename_component(DIR ${SUB_PACKAGE} DIRECTORY)
  add_subdirectory(${DIR})
endforeach()

# In order for documentation tests to work this has to be after the tests are setup.
add_subdirectory("${CMAKE_SOURCE_DIR}/docs")

# Prevent caching of this value
unset(BUILD_TESTING)
