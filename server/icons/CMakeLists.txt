include(GolangTestUtils)

provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

include(ExternalProject)
ExternalProject_Add(simple-icons
  GIT_REPOSITORY https://github.com/simple-icons/simple-icons.git
  GIT_TAG 14.0.0
  SOURCE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/sources/simple-icons"
  BINARY_DIR ""
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  INSTALL_COMMAND ""
  TEST_COMMAND ""
)

add_custom_target(
  download.simple-icons
  # By having the byproducts here, cmake will automatically clean up this directory when the clean target is run.
  BYPRODUCTS ${CMAKE_CURRENT_SOURCE_DIR}/sources/simple-icons
  # Essentially a no-op
  COMMAND ${CMAKE_COMMAND} -E true
  DEPENDS simple-icons
)
