find_program(KUBECTL_EXECUTABLE NAMES kubectl)

if(KUBECTL_EXECUTABLE)
  execute_process(
    COMMAND ${KUBECTL_EXECUTABLE} version --output=json
    OUTPUT_VARIABLE KUBECTL_VERSION_STRING
    ERROR_VARIABLE NO-OP
    OUTPUT_STRIP_TRAILING_WHITESPACE
  )

  string(JSON KUBECTL_VERSION GET "${KUBECTL_VERSION_STRING}" "clientVersion" "gitVersion")
  string(REPLACE "v" "" KUBECTL_VERSION "${KUBECTL_VERSION}")
  if(NOT "${KUBECTL_MIN_VERSION}" STREQUAL "")
    if("${KUBECTL_VERSION}" VERSION_LESS "${KUBECTL_MIN_VERSION}")
      message(FATAL_ERROR "kubectl version ${KUBECTL_VERSION} is too old. Minimum required is ${KUBECTL_MIN_VERSION}")
    else()
      message(STATUS "Found kubectl: ${KUBECTL_EXECUTABLE} (found version \"${KUBECTL_VERSION}\", minimum \"${KUBECTL_MIN_VERSION}\")")
    endif()
  else()
    message(STATUS "Found kubectl: ${KUBECTL_EXECUTABLE} (found version \"${KUBECTL_VERSION}\")")
  endif()
else()
  message(FATAL_ERROR "Could not find kubectl")
endif()
