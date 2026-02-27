message(STATUS "Looking for all of the Go test coverage files in ${CMAKE_SOURCE_DIR}/build")

set(outputCoverageFile "${CMAKE_SOURCE_DIR}/build/tests/go-coverage.txt")
set(outputHtmlFile "${CMAKE_SOURCE_DIR}/build/tests/go-coverage.html")
file(REMOVE ${outputCoverageFile} ${outputHtmlFile})
file(GLOB_RECURSE coverageFiles "${CMAKE_SOURCE_DIR}/build/tests/*.txt")
execute_process(
  COMMAND go run ${CMAKE_SOURCE_DIR}/cmake/scripts/merge_go_coverage.go ${outputCoverageFile} ${coverageFiles}
  COMMAND go tool cover -html=${outputCoverageFile} -o ${outputHtmlFile}
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)
file(REMOVE ${coverageFiles})
