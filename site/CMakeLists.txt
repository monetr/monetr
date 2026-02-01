add_custom_target(
  development.site.up
  COMMAND ${CMAKE_COMMAND} -E env GIT_BRANCH=${GIT_BRANCH} ${RSPRESS_EXECUTABLE} dev
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  DEPENDS
    dependencies.node_modules
)
