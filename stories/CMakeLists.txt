set(UI_SRC_DIR "${CMAKE_SOURCE_DIR}/interface" CACHE INTERNAL "")
set(PUBLIC_DIR "${UI_SRC_DIR}/public")

# Get all UI files recursively
file(GLOB_RECURSE ALL_UI_FILES "${UI_SRC_DIR}/src/*")

# Get all test UI files (those matching the pattern '*.spec.*')
file(GLOB_RECURSE TEST_UI_FILES "${UI_SRC_DIR}/src/*.spec.*")

# Get the list of APP_UI_FILES by removing TEST_UI_FILES from ALL_UI_FILES
list(FILTER ALL_UI_FILES EXCLUDE REGEX ".*\\.spec\\..*")

# Get all files in the PUBLIC_DIR (non-recursive)
file(GLOB PUBLIC_FILES "${PUBLIC_DIR}/*")

file(GLOB STORY_FILES "${CMAKE_CURRENT_SOURCE_DIR}/*.*")

set(STORY_OUTPUT ${CMAKE_BINARY_DIR}/stories/dist)
add_custom_command(
  OUTPUT ${STORY_OUTPUT}
  BYPRODUCTS ${STORY_OUTPUT}
  COMMENT "Building storybook"
  COMMAND ${CMAKE_COMMAND} -E remove_directory "${STORY_OUTPUT}"
  COMMAND ${CMAKE_COMMAND} -E env NODE_ENV=production ${STORYBOOK_EXECUTABLE} build --disable-telemetry -c ${CMAKE_CURRENT_SOURCE_DIR} -o ${STORY_OUTPUT}
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  DEPENDS
    dependencies.node_modules
    tools.storybook
    ${ALL_UI_FILES}
    ${PUBLIC_FILES}
    ${STORY_FILES}
)

add_custom_target(
  build.stories
  COMMAND true
  DEPENDS ${STORY_OUTPUT}
)

add_custom_target(
  development.stories.up
  COMMAND ${STORYBOOK_EXECUTABLE} dev -p 6006 --disable-telemetry -c ${CMAKE_CURRENT_SOURCE_DIR}
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  DEPENDS
    dependencies.node_modules
    tools.storybook
    ${ALL_UI_FILES}
    ${PUBLIC_FILES}
    ${STORY_FILES}
)
