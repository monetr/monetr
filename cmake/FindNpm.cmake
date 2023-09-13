find_program(NPM_EXECUTABLE NAMES npm)

if(NPM_EXECUTABLE)
  execute_process(
    COMMAND ${NPM_EXECUTABLE} -v
    OUTPUT_VARIABLE NPM_VERSION
    OUTPUT_STRIP_TRAILING_WHITESPACE
  )

  if(NOT "${NPM_MIN_VERSION}" STREQUAL "")
    if("${NPM_VERSION}" VERSION_LESS "${NPM_MIN_VERSION}")
      message(FATAL_ERROR "npm version ${NPM_VERSION} is too old. Minimum required is ${NPM_MIN_VERSION}")
    else()
      message(STATUS "Found npm: ${NPM_EXECUTABLE} (found version \"${NPM_VERSION}\", minumum \"${NPM_MIN_VERSION}\")")
    endif()
  else()
    message(STATUS "Found npm: ${NPM_EXECUTABLE} (found version \"${NPM_VERSION}\")")
  endif()
else()
  message(FATAL_ERROR "Could not find npm")
endif()
