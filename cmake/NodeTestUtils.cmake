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

      set(TEST_ARGS "run" "--config=${CMAKE_SOURCE_DIR}/interface/rstest.config.ts")
      if(TEST_COVERAGE)
        list(APPEND TEST_ARGS "--coverage.enabled")
      endif()

      add_test(
        NAME ${PACKAGE}/${SPEC_NAME}
        COMMAND ${RSTEST_EXECUTABLE} ${TEST_ARGS} ${CURRENT_SOURCE_DIR}/${SPEC_FILE}
        WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/interface
      )
      set_tests_properties(${PACKAGE}/${SPEC_NAME} PROPERTIES
        FIXTURES_REQUIRED node_modules
        # Node gets really picky about having multiple instances running at the
        # same time, this helps a bit by making sure that we are not putting
        # multiple node instances on the same CPU core.
        PROCESSORS 4
        PROCESSOR_AFFINITY ON
        TIMEOUT 45
      )
      if(TEST_COVERAGE)
        set_property(
          TEST ${PACKAGE}/${SPEC_NAME}
          # Look inside rstest.config.ts, this env variable makes it so we
          # write the code coverage to the right place. But we can provide it
          # every time even if we aren't collecting code coverage.
          PROPERTY ENVIRONMENT "RSTEST_COVERAGE_DIR=${PACKAGE_COVERAGE_DIRECTORY}/${SPEC_NAME}"
        )
      endif()
      set_property(
        TEST ${PACKAGE}/${SPEC_NAME}
        PROPERTY LABELS "interface"
      )
    endforeach()
  endif()
endmacro()
