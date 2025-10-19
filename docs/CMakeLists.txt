# Define the directories
set(DOCS_SRC_DIR "${CMAKE_CURRENT_SOURCE_DIR}")
set(PUBLIC_DIR "${DOCS_SRC_DIR}/public")

# Get all DOCS files recursively
file(GLOB_RECURSE ALL_DOCS_FILES "${DOCS_SRC_DIR}/src/*")

# Get all files in the PUBLIC_DIR (non-recursive)
file(GLOB PUBLIC_FILES "${PUBLIC_DIR}/*")

# Get the list of DOCS_CONFIG_FILES
file(GLOB DOCS_CONFIG_FILES
  "${DOCS_SRC_DIR}/*.json"
  "${DOCS_SRC_DIR}/*.config.*"
)

# NOTE: This cannot be outside this directory otherwise next complains.
set(DOCUMENTATION_OUT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/out")
set(DOCUMENTATION_DIST
  "${DOCUMENTATION_OUT_DIR}/404"
  "${DOCUMENTATION_OUT_DIR}/_next"
  "${DOCUMENTATION_OUT_DIR}/contact"
  "${DOCUMENTATION_OUT_DIR}/documentation"
  "${DOCUMENTATION_OUT_DIR}/favicon.ico"
  "${DOCUMENTATION_OUT_DIR}/logo192.png"
  "${DOCUMENTATION_OUT_DIR}/.nojekyll"
  "${DOCUMENTATION_OUT_DIR}/404.html"
  "${DOCUMENTATION_OUT_DIR}/index.html"
  "${DOCUMENTATION_OUT_DIR}/robots.txt"
  "${DOCUMENTATION_OUT_DIR}/sitemap-0.xml"
  "${DOCUMENTATION_OUT_DIR}/sitemap.xml"
)

execute_process(
  COMMAND ${GIT_EXECUTABLE} rev-parse --abbrev-ref HEAD
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  OUTPUT_VARIABLE GIT_BRANCH
  ERROR_VARIABLE NOOP
  OUTPUT_STRIP_TRAILING_WHITESPACE
)

add_custom_command(
  OUTPUT ${DOCUMENTATION_DIST} ${CMAKE_CURRENT_SOURCE_DIR}/.next ${CMAKE_CURRENT_SOURCE_DIR}/out
  BYPRODUCTS ${DOCUMENTATION_DIST} ${CMAKE_CURRENT_SOURCE_DIR}/.next ${CMAKE_CURRENT_SOURCE_DIR}/out
  # Since output files are hashed, make sure that there aren't any files that might be left over from the previous build
  # in the output directory.
  COMMAND ${GIT_EXECUTABLE} clean -f -X ${DOCUMENTATION_OUT_DIR}
  # Once the output directory is clean, then actually build the user interface.
  COMMAND ${CMAKE_COMMAND} -E env GIT_BRANCH=${GIT_BRANCH} ${NEXT_EXECUTABLE} build ${DOCS_SRC_DIR}
  COMMAND ${SITEMAP_EXECUTABLE} --config ${DOCS_SRC_DIR}/next-sitemap.config.js
  COMMAND ${NODE_EXECUTABLE} ${DOCS_SRC_DIR}/intl-temp.js ${CMAKE_CURRENT_SOURCE_DIR}/out
  # This fixes an issue with github pages not liking underscore prefixed directories.
  COMMAND ${CMAKE_COMMAND} -E touch ${DOCUMENTATION_OUT_DIR}/.nojekyll
  COMMENT "Building monetr.app documentation and website"
  WORKING_DIRECTORY ${DOCS_SRC_DIR}
  DEPENDS
    dependencies.node_modules
    tools.next
    ${ALL_DOCS_FILES}
    ${PUBLIC_FILES}
    ${DOCS_CONFIG_FILES}
)

add_custom_target(
  build.docs
  DEPENDS ${DOCUMENTATION_DIST}
  COMMENT "Documentation has been built at: ${DOCUMENTATION_OUT_DIR}"
)

add_custom_target(
  development.documentation.up
  COMMAND ${CMAKE_COMMAND} -E env GIT_BRANCH=${GIT_BRANCH} ${NEXT_EXECUTABLE} dev
  WORKING_DIRECTORY ${DOCS_SRC_DIR}
  DEPENDS
    dependencies.node_modules
    tools.next
    ${ALL_DOCS_FILES}
    ${PUBLIC_FILES}
    ${DOCS_CONFIG_FILES}
)

###############################################################################
# Documentation Tests                                                         #
###############################################################################

file(
  GLOB_RECURSE SUB_PACKAGES
  RELATIVE ${CMAKE_CURRENT_SOURCE_DIR} "${CMAKE_SOURCE_DIR}/docs/src/*CMakeLists.txt"
)
# # Loop through the files and get their directories
foreach(SUB_PACKAGE IN LISTS SUB_PACKAGES)
  get_filename_component(DIR ${SUB_PACKAGE} DIRECTORY)
  add_subdirectory(${DIR})
endforeach()

if(BUILD_TESTING)
  add_test(
    NAME docs/precompile
    COMMAND ${CMAKE_COMMAND} --build ${CMAKE_BINARY_DIR} -t build.docs
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  )
  set_tests_properties(docs/precompile PROPERTIES
    FIXTURES_REQUIRED node_modules
    FIXTURES_SETUP documentation_output
    TIMEOUT 240
  )
  set_property(
    TEST docs/precompile
    PROPERTY LABELS "docs"
  )

  set(HYPERLINK_GITHUB_ARGS "--check-anchors" "--sources" "${CMAKE_SOURCE_DIR}/docs/src/pages")
  if(NOT "$ENV{CI}" STREQUAL "")
    list(APPEND HYPERLINK_GITHUB_ARGS "--github-actions")
  endif()
  add_test(
    NAME docs/check-links
    COMMAND ${HYPERLINK_EXECUTABLE} ${CMAKE_SOURCE_DIR}/docs/out ${HYPERLINK_GITHUB_ARGS}
  )
  set_tests_properties(docs/check-links PROPERTIES
    FIXTURES_REQUIRED documentation_output
    TIMEOUT 45
  )
  set_property(
    TEST docs/check-links
    PROPERTY LABELS "docs"
  )
endif()
