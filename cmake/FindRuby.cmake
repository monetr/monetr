find_program(RUBY_EXECUTABLE NAMES ruby)

if(RUBY_EXECUTABLE)
  execute_process(
    COMMAND ${RUBY_EXECUTABLE} --version
    OUTPUT_VARIABLE RUBY_VERSION_STRING
    OUTPUT_STRIP_TRAILING_WHITESPACE
  )

  string(REGEX REPLACE "ruby ([0-9]+\\.[0-9]+\\.[0-9]+).*" "\\1" RUBY_VERSION "${RUBY_VERSION_STRING}")

  if(NOT "${RUBY_MIN_VERSION}" STREQUAL "")
    if("${RUBY_VERSION}" VERSION_LESS "${RUBY_MIN_VERSION}")
      message(FATAL_ERROR "Ruby version ${RUBY_VERSION} is too old. Minimum required is ${RUBY_MIN_VERSION}")
    else()
      message(STATUS "Found Ruby: ${RUBY_EXECUTABLE} (found version \"${RUBY_VERSION}\", minimum \"${RUBY_MIN_VERSION}\")")
    endif()
  else()
    message(STATUS "Found Ruby: ${RUBY_EXECUTABLE} (found version \"${RUBY_VERSION}\")")
  endif()

else()
  message(FATAL_ERROR "Could not find Ruby")
endif()
