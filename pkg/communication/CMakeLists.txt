include(GolangUtils)
include(GolangTestUtils)

# Generate the mock interface for user communication.
add_custom_target(
  generate.mock.EmailCommunication ALL
  COMMAND ${MOCKGEN_EXECUTABLE} -source=${CMAKE_CURRENT_SOURCE_DIR}/email.go -package=mockgen -destination=${CMAKE_SOURCE_DIR}/pkg/internal/mockgen/email.go EmailCommunication
  DEPENDS
    ${CMAKE_CURRENT_SOURCE_DIR}/email.go
    ${MOCKGEN_EXECUTABLE}
)

set(TEST_GO_TAGS "local")
provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

# if (BUILD_TESTING)
#   add_test(
#     NAME generate/mock/${PACKAGE}/UserCommunication
#     COMMAND ${CMAKE_COMMAND} --build ${CMAKE_BINARY_DIR} -t generate.mock.UserCommunication
#   )
#   set_tests_properties(generate/mock/${PACKAGE}/UserCommunication
#     PROPERTIES
#     FIXTURES_SETUP "go.mocks"
#   )
# endif()
