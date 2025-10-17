set(NODE_BIN ${CMAKE_BINARY_DIR}/node/bin)
set(NODE_MIN_VERSION "20.0.0")
set(NPM_MIN_VERSION "9.0.0")

find_package(Node REQUIRED)
find_program(NPM_EXECUTABLE NAMES npm)
find_package(Pnpm REQUIRED)

set(NODE_MODULES ${CMAKE_SOURCE_DIR}/node_modules)
set(NODE_MODULES_BIN ${NODE_MODULES}/.bin)
set(NODE_MODULES_MARKER ${CMAKE_BINARY_DIR}/node-modules-marker.txt)
if(WIN32)
  set(JS_EXECUTABLE_SUFFIX ".CMD")
else()
  set(JS_EXECUTABLE_SUFFIX "")
endif()
set(HYPERLINK_EXECUTABLE ${NODE_MODULES_BIN}/hyperlink${JS_EXECUTABLE_SUFFIX})
set(JEST_EXECUTABLE ${NODE_MODULES_BIN}/jest${JS_EXECUTABLE_SUFFIX})
set(NEXT_EXECUTABLE ${NODE_MODULES_BIN}/next${JS_EXECUTABLE_SUFFIX})
set(REACT_EMAIL_EXECUTABLE ${NODE_MODULES_BIN}/email${JS_EXECUTABLE_SUFFIX})
set(RSBUILD_EXECUTABLE ${NODE_MODULES_BIN}/rsbuild${JS_EXECUTABLE_SUFFIX})
set(RSPACK_EXECUTABLE ${NODE_MODULES_BIN}/rspack${JS_EXECUTABLE_SUFFIX})
set(SITEMAP_EXECUTABLE ${NODE_MODULES_BIN}/next-sitemap${JS_EXECUTABLE_SUFFIX})
set(SPELLCHECKER_EXECUTABLE ${NODE_MODULES_BIN}/spellchecker${JS_EXECUTABLE_SUFFIX})
set(STORYBOOK_EXECUTABLE ${NODE_MODULES_BIN}/storybook${JS_EXECUTABLE_SUFFIX})

set(PNPM_ARGUMENTS "--frozen-lockfile" "--ignore-scripts")

add_custom_command(
  OUTPUT ${NODE_MODULES}
         ${NODE_MODULES_MARKER}
         ${HYPERLINK_EXECUTABLE}
         ${JEST_EXECUTABLE}
         ${NEXT_EXECUTABLE}
         ${REACT_EMAIL_EXECUTABLE}
         ${RSBUILD_EXECUTABLE}
         ${RSPACK_EXECUTABLE}
         ${SITEMAP_EXECUTABLE}
         ${SPELLCHECKER_EXECUTABLE}
         ${STORYBOOK_EXECUTABLE}
         ${CMAKE_SOURCE_DIR}/docs/node_modules
         ${CMAKE_SOURCE_DIR}/emails/node_modules
         ${CMAKE_SOURCE_DIR}/interface/node_modules
         ${CMAKE_SOURCE_DIR}/stories/node_modules
  BYPRODUCTS ${NODE_MODULES}
             ${NODE_MODULES_MARKER}
             ${HYPERLINK_EXECUTABLE}
             ${JEST_EXECUTABLE}
             ${NEXT_EXECUTABLE}
             ${REACT_EMAIL_EXECUTABLE}
             ${RSBUILD_EXECUTABLE}
             ${RSPACK_EXECUTABLE}
             ${SITEMAP_EXECUTABLE}
             ${SPELLCHECKER_EXECUTABLE}
             ${STORYBOOK_EXECUTABLE}
             ${CMAKE_SOURCE_DIR}/docs/node_modules
             ${CMAKE_SOURCE_DIR}/emails/node_modules
             ${CMAKE_SOURCE_DIR}/interface/node_modules
             ${CMAKE_SOURCE_DIR}/stories/node_modules
  # Run the actual pnpm install with our args
  COMMAND ${PNPM_EXECUTABLE} install ${PNPM_ARGUMENTS}
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
  tools.storybook
  DEPENDS dependencies.node_modules
)

add_custom_target(
  tools.spellchecker
  DEPENDS dependencies.node_modules
)
