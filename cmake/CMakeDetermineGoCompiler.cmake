if(NOT CMAKE_Go_COMPILER)
  if(NOT $ENV{GO_COMPILER} STREQUAL "")
    get_filename_component(CMAKE_Go_COMPILER_INIT $ENV{GO_COMPILER} PROGRAM PROGRAM_ARGS CMAKE_Go_FLAGS_ENV_INIT)

    if(CMAKE_Go_FLAGS_ENV_INIT)
      set(CMAKE_Go_COMPILER_ARG1 "${CMAKE_Go_FLAGS_ENV_INIT}" CACHE STRING "First argument to Go compiler")
    endif()

    if(NOT EXISTS ${CMAKE_Go_COMPILER_INIT})
      message(SEND_ERROR "Could not find compiler set in environment variable GO_COMPILER:\n$ENV{GO_COMPILER}.")
    endif()
  endif()

  set(Go_BIN_PATH
    $ENV{GOPATH}
    $ENV{GOROOT}
    $ENV{GOROOT}/../bin
    $ENV{GO_COMPILER}
    /usr/bin
    /usr/local/bin
  )

  if(CMAKE_Go_COMPILER_INIT)
    set(CMAKE_Go_COMPILER ${CMAKE_Go_COMPILER_INIT} CACHE PATH "Go Compiler")
  else()
    find_program(CMAKE_Go_COMPILER NAMES go)

    if(CMAKE_Go_COMPILER)
      execute_process(
        COMMAND ${CMAKE_Go_COMPILER} version
        OUTPUT_VARIABLE GO_VERSION_STRING
        OUTPUT_STRIP_TRAILING_WHITESPACE
      )

      string(REGEX REPLACE "go version go([0-9]+\\.[0-9]+(\\.[0-9]+)?) .*" "\\1" GO_VERSION "${GO_VERSION_STRING}")

      if(NOT "${GO_MIN_VERSION}" STREQUAL "")
        if("${GO_VERSION}" VERSION_LESS "${GO_MIN_VERSION}")
          message(FATAL_ERROR "Go version ${GO_VERSION} is too old. Minimum required is ${GO_MIN_VERSION}")
        else()
          message(STATUS "Found Go: ${CMAKE_Go_COMPILER} (found version \"${GO_VERSION}\", minimum \"${GO_MIN_VERSION}\")")
        endif()
      else()
        message(STATUS "Found Go: ${CMAKE_Go_COMPILER} (found version \"${GO_VERSION}\")")
      endif()
    else()
      message(FATAL_ERROR "Could not find Go")
    endif()
  endif()

endif()

mark_as_advanced(CMAKE_Go_COMPILER)

configure_file(
  ${CMAKE_CURRENT_SOURCE_DIR}/cmake/CMakeGoCompiler.cmake.in
  ${CMAKE_PLATFORM_INFO_DIR}/CMakeGoCompiler.cmake @ONLY
)

set(CMAKE_Go_COMPILER_ENV_VAR "GO_COMPILER")
