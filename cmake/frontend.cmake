set(NODE_BIN ${CMAKE_BINARY_DIR}/node/bin)
set(NODE_MIN_VERSION "16.0.0")
set(NPM_MIN_VERSION "8.0.0")

find_package(Node REQUIRED)
find_package(Npm REQUIRED)
find_package(Pnpm REQUIRED)

# Define the directories
set(PUBLIC_DIR "${CMAKE_SOURCE_DIR}/public")
set(UI_SRC_DIR "${CMAKE_SOURCE_DIR}/ui")

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
  "${CMAKE_SOURCE_DIR}/tsconfig.json"
  "${CMAKE_SOURCE_DIR}/*.config.js"
  "${CMAKE_SOURCE_DIR}/*.config.cjs"
)

set(NODE_MODULES ${CMAKE_SOURCE_DIR}/node_modules)
set(NODE_MODULES_MARKER ${CMAKE_BINARY_DIR}/node-modules-marker.txt)
set(JEST_EXECUTABLE ${NODE_MODULES}/.bin/jest)
set(RSPACK_EXECUTABLE ${NODE_MODULES}/.bin/rspack)
add_custom_command(
  OUTPUT ${NODE_MODULES} ${NODE_MODULES_MARKER} ${JEST_EXECUTABLE} ${RSPACK_EXECUTABLE}
  BYPRODUCTS ${NODE_MODULES} ${NODE_MODULES_MARKER} ${JEST_EXECUTABLE} ${RSPACK_EXECUTABLE}
  COMMAND ${PNPM_EXECUTABLE} install
  # By having a marker we make sure that if we cancel the install but the node_modules dir was created we still end up
  # doing install again if we didn't finish the first time.
  COMMAND ${CMAKE_COMMAND} -E touch ${NODE_MODULES_MARKER}
  COMMENT "Installing node/ui dependencies"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  DEPENDS
    ${CMAKE_SOURCE_DIR}/package.json
    ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
    ${PNPM_EXECUTABLE}
)

set(UI_DIST
  ${CMAKE_SOURCE_DIR}/pkg/ui/static/assets
  ${CMAKE_SOURCE_DIR}/pkg/ui/static/index.html
  ${CMAKE_SOURCE_DIR}/pkg/ui/static/public
  ${CMAKE_SOURCE_DIR}/pkg/ui/static/logo192.png
  ${CMAKE_SOURCE_DIR}/pkg/ui/static/logo512.png
  ${CMAKE_SOURCE_DIR}/pkg/ui/static/manifest.json
  ${CMAKE_SOURCE_DIR}/pkg/ui/static/robots.txt
)
add_custom_command(
  OUTPUT ${UI_DIST}
  BYPRODUCTS ${UI_DIST}
  COMMAND ${GIT_EXECUTABLE} clean -f -X ${CMAKE_SOURCE_DIR}/pkg/ui/static
  COMMAND ${RSPACK_EXECUTABLE} build --mode production
  COMMENT "Building monetr's user interface"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  DEPENDS
    ${CMAKE_SOURCE_DIR}/package.json
    ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
    ${NODE_MODULES}
    ${RSPACK_EXECUTABLE}
    ${APP_UI_FILES}
    ${PUBLIC_FILES}
    ${UI_CONFIG_FILES}
)

add_custom_target(
  node_modules
  DEPENDS ${NODE_MODULES}
)
