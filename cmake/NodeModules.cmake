set(NODE_BIN ${CMAKE_BINARY_DIR}/node/bin)
set(NODE_MIN_VERSION "20.12.0")
set(NPM_MIN_VERSION "9.0.0")

find_package(Node REQUIRED)
find_program(NPM_EXECUTABLE NAMES npm)
find_package(Pnpm REQUIRED)

set(NODE_MODULES ${CMAKE_SOURCE_DIR}/node_modules)
set(NODE_MODULES_MARKER ${CMAKE_BINARY_DIR}/node-modules-marker.txt)
# Global Commands
set(BIOME_EXECUTABLE ${PNPM_EXECUTABLE} biome)

# Documentation Commands
set(HYPERLINK_EXECUTABLE ${PNPM_EXECUTABLE} hyperlink)
set(SITEMAP_EXECUTABLE ${PNPM_EXECUTABLE} next-sitemap)
set(SPELLCHECKER_EXECUTABLE ${PNPM_EXECUTABLE} spellchecker)
set(NEXT_EXECUTABLE ${PNPM_EXECUTABLE} next)

# New Documentation Commands
set(RSPRESS_EXECUTABLE ${PNPM_EXECUTABLE} rspress)

# Email Template Commands
set(REACT_EMAIL_EXECUTABLE ${PNPM_EXECUTABLE} email)

# Frontend Commands
set(JEST_EXECUTABLE ${PNPM_EXECUTABLE} jest)
set(RSBUILD_EXECUTABLE ${PNPM_EXECUTABLE} rsbuild)

set(PNPM_ARGUMENTS "--frozen-lockfile" "--ignore-scripts")

add_custom_command(
  OUTPUT ${NODE_MODULES}
         ${NODE_MODULES_MARKER}
         ${CMAKE_SOURCE_DIR}/docs/node_modules
         ${CMAKE_SOURCE_DIR}/emails/node_modules
         ${CMAKE_SOURCE_DIR}/interface/node_modules
         ${CMAKE_SOURCE_DIR}/stories/node_modules
         ${CMAKE_SOURCE_DIR}/site/node_modules
  BYPRODUCTS ${NODE_MODULES}
             ${NODE_MODULES_MARKER}
             ${CMAKE_SOURCE_DIR}/docs/node_modules
             ${CMAKE_SOURCE_DIR}/emails/node_modules
             ${CMAKE_SOURCE_DIR}/interface/node_modules
             ${CMAKE_SOURCE_DIR}/stories/node_modules
             ${CMAKE_SOURCE_DIR}/site/node_modules
  # Run the actual pnpm install with our args
  COMMAND ${PNPM_EXECUTABLE} install ${PNPM_ARGUMENTS}
  # By having a marker we make sure that if we cancel the install but the node_modules dir was created we still end up
  # doing install again if we didn't finish the first time.
  COMMAND ${CMAKE_COMMAND} -E touch ${NODE_MODULES_MARKER}
  COMMENT "Installing node/ui dependencies"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  DEPENDS
    ${CMAKE_SOURCE_DIR}/docs/package.json
    ${CMAKE_SOURCE_DIR}/emails/package.json
    ${CMAKE_SOURCE_DIR}/interface/package.json
    ${CMAKE_SOURCE_DIR}/site/package.json
    ${CMAKE_SOURCE_DIR}/package.json
    ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
    tools.pnpm
)

add_custom_target(
  dependencies.node_modules
  DEPENDS ${NODE_MODULES}
)

add_custom_target(
  tools.biome
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.rsbuild
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.rspack
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.jest
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.react-email
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.next
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.next-sitemap
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.hyperlink
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.spellchecker
  DEPENDS dependencies.node_modules
)
