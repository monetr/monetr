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

set(STORYBOOK_DOCS "${DOCUMENTATION_OUT_DIR}/_storybook")
set(STORY_OUTPUT ${CMAKE_BINARY_DIR}/stories/dist)
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
  COMMAND ${CMAKE_COMMAND} -E make_directory ${STORYBOOK_DOCS}
  COMMAND ${CMAKE_COMMAND} -E copy_directory ${STORY_OUTPUT}/ ${STORYBOOK_DOCS}
  COMMAND ${CMAKE_COMMAND} -E copy ${STORY_OUTPUT}/mockServiceWorker.js ${CMAKE_SOURCE_DIR}/docs/out/mockServiceWorker.js
  COMMENT "Building monetr.app documentation and website"
  WORKING_DIRECTORY ${DOCS_SRC_DIR}
  DEPENDS
    dependencies.node_modules
    tools.next
    ${ALL_DOCS_FILES}
    ${PUBLIC_FILES}
    ${DOCS_CONFIG_FILES}
    build.stories
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
