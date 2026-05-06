if(NOT METADATA_FILE)
  message(FATAL_ERROR "METADATA_FILE not provided")
endif()
if(NOT EXISTS "${METADATA_FILE}")
  message(FATAL_ERROR "Metadata file not found: ${METADATA_FILE}")
endif()
if(NOT COSIGN_EXECUTABLE)
  message(FATAL_ERROR "COSIGN_EXECUTABLE not provided")
endif()
if(NOT REGISTRIES)
  message(FATAL_ERROR "REGISTRIES not provided")
endif()
if(NOT SIGN_CERT_IDENTITY_REGEX)
  message(FATAL_ERROR "SIGN_CERT_IDENTITY_REGEX not provided")
endif()
if(NOT SIGN_OIDC_ISSUER)
  message(FATAL_ERROR "SIGN_OIDC_ISSUER not provided")
endif()

file(READ "${METADATA_FILE}" METADATA_JSON)
string(JSON DIGEST ERROR_VARIABLE JSON_ERR GET "${METADATA_JSON}" "containerimage.digest")
if(JSON_ERR)
  message(FATAL_ERROR "Failed to extract containerimage.digest from ${METADATA_FILE}: ${JSON_ERR}")
endif()

set(COSIGN_ARGS verify
  --certificate-identity-regexp ${SIGN_CERT_IDENTITY_REGEX}
  --certificate-oidc-issuer ${SIGN_OIDC_ISSUER}
)
if(NOT SIGN_TLOG_UPLOAD)
  list(APPEND COSIGN_ARGS --insecure-ignore-tlog)
endif()

string(REPLACE "," ";" REGISTRY_LIST "${REGISTRIES}")

foreach(REGISTRY IN LISTS REGISTRY_LIST)
  message(STATUS "cosign verify ${REGISTRY}@${DIGEST}")
  execute_process(
    COMMAND ${COSIGN_EXECUTABLE} ${COSIGN_ARGS} "${REGISTRY}@${DIGEST}"
    RESULT_VARIABLE RES
  )
  if(NOT RES EQUAL 0)
    message(FATAL_ERROR "cosign verify failed for ${REGISTRY}@${DIGEST}")
  endif()
endforeach()
