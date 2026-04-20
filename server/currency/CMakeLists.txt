include(GolangTestUtils)

set(TEST_GO_TAGS "noui")
provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

set(CLDR_VERSION "46.1.0")

include(ExternalProject)
ExternalProject_Add(cldr-numbers-full
  URL https://registry.npmjs.org/cldr-numbers-full/-/cldr-numbers-full-${CLDR_VERSION}.tgz
  DOWNLOAD_NO_EXTRACT OFF
  DOWNLOAD_EXTRACT_TIMESTAMP OFF
  DOWNLOAD_NO_PROGRESS ON
  SOURCE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/sources/cldr-numbers-full"
  BINARY_DIR ""
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  INSTALL_COMMAND ""
  TEST_COMMAND ""
)

ExternalProject_Add(cldr-core
  URL https://registry.npmjs.org/cldr-core/-/cldr-core-${CLDR_VERSION}.tgz
  DOWNLOAD_NO_EXTRACT OFF
  DOWNLOAD_EXTRACT_TIMESTAMP OFF
  DOWNLOAD_NO_PROGRESS ON
  SOURCE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/sources/cldr-core"
  BINARY_DIR ""
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  INSTALL_COMMAND ""
  TEST_COMMAND ""
)

add_custom_target(
  download.cldr-json
  # By having the byproducts here, cmake will automatically clean up this directory when the clean target is run.
  BYPRODUCTS ${CMAKE_CURRENT_SOURCE_DIR}/sources/cldr-core ${CMAKE_CURRENT_SOURCE_DIR}/sources/cldr-numbers-full
  # Essentially a no-op
  COMMAND ${CMAKE_COMMAND} -E true
  DEPENDS cldr-core cldr-numbers-full
)
