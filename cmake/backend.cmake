set(GO_MIN_VERSION "1.21.0")

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

set(MONETR_CLI_PKG github.com/monetr/monetr/server/cmd)

set(GO_SRC_DIR "${CMAKE_SOURCE_DIR}/server")
file(GLOB_RECURSE ALL_GO_FILES
  "${GO_SRC_DIR}/*.go"
  "${GO_SRC_DIR}/*.sql"
)
file(GLOB_RECURSE APP_GO_FILES
  "${GO_SRC_DIR}/*.go"
  "${GO_SRC_DIR}/*.sql"
)
list(FILTER APP_GO_FILES EXCLUDE REGEX ".+_test\\.go")

if((WIN32) OR ("$ENV{GOOS}" STREQUAL "windows"))
  set(MONETR_EXECUTABLE ${CMAKE_BINARY_DIR}/monetr.exe)
else()
  set(MONETR_EXECUTABLE ${CMAKE_BINARY_DIR}/monetr)
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
  COMMAND ${CMAKE_Go_COMPILER} build
          -tags="${MONETR_EXECUTABLE_TAGS}"
          -ldflags="${MONETR_EXECUTABLE_LD_FLAGS}"
          -o ${MONETR_EXECUTABLE}
          ${MONETR_CLI_PKG}
  COMMENT "Building monetr binary"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  # These depends can change based on some of the build options in the main file.
  DEPENDS
    ${GO_MODULES}
    ${APP_GO_FILES}
    build.interface
    build.email
    download.legal
    sourcemaps.golang
)

if(BUILD_SIMPLE_ICONS)
  message(STATUS "Simple Icons enabled, icons will be embedded at compile time")
  add_custom_command(
    OUTPUT ${MONETR_EXECUTABLE} APPEND
    DEPENDS download.simple-icons
  )
endif()
