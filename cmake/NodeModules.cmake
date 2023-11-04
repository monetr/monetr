set(NODE_BIN ${CMAKE_BINARY_DIR}/node/bin)
set(NODE_MIN_VERSION "16.0.0")
set(NPM_MIN_VERSION "8.0.0")

find_package(Node REQUIRED)
find_package(Npm REQUIRED)
find_package(Pnpm REQUIRED)

set(NODE_MODULES ${CMAKE_SOURCE_DIR}/node_modules)
set(NODE_MODULES_BIN ${NODE_MODULES}/.bin)
set(NODE_MODULES_MARKER ${CMAKE_BINARY_DIR}/node-modules-marker.txt)
set(JEST_EXECUTABLE ${NODE_MODULES_BIN}/jest)
set(RSPACK_EXECUTABLE ${NODE_MODULES_BIN}/rspack)
set(REACT_EMAIL_EXECUTABLE ${NODE_MODULES_BIN}/email)
set(NEXT_EXECUTABLE ${NODE_MODULES_BIN}/next)
set(SITEMAP_EXECUTABLE ${NODE_MODULES_BIN}/next-sitemap)
set(STORYBOOK_EXECUTABLE ${NODE_MODULES_BIN}/storybook)

add_custom_command(
  OUTPUT ${NODE_MODULES}
         ${NODE_MODULES_MARKER}
         ${JEST_EXECUTABLE}
         ${RSPACK_EXECUTABLE}
         ${REACT_EMAIL_EXECUTABLE}
         ${NEXT_EXECUTABLE}
         ${SITEMAP_EXECUTABLE}
         ${CMAKE_SOURCE_DIR}/docs/node_modules
         ${CMAKE_SOURCE_DIR}/emails/node_modules
         ${CMAKE_SOURCE_DIR}/interface/node_modules
         ${CMAKE_SOURCE_DIR}/stories/node_modules
  BYPRODUCTS ${NODE_MODULES}
             ${NODE_MODULES_MARKER}
             ${JEST_EXECUTABLE}
             ${RSPACK_EXECUTABLE}
             ${REACT_EMAIL_EXECUTABLE}
             ${NEXT_EXECUTABLE}
             ${SITEMAP_EXECUTABLE}
             ${CMAKE_SOURCE_DIR}/docs/node_modules
             ${CMAKE_SOURCE_DIR}/emails/node_modules
             ${CMAKE_SOURCE_DIR}/interface/node_modules
             ${CMAKE_SOURCE_DIR}/stories/node_modules
  COMMAND ${PNPM_EXECUTABLE} install
  # By having a marker we make sure that if we cancel the install but the node_modules dir was created we still end up
  # doing install again if we didn't finish the first time.
  COMMAND ${CMAKE_COMMAND} -E touch ${NODE_MODULES_MARKER}
  COMMENT "Installing node/ui dependencies"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  DEPENDS
    ${CMAKE_SOURCE_DIR}/docs/package.json
    ${CMAKE_SOURCE_DIR}/interface/package.json
    ${CMAKE_SOURCE_DIR}/emails/package.json
    ${CMAKE_SOURCE_DIR}/package.json
    ${CMAKE_SOURCE_DIR}/pnpm-lock.yaml
    tools.pnpm
)

add_custom_target(
  dependencies.node_modules
  DEPENDS ${NODE_MODULES}
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
  tools.storybook
  DEPENDS dependencies.node_modules
)

