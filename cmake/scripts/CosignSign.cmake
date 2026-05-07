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
if(NOT DIGEST_FILE)
  message(FATAL_ERROR "DIGEST_FILE not provided")
endif()

file(READ "${METADATA_FILE}" METADATA_JSON)
string(JSON DIGEST ERROR_VARIABLE JSON_ERR GET "${METADATA_JSON}" "containerimage.digest")
if(JSON_ERR)
  message(FATAL_ERROR "Failed to extract containerimage.digest from ${METADATA_FILE}: ${JSON_ERR}")
endif()

# Write digest to a stable file so downstream consumers don't need jq to parse
# the buildx metadata JSON.
file(WRITE "${DIGEST_FILE}" "${DIGEST}")

# When running under GitHub Actions, also expose the digest as a step output
# (digest=<value> appended to $GITHUB_OUTPUT). Subsequent steps can then
# reference it as ${{ steps.<id>.outputs.digest }} without reading the file.
if(DEFINED ENV{GITHUB_OUTPUT})
  file(APPEND "$ENV{GITHUB_OUTPUT}" "digest=${DIGEST}\n")
endif()

set(COSIGN_ARGS sign --yes)
if(SIGN_RECURSIVE)
  list(APPEND COSIGN_ARGS --recursive)
endif()
if(NOT SIGN_TLOG_UPLOAD)
  if(NOT SIGNING_CONFIG_FILE)
    message(FATAL_ERROR "SIGNING_CONFIG_FILE not provided (required when SIGN_TLOG_UPLOAD is OFF)")
  endif()
  # cosign v3 turns --use-signing-config on by default, and that mode now
  # rejects --tlog-upload=false. To keep PR builds out of the public Rekor
  # transparency log we hand cosign a signing-config file built from the
  # default Sigstore services with the Rekor entries stripped. TSA is also
  # stripped to match the prior --tlog-upload=false behavior, which never
  # enabled signed timestamps. Generating via cosign itself keeps the schema in
  # sync with whichever cosign version cosign-installer ships.
  message(STATUS "cosign signing-config create -> ${SIGNING_CONFIG_FILE}")
  execute_process(
    COMMAND ${COSIGN_EXECUTABLE} signing-config create
      --with-default-services
      --no-default-rekor
      --no-default-tsa
      --out "${SIGNING_CONFIG_FILE}"
    RESULT_VARIABLE SIGNING_CONFIG_RES
  )
  if(NOT SIGNING_CONFIG_RES EQUAL 0)
    message(FATAL_ERROR "cosign signing-config create failed (exit ${SIGNING_CONFIG_RES})")
  endif()
  list(APPEND COSIGN_ARGS --signing-config "${SIGNING_CONFIG_FILE}")
endif()

# Registries arrive as a comma-separated string (the wrapper joined the CMake
# list with , to survive add_custom_command argument splitting).
string(REPLACE "," ";" REGISTRY_LIST "${REGISTRIES}")

foreach(REGISTRY IN LISTS REGISTRY_LIST)
  message(STATUS "cosign sign ${REGISTRY}@${DIGEST}")
  execute_process(
    COMMAND ${COSIGN_EXECUTABLE} ${COSIGN_ARGS} "${REGISTRY}@${DIGEST}"
    RESULT_VARIABLE RES
  )
  if(NOT RES EQUAL 0)
    message(FATAL_ERROR "cosign sign failed for ${REGISTRY}@${DIGEST}")
  endif()
endforeach()
