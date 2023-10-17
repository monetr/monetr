set(GO_BIN_DIR ${CMAKE_BINARY_DIR}/go/bin)
file(MAKE_DIRECTORY ${GO_BIN_DIR})

function(go_install)
  # Initialize variables
  set(OUTPUT)
  set(PACKAGE)
  set(VERSION latest)

  # Parse arguments
  set(args ${ARGN})
  while(args)
    list(GET args 0 arg)
    list(REMOVE_AT args 0)

    if("${arg}" STREQUAL "OUTPUT")
      list(GET args 0 OUTPUT)
      list(REMOVE_AT args 0)
    elseif("${arg}" STREQUAL "PACKAGE")
      list(GET args 0 PACKAGE)
      list(REMOVE_AT args 0)
    elseif("${arg}" STREQUAL "VERSION")
      list(GET args 0 VERSION)
      list(REMOVE_AT args 0)
    endif()
  endwhile()

  # If version was not specified default to the latest version of the package.
  if ("${VERSION}" STREQUAL "")
    set(VERSION "latest")
  endif()

  add_custom_command(
    OUTPUT ${OUTPUT}
    BYPRODUCTS ${OUTPUT}
    COMMAND ${CMAKE_COMMAND} -E env GOBIN=${GO_BIN_DIR} ${CMAKE_Go_COMPILER} install ${PACKAGE}@${VERSION}
    COMMENT "Setting up ${PACKAGE}@${VERSION} for local development"
  )
endfunction()

# mockgen for local development
set(MOCKGEN_EXECUTABLE ${GO_BIN_DIR}/mockgen)
set(MOCKGEN_VERSION "v1.6.0")
go_install(
  OUTPUT ${MOCKGEN_EXECUTABLE}
  PACKAGE "github.com/golang/mock/mockgen"
  VERSION ${MOCKGEN_VERSION}
)
