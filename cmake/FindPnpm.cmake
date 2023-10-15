set(LOCAL_PNPM_EXECUTABLE "${NODE_BIN}/pnpm")
set(PNPM_EXECUTABLE ${LOCAL_PNPM_EXECUTABLE} CACHE INTERNAL "Path to the local version of pnpm's executable")

# TODO This is either a huge version mismatch between what I have and what Tim is seeing on Windows in how npm behaves
# or windows legit just handles the prefix arg differently?
if(WIN32)
  set(NPM_PREFIX ${CMAKE_BINARY_DIR}/node/bin)
else()
  set(NPM_PREFIX ${CMAKE_BINARY_DIR}/node)
endif()
file(MAKE_DIRECTORY "${NPM_PREFIX}")

add_custom_command(
  OUTPUT ${PNPM_EXECUTABLE}
  BYPRODUCTS ${LOCAL_PNPM_EXECUTABLE}
  COMMAND ${NPM_EXECUTABLE} install --no-fund --no-audit --global --prefix ${NPM_PREFIX} pnpm
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  COMMENT "Setting up a local version of pnpm, this will not affect the host system"
)

# Try to find pnpm on the host machine.
# find_program(HOST_PNPM_EXECUTABLE NAMES pnpm PATHS ENV PATH NO_DEFAULT_PATH)
#
#
# function(use_pnpm_target USE_LOCAL)
#   # If the pnpm executable is equal to our local one then we need to keep the local target, or if we have specified use
#   # local.
#   if(("${HOST_PNPM_EXECUTABLE}" STREQUAL "${LOCAL_PNPM_EXECUTABLE}") OR USE_LOCAL)
#     set(PNPM_EXECUTABLE ${LOCAL_PNPM_EXECUTABLE} CACHE INTERNAL "Path to the local version of pnpm's executable")
#     file(MAKE_DIRECTORY "${CMAKE_BINARY_DIR}/node")
#     add_custom_command(
#       OUTPUT ${PNPM_EXECUTABLE}
#       BYPRODUCTS ${LOCAL_PNPM_EXECUTABLE}
#       COMMAND ${NPM_EXECUTABLE} install --no-fund --no-audit --global --prefix  ${CMAKE_BINARY_DIR}/node pnpm
#       WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
#       COMMENT "Setting up pnpm"
#     )
#   else()
#     set(PNPM_EXECUTABLE ${HOST_PNPM_EXECUTABLE} CACHE INTERNAL "Path to the local version of pnpm's executable")
#     add_executable(${PNPM_EXECUTABLE} IMPORTED GLOBAL)
#   endif()
# endfunction()
#
# # If we have a pnpm on the host.
# if(HOST_PNPM_EXECUTABLE)
#   # Check to see what version it is.
#   execute_process(
#     COMMAND ${HOST_PNPM_EXECUTABLE} -v
#     OUTPUT_VARIABLE PNPM_VERSION_STRING
#     OUTPUT_STRIP_TRAILING_WHITESPACE
#   )
#
#   # This might not even be necessary since the format is already 3 numbers?
#   string(REGEX REPLACE "(\\d+\\.\\d+\\.\\d+)" "\\1" PNPM_VERSION "${PNPM_VERSION_STRING}")
#
#   # If we are setting a minimum version then assert that the version that is installed is greater.
#   if(NOT "${PNPM_MIN_VERSION}" STREQUAL "")
#     # If we have set a min version and the version on the host is less than that then we need to provision a local one
#     # instead. This way we don't need to make the host upgrade.
#     if("${PNPM_VERSION}" VERSION_LESS "${PNPM_MIN_VERSION}")
#       # Let them know we found the host one, but that it's too old and that a local one will be used instead.
#       message(STATUS "Found pnpm: ${HOST_PNPM_EXECUTABLE} but version is too old. (found version \"${PNPM_VERSION}\", minimum \"${PNPM_MIN_VERSION}\")")
#       message(STATUS "  pnpm will be installed locally instead. This will not overwrite the host version.")
#       # Create the target to setup pnpm locally.
#       use_pnpm_target(ON)
#     else()
#       # The host version isn't too old and we can use it!
#       message(STATUS "Found pnpm: ${HOST_PNPM_EXECUTABLE} (found version \"${PNPM_VERSION}\", minimum \"${PNPM_MIN_VERSION}\")")
#       # But because we have targets that have the ${HOST_PNPM_EXECUTABLE} as a dependency, we need to create a bogus target
#       # or command that is basically a no-op. This way we don't get errors about "target for ${HOST_PNPM_EXECUTABLE} does not
#       # exist".
#       use_pnpm_target(OFF)
#     endif()
#   else()
#     # We haven't set a minimum version on pnpm, and the host has one installed. So use that version.
#     message(STATUS "Found pnpm: ${HOST_PNPM_EXECUTABLE} (found version \"${PNPM_VERSION}\")")
#     # But because we have targets that have the ${HOST_PNPM_EXECUTABLE} as a dependency, we need to create a bogus target
#     # or command that is basically a no-op. This way we don't get errors about "target for ${HOST_PNPM_EXECUTABLE} does not
#     # exist".
#     use_pnpm_target(OFF)
#   endif()
# else()
#   # The host does not have pnpm installed at all, so we can just install our own locally.
#   message(STATUS "Not found: pnpm - A local version will be installed instead.")
#   use_pnpm_target(ON)
# endif()
