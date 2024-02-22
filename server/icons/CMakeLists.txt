include(GolangTestUtils)

provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

set(SIMPLE_ICONS ${CMAKE_CURRENT_SOURCE_DIR}/sources/simple-icons/icons)
add_custom_command(
  OUTPUT ${SIMPLE_ICONS}
  COMMAND ${GIT_EXECUTABLE} submodule update --init server/icons/sources/simple-icons
  COMMENT "Retrieving simple-icons"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)

add_custom_target(
  download.simple-icons
  DEPENDS ${SIMPLE_ICONS}
)
