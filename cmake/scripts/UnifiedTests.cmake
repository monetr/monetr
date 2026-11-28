# Runs the Go tests and the interface tests in parallel, both producing Go
# test2json events on stdout. Meant to be run by gotestsum via --raw-command so
# both show up in the same output:
#
#   gotestsum --format testname --raw-command -- cmake -P cmake/scripts/UnifiedTests.cmake
#
# Every COMMAND in a single execute_process starts at the same time as a
# pipeline, which is what gets these running in parallel:
#
#   go test -json | rstest | grep
#
# go test's events are piped into rstest, where the reporter in
# interface/test2json.reporter.ts writes them out line by line alongside its
# own events. Having a single writer means large events from either side can
# never get split and interleaved mid-line.
#
# rsbuild.config.ts prints a banner on stdout when it loads, so grep only keeps
# lines that look like JSON, gotestsum would count anything else as an error.
# It's line buffered so events still stream live instead of showing up in
# chunks.

get_filename_component(SOURCE_DIR "${CMAKE_CURRENT_LIST_DIR}/../.." ABSOLUTE)

execute_process(
  COMMAND go -C ${SOURCE_DIR} test -json -race -cover -coverpkg=github.com/monetr/monetr/server/... -count=1 github.com/monetr/monetr/server/...
  COMMAND ${CMAKE_COMMAND} -E env TEST2JSON=1 TEST2JSON_FORWARD_STDIN=1 -- pnpm exec rstest run
  COMMAND grep --line-buffered "^{"
  WORKING_DIRECTORY ${SOURCE_DIR}/interface
  RESULTS_VARIABLE RESULTS
)

list(GET RESULTS 0 GO_RESULT)
list(GET RESULTS 1 RSTEST_RESULT)
if(NOT GO_RESULT EQUAL 0 OR NOT RSTEST_RESULT EQUAL 0)
  cmake_language(EXIT 1)
endif()
