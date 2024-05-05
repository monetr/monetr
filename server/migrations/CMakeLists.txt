include(GolangTestUtils)

provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

add_custom_target(
  development.migrate
  COMMAND ${CMAKE_Go_COMPILER} run ${MONETR_CLI_PKG} database migrate
  COMMENT "Performing database migrations"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)
