include(GolangTestUtils)

provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

include(ExternalProject)
ExternalProject_Add(cldr-json
  GIT_REPOSITORY https://github.com/unicode-org/cldr-json.git
  GIT_TAG 46.1.0
  SOURCE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/sources/cldr-json"
  BINARY_DIR ""
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  INSTALL_COMMAND ""
  TEST_COMMAND ""
)

add_custom_target(
  download.cldr-json
  # By having the byproducts here, cmake will automatically clean up this directory when the clean target is run.
  BYPRODUCTS ${CMAKE_CURRENT_SOURCE_DIR}/sources/cldr-json
  # Essentially a no-op
  COMMAND ${CMAKE_COMMAND} -E true
  DEPENDS cldr-json
)
