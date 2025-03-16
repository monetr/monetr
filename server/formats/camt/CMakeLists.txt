include(GolangUtils)
include(GolangTestUtils)

# Generate the mock interface for user communication.
add_custom_target(
  generate.mock.Camt053 ALL
  COMMAND ${CMAKE_Go_COMPILER} generate ${CMAKE_CURRENT_SOURCE_DIR}/doc.go
  DEPENDS ${CMAKE_CURRENT_SOURCE_DIR}/doc.go
)

provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})
