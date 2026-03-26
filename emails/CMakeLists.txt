set(EMAIL_TEMPLATE_INPUT_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/src/emails")
set(EMAIL_IN_TREE_OUTPUT "${CMAKE_SOURCE_DIR}/server/communication/email_templates")

# Collect all template input files for dependency tracking
file(GLOB_RECURSE EMAIL_TEMPLATE_INPUTS "${EMAIL_TEMPLATE_INPUT_DIRECTORY}/*/index.tsx")
file(GLOB_RECURSE EMAIL_COMPONENT_INPUTS "${CMAKE_CURRENT_SOURCE_DIR}/src/components/*.tsx")
file(GLOB_RECURSE EMAIL_STYLE_INPUTS "${CMAKE_CURRENT_SOURCE_DIR}/src/**/*.scss")

# Determine expected output files from template names
set(TEMPLATE_OUTPUTS)
foreach(FILE ${EMAIL_TEMPLATE_INPUTS})
  string(REPLACE "${EMAIL_TEMPLATE_INPUT_DIRECTORY}/" "" TEMPLATE_NAME "${FILE}")
  string(REPLACE "/index.tsx" "" TEMPLATE_NAME "${TEMPLATE_NAME}")
  list(APPEND TEMPLATE_OUTPUTS "${EMAIL_IN_TREE_OUTPUT}/${TEMPLATE_NAME}.html")
  list(APPEND TEMPLATE_OUTPUTS "${EMAIL_IN_TREE_OUTPUT}/${TEMPLATE_NAME}.txt")
endforeach()

add_custom_command(
  OUTPUT ${TEMPLATE_OUTPUTS}
  BYPRODUCTS ${TEMPLATE_OUTPUTS}
  COMMAND ${CMAKE_COMMAND} -E env EMAIL_OUT_DIR=${EMAIL_IN_TREE_OUTPUT} ${RSBUILD_EXECUTABLE} build
  COMMENT "Building email templates"
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  DEPENDS
    tools.rsbuild
    ${EMAIL_TEMPLATE_INPUTS}
    ${EMAIL_COMPONENT_INPUTS}
    ${EMAIL_STYLE_INPUTS}
    ${CMAKE_CURRENT_SOURCE_DIR}/rsbuild.config.ts
    ${CMAKE_CURRENT_SOURCE_DIR}/src/build/entry.ts
    ${CMAKE_CURRENT_SOURCE_DIR}/src/build/rsbuildPluginEmail.ts
)

add_custom_target(
  build.email
  DEPENDS ${TEMPLATE_OUTPUTS}
)

add_custom_target(
  development.email
  COMMENT "Running the local development tools for email templates"
  COMMAND ${RSBUILD_EXECUTABLE} dev
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  DEPENDS
    tools.rsbuild
)
