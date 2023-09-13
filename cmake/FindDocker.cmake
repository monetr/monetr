# Look for docker from the path variable. If someone has docker for desktop installed then this might
# find that instead otherwise.
find_program(DOCKER_EXECUTABLE NAMES docker PATHS ENV PATH NO_DEFAULT_PATH)

if(DOCKER_EXECUTABLE)
  message(TRACE "Detected Docker executable at: ${DOCKER_EXECUTABLE}")
  execute_process(
    COMMAND ${DOCKER_EXECUTABLE} version --format=json
    OUTPUT_VARIABLE DOCKER_VERSION_STRING
    ERROR_VARIABLE NO-OP
    OUTPUT_STRIP_TRAILING_WHITESPACE
  )

  string(JSON DOCKER_VERSION GET "${DOCKER_VERSION_STRING}" "Client" "Version")
  if(NOT "${DOCKER_MIN_VERSION}" STREQUAL "")
    if("${DOCKER_VERSION}" VERSION_LESS "${DOCKER_MIN_VERSION}")
      message(FATAL_ERROR "Docker version ${DOCKER_VERSION} is too old. Minimum required is ${DOCKER_MIN_VERSION}")
    else()
      message(STATUS "Found Docker: ${DOCKER_EXECUTABLE} (found version \"${DOCKER_VERSION}\", minimum \"${DOCKER_MIN_VERSION}\")")
    endif()
  else()
    message(STATUS "Found Docker: ${DOCKER_EXECUTABLE} (found version \"${DOCKER_VERSION}\")")
  endif()

  # Check to see if the docker server was detected or is running.
  string(JSON SERVER_VERSION ERROR_VARIABLE NO-OP GET "${DOCKER_VERSION_STRING}" "Server" "Version")
  if ("${SERVER_VERSION}" STREQUAL "Server-Version-NOTFOUND")
    message(AUTHOR_WARNING "Docker server does not appear to be running, container features may not work properly!")
  else()
    message(TRACE "Docker server appears to be running, detected version: ${SERVER_VERSION}")
  endif()
else()
  message(WARNING "Could not find Docker")
endif()
