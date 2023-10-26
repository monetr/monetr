# Define the directories
set(UI_SRC_DIR "${CMAKE_CURRENT_SOURCE_DIR}")
set(PUBLIC_DIR "${UI_SRC_DIR}/public")

# Get all UI files recursively
file(GLOB_RECURSE ALL_UI_FILES "${UI_SRC_DIR}/*")

# Get all test UI files (those matching the pattern '*.spec.*')
file(GLOB_RECURSE TEST_UI_FILES "${UI_SRC_DIR}/*.spec.*")

# Get the list of APP_UI_FILES by removing TEST_UI_FILES from ALL_UI_FILES
list(FILTER ALL_UI_FILES EXCLUDE REGEX ".*\\.spec\\..*")
set(APP_UI_FILES ${ALL_UI_FILES})

# Get all files in the PUBLIC_DIR (non-recursive)
file(GLOB PUBLIC_FILES "${PUBLIC_DIR}/*")

# Get the list of UI_CONFIG_FILES
file(GLOB UI_CONFIG_FILES
  "${UI_SRC_DIR}/tsconfig.json"
  "${UI_SRC_DIR}/*.config.js"
  "${UI_SRC_DIR}/*.config.cjs"
)

set(UI_DIST
  ${CMAKE_SOURCE_DIR}/server/ui/static/assets
  ${CMAKE_SOURCE_DIR}/server/ui/static/index.html
  ${CMAKE_SOURCE_DIR}/server/ui/static/public
  ${CMAKE_SOURCE_DIR}/server/ui/static/logo192.png
  ${CMAKE_SOURCE_DIR}/server/ui/static/logo512.png
  ${CMAKE_SOURCE_DIR}/server/ui/static/manifest.json
  ${CMAKE_SOURCE_DIR}/server/ui/static/robots.txt
  CACHE INTERNAL ""
)
add_custom_command(
  OUTPUT ${UI_DIST}
  BYPRODUCTS ${UI_DIST}
  COMMAND ${GIT_EXECUTABLE} clean -f -X ${CMAKE_SOURCE_DIR}/server/ui/static
  COMMAND ${RSPACK_EXECUTABLE} build --mode production
  COMMENT "Building monetr's user interface"
  WORKING_DIRECTORY ${UI_SRC_DIR}
  DEPENDS
    ${UI_SRC_DIR}/package.json
    ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
    dependencies.node_modules
    tools.rspack
    ${APP_UI_FILES}
    ${PUBLIC_FILES}
    ${UI_CONFIG_FILES}
)

add_custom_target(
  build.interface ALL
  DEPENDS ${UI_DIST}
)

file(GLOB_RECURSE SUB_PACKAGES RELATIVE ${CMAKE_CURRENT_SOURCE_DIR} "${CMAKE_CURRENT_SOURCE_DIR}/src/*CMakeLists.txt")
# # Loop through the files and get their directories
foreach(SUB_PACKAGE IN LISTS SUB_PACKAGES)
  get_filename_component(DIR ${SUB_PACKAGE} DIRECTORY)
  add_subdirectory(${DIR})
endforeach()
