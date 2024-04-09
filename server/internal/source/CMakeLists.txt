

set(GO_SRC_DIR "${CMAKE_SOURCE_DIR}/server")
file(GLOB_RECURSE APP_GO_FILES
  "${GO_SRC_DIR}/*.go"
)
list(FILTER APP_GO_FILES EXCLUDE REGEX ".+_test\\.go")
list(FILTER APP_GO_FILES EXCLUDE REGEX "source/embed")

set(SOURCEMAP_MARKER ${CMAKE_BINARY_DIR}/golang-sourcemap.txt)
add_custom_command(
  OUTPUT ${SOURCEMAP_MARKER}
  COMMENT "Finished generating source maps for sentry from golang"
  COMMAND ${CMAKE_COMMAND} -E touch ${SOURCEMAP_MARKER}
)

add_custom_target(
  sourcemaps.golang
  DEPENDS ${SOURCEMAP_MARKER}
)

foreach(FILE ${APP_GO_FILES})
  string(REPLACE "${CMAKE_SOURCE_DIR}/" "${CMAKE_CURRENT_SOURCE_DIR}/embed/" OUTPUT_GO_FILE "${FILE}")
  string(REPLACE "${CMAKE_SOURCE_DIR}/" "" FILE_FRIENDLY "${FILE}")
  get_filename_component(OUTPUT_GO_DIRECTORY ${OUTPUT_GO_FILE} DIRECTORY)
  add_custom_command(
    OUTPUT ${OUTPUT_GO_FILE}.txt
    BYPRODUCTS ${OUTPUT_GO_FILE}.txt
    COMMAND ${CMAKE_COMMAND} -E make_directory ${OUTPUT_GO_DIRECTORY}
    COMMAND ${CMAKE_COMMAND} -E copy ${FILE} ${OUTPUT_GO_FILE}.txt
    COMMENT "Generating source map for ${FILE_FRIENDLY}"
    DEPENDS ${FILE}
  )
  add_custom_command(OUTPUT ${SOURCEMAP_MARKER} APPEND DEPENDS ${OUTPUT_GO_FILE}.txt)
endforeach()

