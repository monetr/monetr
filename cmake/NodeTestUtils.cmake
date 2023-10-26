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

      set(TEST_ARGS "--config" "${UI_SRC_DIR}/jest.config.ts")
      if(TEST_COVERAGE)
        # If we are collecting code coverage then we want to add these flags to jest. Because we are running tests one
        # file at a time we need to pass --watchAll=false in order for jest to properly collect coverage.
        list(APPEND TEST_ARGS "--coverage" "--coverageDirectory=${PACKAGE_COVERAGE_DIRECTORY}/${SPEC_NAME}" "--watchAll=false")
      endif()

      add_test(
        NAME ${PACKAGE}/${SPEC_NAME}
        COMMAND ${JEST_EXECUTABLE} ${TEST_ARGS} ${CURRENT_SOURCE_DIR}/${SPEC_FILE}
        WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
      )
      set_tests_properties(${PACKAGE}/${SPEC_NAME} PROPERTIES FIXTURES_REQUIRED node_modules)
    endforeach()
  endif()
endmacro()
