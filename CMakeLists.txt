cmake_minimum_required(VERSION 3.23)

option(BUILD_THIRD_PARTY_NOTICE "Generate a third party notice as part of the build stage, requires Ruby." OFF)
option(BUILD_SIMPLE_ICONS "Embed simple-icons into the final binary, used for providing icons for transactions." ON)
option(BUILD_DEPLOYMENT_TARGETS "Include targets for deploying monetr using helm and kubectl" OFF)
option(BUILD_CONTAINER_TAG_LATEST "Tag containers with latest when building them." ON)

if (CMAKE_BUILD_TYPE STREQUAL "Release")
  message(STATUS "Mode: Release - Build must include third party notice")
  set(BUILD_THIRD_PARTY_NOTICE ON)
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


find_package(Git REQUIRED)

set(DOCKER_MIN_VERSION "20.0.0")
find_package(Docker QUIET)

# The UI tasks are split out into their own file.
# Import this _after_ we have found all the packages above.
include(NodeModules)
include(GolangUtils)
add_subdirectory("${CMAKE_SOURCE_DIR}/interface")
add_subdirectory("${CMAKE_SOURCE_DIR}/docs")
add_subdirectory("${CMAKE_SOURCE_DIR}/emails")
include("${CMAKE_SOURCE_DIR}/cmake/backend.cmake")

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
    COMMAND ${DOCKER_EXECUTABLE} build
        --iidfile ${DOCKER_CONTAINER}
        -f ${CMAKE_SOURCE_DIR}/Dockerfile
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

set(LICENSED_VERSION "4.4.0")
if(BUILD_THIRD_PARTY_NOTICE)
  find_package(Ruby REQUIRED)
  set(GEM_HOME ${CMAKE_BINARY_DIR}/ruby)
  set(RUBY_BIN ${GEM_HOME}/bin)
  set(LICENSED_EXECUTABLE ${RUBY_BIN}/licensed)
  file(MAKE_DIRECTORY ${RUBY_BIN})
  add_custom_command(
    OUTPUT ${LICENSED_EXECUTABLE}
    BYPRODUCTS ${GEM_HOME}
    COMMAND GEM_HOME=${GEM_HOME} gem install licensed -v ${LICENSED_VERSION}
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Setting up licensed"
  )

  set(LICENSED_CACHE ${CMAKE_BINARY_DIR}/license.cache)
  add_custom_command(
    OUTPUT ${LICENSED_CACHE}
    BYPRODUCTS ${LICENSED_CACHE}
    COMMAND GEM_HOME=${GEM_HOME} ${LICENSED_EXECUTABLE} cache
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Caching dependency licenses"
    DEPENDS
      ${LICENSED_EXECUTABLE}
      dependencies.node_modules
      ${CMAKE_SOURCE_DIR}/package.json
      ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
      ${CMAKE_SOURCE_DIR}/go.mod
      ${CMAKE_SOURCE_DIR}/go.sum
  )

  set(LICENSED_INDIVIDUAL_NOTICES
    ${LICENSED_CACHE}/monetr-API/NOTICE
    ${LICENSED_CACHE}/monetr-UI/NOTICE
  )
  add_custom_command(
    OUTPUT ${LICENSED_INDIVIDUAL_NOTICES}
    BYPRODUCTS ${LICENSED_INDIVIDUAL_NOTICES}
    COMMAND GEM_HOME=${GEM_HOME} ${LICENSED_EXECUTABLE} notices
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Gathering license information from dependencies"
    DEPENDS
      ${LICENSED_EXECUTABLE}
      ${LICENSED_CACHE}
      ${CMAKE_SOURCE_DIR}/package.json
      ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
      ${CMAKE_SOURCE_DIR}/go.mod
      ${CMAKE_SOURCE_DIR}/go.sum
  )

  set(THIRD_PARTY_NOTICE ${CMAKE_SOURCE_DIR}/server/build/NOTICE.md)
  add_custom_command(
    OUTPUT ${THIRD_PARTY_NOTICE}
    BYPRODUCTS ${THIRD_PARTY_NOTICE}
    COMMAND ${CMAKE_COMMAND} -E cat ${LICENSED_INDIVIDUAL_NOTICES} > ${THIRD_PARTY_NOTICE}
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Generating a third party notice file from dependency license data"
    DEPENDS
      ${LICENSED_INDIVIDUAL_NOTICES}
  )

  # Since we have ruby available then we want to build the third party notice file as part of the
  # binary build process.
  add_custom_command(
    OUTPUT ${MONETR_EXECUTABLE} APPEND
    DEPENDS ${THIRD_PARTY_NOTICE}
  )

  add_custom_target(
    license.cache
    DEPENDS ${LICENSED_CACHE}
  )
else()
  # If ruby is not detected, print something saying that third party notices cannot be generated.
  message(STATUS "======================================================================")
  message(STATUS "Third party notice disabled, you can enable this by passing:")
  message(STATUS "  CMAKE_OPTIONS=-DBUILD_THIRD_PARTY_NOTICE=ON")
  message(STATUS "  When you execute make monetr")
  message(STATUS "======================================================================")
endif()

add_custom_target(
  build.monetr
  DEPENDS ${MONETR_EXECUTABLE}
)

add_custom_target(
  development.migrate
  COMMAND ${CMAKE_Go_COMPILER} run ${MONETR_CLI_PKG} database migrate
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)

################################################################################
#                                     TESTING                                  #
################################################################################

enable_testing()
option(BUILD_TESTING "Build the testing tree." OFF)

option(TEST_JUNIT "Tests should produce a JUnit XML file" OFF)
option(TEST_RACE "Go tests should be run with the -race flag" ON)
option(TEST_COVERAGE "Coverage data should be collected from tests" OFF)
option(TEST_USE_GOTESTSUM "Wrap test execution in gotestsum" OFF)

include(GolangTestUtils)

if(BUILD_TESTING)
  if(CMAKE_Go_COMPILER)
    add_test(
      NAME migrate.db
      COMMAND ${CMAKE_COMMAND} --build ${CMAKE_BINARY_DIR} -t development.migrate
      WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    )
    set_tests_properties(migrate.db PROPERTIES FIXTURES_SETUP DB)
  else()
    message(STATUS "Go tests will not be run since Go installation could not be found.")
  endif()

  add_test(
    NAME ui.dependencies
    COMMAND ${CMAKE_COMMAND} --build ${CMAKE_BINARY_DIR} -t dependencies.node_modules
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  )
  set_tests_properties(ui.dependencies PROPERTIES FIXTURES_SETUP node_modules DEPENDS ${PNPM_EXECUTABLE})
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
