find_program(NODE_EXECUTABLE NAMES node)

if(NODE_EXECUTABLE)
  execute_process(
    COMMAND ${NODE_EXECUTABLE} --version
    OUTPUT_VARIABLE NODE_VERSION_STRING
    OUTPUT_STRIP_TRAILING_WHITESPACE
  )

  string(REGEX REPLACE "v([0-9]+\\.[0-9]+\\.[0-9]+)" "\\1" NODE_VERSION "${NODE_VERSION_STRING}")

  if(NOT "${NODE_MIN_VERSION}" STREQUAL "")
    if("${NODE_VERSION}" VERSION_LESS "${NODE_MIN_VERSION}")
      message(FATAL_ERROR "Node version ${NODE_VERSION} is too old. Minimum required is ${NODE_MIN_VERSION}")
    else()
      message(STATUS "Found node: ${NODE_EXECUTABLE} (found version \"${NODE_VERSION}\", minimum \"${NODE_MIN_VERSION}\")")
    endif()
  else()
    message(STATUS "Found node: ${NODE_EXECUTABLE} (found version \"${NODE_VERSION}\")")
  endif()
else()
  message(FATAL_ERROR "Could not find Node")
endif()
