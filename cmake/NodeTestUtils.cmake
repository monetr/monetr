macro(provision_node_tests CURRENT_SOURCE_DIR)
  if (BUILD_TESTING)
    string(REPLACE "${CMAKE_SOURCE_DIR}/" "" PACKAGE "${CURRENT_SOURCE_DIR}")
    message(STATUS "Preparing tests for: ${PACKAGE}")

    set(PACKAGE_TEST_DIRECTORY ${CMAKE_BINARY_DIR}/tests/${PACKAGE})
    set(PACKAGE_COVERAGE_DIRECTORY ${PACKAGE_TEST_DIRECTORY}/coverage)

    file(MAKE_DIRECTORY ${PACKAGE_COVERAGE_DIRECTORY})
    file(MAKE_DIRECTORY ${PACKAGE_TEST_DIRECTORY})

    file(GLOB SPEC_FILES RELATIVE ${CURRENT_SOURCE_DIR} "${CURRENT_SOURCE_DIR}/*.spec.*")

    foreach(SPEC_FILE IN LISTS SPEC_FILES)
      string(REGEX REPLACE "([a-zA-Z0-9_]+)\\.spec.+" "\\1" SPEC_NAME "${SPEC_FILE}")


      set(TEST_ARGS "--timeout=5000" "--preload=${CMAKE_SOURCE_DIR}/interface/src/setupTests.ts")
      if(TEST_COVERAGE)
        # If we are collecting code coverage then we want to add these flags to jest. Because we are running tests one
        # file at a time we need to pass --watchAll=false in order for jest to properly collect coverage.
        list(APPEND TEST_ARGS "--coverage" "--coverage-reporter=lcov" "--coverage-dir=${PACKAGE_COVERAGE_DIRECTORY}/${SPEC_NAME}")
      endif()

      set(COMMAND_PREFIX)
      if(NOT DEFINED ENV{CI})
        set(COMMAND_PREFIX ${CMAKE_COMMAND} -E env FORCE_COLOR=1 --)
      endif()

      add_test(
        NAME ${PACKAGE}/${SPEC_NAME}
        COMMAND ${COMMAND_PREFIX} ${BUN_EXECUTABLE} test ${TEST_ARGS} ${CURRENT_SOURCE_DIR}/${SPEC_FILE}
        WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/interface
      )
      set_tests_properties(${PACKAGE}/${SPEC_NAME} PROPERTIES 
        FIXTURES_REQUIRED node_modules
        TIMEOUT 45
      )
      set_property(
        TEST ${PACKAGE}/${SPEC_NAME}
        PROPERTY LABELS "interface"
      )
    endforeach()
  endif()
endmacro()
