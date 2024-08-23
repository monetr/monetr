include(ExternalProject)
ExternalProject_Add(legal
  GIT_REPOSITORY https://github.com/monetr/legal.git
  GIT_TAG ca08f9e565a9153e01db3cbb237564bbd807aee7
  SOURCE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/data/legal"
  BINARY_DIR ""
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  INSTALL_COMMAND ""
  TEST_COMMAND ""
)

add_custom_target(
  download.legal
  # By having the byproducts here, cmake will automatically clean up this directory when the clean target is run.
  BYPRODUCTS ${CMAKE_CURRENT_SOURCE_DIR}/data/legal
  # Essentially a no-op
  COMMAND ${CMAKE_COMMAND} -E true
  DEPENDS legal
)
