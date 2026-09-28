include(GolangTestUtils)

set(TEST_GO_TAGS "noui")
provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

# monetr only needs a small part of the CLDR dataset for currency data. So
# rather than cloning the entire (very large) cldr-json repository, do a
# shallow sparse clone that only pulls the files we need for the pinned version.
# The per-locale currency names are then condensed into a single file by
# cldrgen.
set(CLDR_VERSION "48.2.2")
set(CLDR_SOURCES_DIR "${CMAKE_CURRENT_SOURCE_DIR}/sources")
set(CLDR_PREFIX_DIR "${CMAKE_CURRENT_BINARY_DIR}/cldr-json")
# The download step runs from within the source dir, so the clone goes into a
# folder beneath it that can be safely removed and recreated.
set(CLDR_CHECKOUT_DIR "${CLDR_PREFIX_DIR}/checkout/repo")

find_package(Git REQUIRED)
include(ExternalProject)
ExternalProject_Add(cldr-json
  PREFIX "${CLDR_PREFIX_DIR}"
  SOURCE_DIR "${CLDR_PREFIX_DIR}/checkout"
  DOWNLOAD_COMMAND
    ${CMAKE_COMMAND} -E rm -rf ${CLDR_CHECKOUT_DIR}
    COMMAND ${GIT_EXECUTABLE} -c advice.detachedHead=false clone --quiet --depth 1 --branch ${CLDR_VERSION} --filter=blob:none --sparse https://github.com/unicode-org/cldr-json.git ${CLDR_CHECKOUT_DIR}
    # currencies.json only exists in the per-locale cldr-numbers-full folders.
    COMMAND ${GIT_EXECUTABLE} -C ${CLDR_CHECKOUT_DIR} sparse-checkout set --no-cone
      /cldr-json/cldr-core/supplemental/currencyData.json
      /cldr-json/cldr-core/supplemental/likelySubtags.json
      currencies.json
  UPDATE_COMMAND ""
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  INSTALL_COMMAND ""
  TEST_COMMAND ""
)

ExternalProject_Get_Property(cldr-json STAMP_DIR)
set(CLDR_STAMP_DIR ${STAMP_DIR})

set(CLDR_OUTPUTS
  ${CLDR_SOURCES_DIR}/currencyData.json
  ${CLDR_SOURCES_DIR}/likelySubtags.json
  ${CLDR_SOURCES_DIR}/currencyNames.json
)
add_custom_command(
  OUTPUT ${CLDR_OUTPUTS}
  COMMAND ${CMAKE_COMMAND} -E copy
    ${CLDR_CHECKOUT_DIR}/cldr-json/cldr-core/supplemental/currencyData.json
    ${CLDR_CHECKOUT_DIR}/cldr-json/cldr-core/supplemental/likelySubtags.json
    ${CLDR_SOURCES_DIR}
  COMMAND ${CMAKE_Go_COMPILER} run ./internal/cldrgen -cldr ${CLDR_CHECKOUT_DIR} -output ${CLDR_SOURCES_DIR}/currencyNames.json
  COMMENT "Generating CLDR ${CLDR_VERSION} currency data"
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  DEPENDS
    cldr-json
    # The done stamp is touched whenever the checkout changes, like when the
    # version is bumped, which will cause this to be regenerated.
    ${CLDR_STAMP_DIR}/cldr-json-done
    ${CMAKE_CURRENT_SOURCE_DIR}/internal/cldrgen/main.go
  VERBATIM
)

add_custom_target(
  download.cldr-json
  DEPENDS ${CLDR_OUTPUTS}
)
