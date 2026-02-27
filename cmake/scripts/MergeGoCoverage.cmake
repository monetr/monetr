message(STATUS "Looking for all of the Go test coverage files in ${CMAKE_SOURCE_DIR}/build")

set(outputFile "${CMAKE_SOURCE_DIR}/build/tests/go-coverage.txt")
file(REMOVE ${outputFile})
file(GLOB_RECURSE coverageFiles "${CMAKE_SOURCE_DIR}/build/tests/*.txt")
execute_process(
  COMMAND go run ${CMAKE_SOURCE_DIR}/cmake/scripts/merge_go_coverage.go ${outputFile} ${coverageFiles}
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)
file(REMOVE ${coverageFiles})
