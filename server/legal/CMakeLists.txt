set(LEGAL_MODULE ${CMAKE_CURRENT_SOURCE_DIR}/data/legal/README.md)
add_custom_command(
  OUTPUT ${LEGAL_MODULE}
  COMMAND ${GIT_EXECUTABLE} submodule update --init server/legal/data/legal
  COMMENT "Retrieving legal documents"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)

add_custom_target(
  download.legal
  DEPENDS ${LEGAL_MODULE}
)
