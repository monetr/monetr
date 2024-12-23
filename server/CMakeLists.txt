set(GO_MIN_VERSION "1.23.0")

# Default to amd64 if the go arch is not specified.
if("${GOARCH}" STREQUAL "")
  set(GOARCH "$ENV{GOARCH}")
endif()
if("${GOARCH}" STREQUAL "")
  set(GOARCH "amd64")
endif()


# GO_MODULES is only a marker file to indicate that modules have actually been fetched.
set(GO_MODULES ${CMAKE_BINARY_DIR}/go-dependencies-marker.txt)
add_custom_command(
  OUTPUT ${GO_MODULES}
  BYPRODUCTS ${GO_MODULES}
  COMMAND ${CMAKE_Go_COMPILER} mod download
  COMMAND ${CMAKE_COMMAND} -E touch ${GO_MODULES}
  COMMENT "Installing go dependencies"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  DEPENDS
    ${CMAKE_SOURCE_DIR}/go.mod
    ${CMAKE_SOURCE_DIR}/go.sum
)

set(MONETR_CLI_PKG ${MONETR_PKG_ROOT}/server/cmd CACHE INTERNAL "")

set(GO_SRC_DIR "${CMAKE_SOURCE_DIR}/server")
file(GLOB_RECURSE ALL_GO_FILES
  "${GO_SRC_DIR}/*.go"
  "${GO_SRC_DIR}/*.s"
  "${GO_SRC_DIR}/*.sql"
)
file(GLOB_RECURSE APP_GO_FILES
  "${GO_SRC_DIR}/*.go"
  "${GO_SRC_DIR}/*.s"
  "${GO_SRC_DIR}/*.sql"
)
list(FILTER APP_GO_FILES EXCLUDE REGEX ".+_test\\.go")

if((WIN32) OR ("$ENV{GOOS}" STREQUAL "windows"))
  set(MONETR_EXECUTABLE ${CMAKE_BINARY_DIR}/monetr.exe)
else()
  set(MONETR_EXECUTABLE ${CMAKE_BINARY_DIR}/monetr)
endif()

set(GO_VARIABLES "CGO_ENABLED=1" "GOARCH=${GOARCH}")
# I don't have anything setup for doing cross compilation on macOS or Windows
# yet.
if(LINUX)
  # If we are cross compiling then we need to tell Golang which gcc to use. For
  # arm64 at least we need to use the one below.
  set(ARM64_GCC "aarch64-linux-gnu-gcc")
  # set(AMD64_GCC "x86_64-linux-gnu-gcc")
  if("${GOARCH}" STREQUAL "arm64")
    list(APPEND GO_VARIABLES "CC=${ARM64_GCC}")
    message(STATUS "Cross-Compiling for ARM64, using alternative gcc: ${ARM64_GCC}")
  # elseif("${GOARCH}" STREQUAL "amd64")
  #   list(APPEND GO_VARIABLES "CC=${AMD64_GCC}")
  #   message(STATUS "Cross-Compiling for AMD64, using alternative gcc: ${AMD64_GCC}")
  endif()
endif()

set(MONETR_GO_TAGS "icons")
if(BUILD_SIMPLE_ICONS)
  list(APPEND MONETR_GO_TAGS "simple_icons")
endif()
if(BUILD_NOSIMD)
  list(APPEND MONETR_GO_TAGS "nosimd")
endif()
set(MONETR_LD_FLAGS "-s" "-w")

# If the hostname variable is present then add that to the LD flags.
if(HOSTNAME)
  list(APPEND MONETR_LD_FLAGS "-X" "main.buildHost=${HOSTNAME}")
endif()

# This is a bit weird, but these commands come from make and are basically environment variables.
list(APPEND MONETR_LD_FLAGS "-X" "main.release=$(RELEASE_VERSION)")
list(APPEND MONETR_LD_FLAGS "-X" "main.revision=$(RELEASE_REVISION)")
list(APPEND MONETR_LD_FLAGS "-X" "main.buildTime=$(BUILD_TIME)")

# Detect if we are building inside a container, if we are make sure to set the build type LDFLAG.
if(NOT WIN32)
  if (EXISTS "/proc/1/cgroup")
    file(READ "/proc/1/cgroup" CONTAINER_DETECTION)
    if ("${CONTAINER_DETECTION}" MATCHES "docker")
      list(APPEND MONETR_LD_FLAGS "-X" "main.buildType=container")
    endif()
  endif()
endif()

list(JOIN MONETR_GO_TAGS "," MONETR_EXECUTABLE_TAGS)
string(REPLACE " " ";" MONETR_EXECUTABLE_LD_FLAGS "${MONETR_LD_FLAGS}")
add_custom_command(
  OUTPUT ${MONETR_EXECUTABLE}
  BYPRODUCTS ${MONETR_EXECUTABLE}
  COMMAND ${GO_VARIABLES} ${CMAKE_Go_COMPILER} build
          -tags="${MONETR_EXECUTABLE_TAGS}"
          -ldflags="${MONETR_EXECUTABLE_LD_FLAGS}"
          -o ${MONETR_EXECUTABLE}
          ${MONETR_CLI_PKG}
  COMMENT "Building monetr binary (using ${GO_VARIABLES})"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  # These depends can change based on some of the build options in the main file.
  DEPENDS
    ${GO_MODULES}
    ${APP_GO_FILES}
    build.interface
    build.email
    sourcemaps.golang
)

if(BUILD_SIMPLE_ICONS)
  if (NOT SIMPLE_ICONS_MESSAGE)
    message(STATUS "Simple Icons enabled, icons will be embedded at compile time")
    set(SIMPLE_ICONS_MESSAGE ON CACHE INTERNAL "Simple icons message has already been shown, dont show again.")
  endif()
  add_custom_command(
    OUTPUT ${MONETR_EXECUTABLE} APPEND
    DEPENDS download.simple-icons
  )
endif()

set(LICENSED_VERSION "5.0.0")
if(BUILD_THIRD_PARTY_NOTICE)
  find_package(Ruby REQUIRED)
  set(GEM_HOME ${CMAKE_BINARY_DIR}/ruby)
  set(RUBY_BIN ${GEM_HOME}/bin)
  set(LICENSED_EXECUTABLE ${RUBY_BIN}/licensed)
  file(MAKE_DIRECTORY ${RUBY_BIN})
  add_custom_command(
    OUTPUT ${LICENSED_EXECUTABLE}
    BYPRODUCTS ${GEM_HOME}
    COMMAND ${CMAKE_COMMAND} -E env GEM_HOME=${GEM_HOME} LANG=en_US.UTF-8 -- gem install licensed -v ${LICENSED_VERSION}
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Setting up licensed"
  )

  set(LICENSED_CACHE ${CMAKE_BINARY_DIR}/license.cache)
  add_custom_command(
    OUTPUT ${LICENSED_CACHE}
    BYPRODUCTS ${LICENSED_CACHE}
    COMMAND ${CMAKE_COMMAND} -E env GEM_HOME=${GEM_HOME} LANG=en_US.UTF-8 -- ${LICENSED_EXECUTABLE} cache
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
    COMMAND ${CMAKE_COMMAND} -E env GEM_HOME=${GEM_HOME} LANG=en_US.UTF-8 -- ${LICENSED_EXECUTABLE} notices
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    COMMENT "Gathering license information from dependencies"
    DEPENDS
      download.simple-icons
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
endif()


add_custom_target(
  build.monetr
  DEPENDS ${MONETR_EXECUTABLE}
)
