# We can only produce junit files if we are using gotestsum.
if((TEST_JUNIT) AND (NOT TEST_USE_GOTESTSUM))
  message(FATAL_ERROR "gotestsum must be enabled for JUnit XML output. Please rerun with -DTEST_USE_GOTESTSUM=ON or -DTEST_JUNIT=OFF")
endif()

macro(provision_golang_tests CURRENT_SOURCE_DIR)
  if (BUILD_TESTING)
    if(CMAKE_Go_COMPILER)
      string(REPLACE "${CMAKE_SOURCE_DIR}/" "" PACKAGE "${CMAKE_CURRENT_SOURCE_DIR}")
      message(STATUS "Preparing tests for: ${PACKAGE}")

      execute_process(
        COMMAND ${CMAKE_Go_COMPILER} run ${CMAKE_SOURCE_DIR}/scripts/find_tests.go ${CURRENT_SOURCE_DIR}
        OUTPUT_VARIABLE RAW_TEST_LIST
        OUTPUT_STRIP_TRAILING_WHITESPACE
        WORKING_DIRECTORY ${CURRENT_SOURCE_DIR}
      )

      set(PACKAGE_TEST_DIRECTORY ${CMAKE_BINARY_DIR}/tests/${PACKAGE})
      set(PACKAGE_COVERAGE_DIRECTORY ${PACKAGE_TEST_DIRECTORY}/coverage)
      file(MAKE_DIRECTORY ${PACKAGE_COVERAGE_DIRECTORY})
      file(MAKE_DIRECTORY ${PACKAGE_TEST_DIRECTORY})

      # If we are on windows then generate the executable properly
      if(WIN32)
        set(PACKAGE_TEST_BINARY ${PACKAGE_TEST_DIRECTORY}/runner.exe)
      else()
        set(PACKAGE_TEST_BINARY ${PACKAGE_TEST_DIRECTORY}/runner)
      endif()

      # Default arguments are basically just "compile a binary from the test package" and "put that binary somewhere".
      # We stash the binary deep in the cmake binary directory so we can execute it later. We also co-locate test
      # artifacts with that binary.
      set(TEST_BINARY_BUILD_ARGS "-c" "-o" "${PACKAGE_TEST_BINARY}")
      if(TEST_RACE)
        # If we are doing race testing then we should include this flag on the binary when it is built. This needs to be
        # done in the build stage and not in the execute stage for Golang.
        list(PREPEND TEST_BINARY_BUILD_ARGS "-race")
      endif()
      if(TEST_COVERAGE)
        # If we want to have code coverage in our tests then we should add the cover flags to the test binary we are
        # creating. We also want to make sure to have the coverpkg set because we are collecting coverage from ALL of
        # monetr. Not just from the package we are currently testing.
        list(PREPEND TEST_BINARY_BUILD_ARGS "-cover" "-coverpkg=github.com/monetr/monetr/server/...")
      endif()

      set(TAGS_FLAG)
      if(TEST_GO_TAGS)
        set(TAGS_FLAG "-tags=${TEST_GO_TAGS}")
      endif()


      # Then we create a test that compiles the current package into a test binary we can use later.
      add_test(
        NAME precompile/${PACKAGE}
        COMMAND ${CMAKE_Go_COMPILER} test ${TEST_BINARY_BUILD_ARGS} ${TAGS_FLAG}
        WORKING_DIRECTORY ${CURRENT_SOURCE_DIR}
      )
      # We make sure that we have our go.mod (basically our dependencies) setup before we do this. We also denote this
      # test as a fixture setup. The actual tests for this package then depend on this fixture being setup.
      # TODO Should we do a fixture cleanup to delete this binary afterwards?
      set_tests_properties(
        precompile/${PACKAGE}
        PROPERTIES
        FIXTURES_REQUIRED go.mod
        FIXTURES_SETUP ${PACKAGE}
        RESOURCE_LOCK GO_BUILD_LOCK
      )

      string(REPLACE "\n" ";" RAW_TEST_LIST "${RAW_TEST_LIST}")
      foreach(TEST ${RAW_TEST_LIST})
        # Convert the bar delimited test name into two names or a single high level test.
        string(REPLACE "|" ";" TEST_PARTS "${TEST}")
        list(LENGTH TEST_PARTS TEST_PARTS_COUNT)
        set(REGEX_TEST_NAME)
        set(FRIENDLY_TEST_NAME)
        set(TEST_OUTPUT_FILE_NAME)
        # If the test has two name parts then it is a name and a sub-test
        if("${TEST_PARTS_COUNT}" STREQUAL "2")
          # Break out the parts into their own things.
          list(GET TEST_PARTS 0 TEST_NAMESPACE)
          list(GET TEST_PARTS 1 SUB_TEST)
          string(REPLACE " " "_" SUB_TEST_REGEX "${SUB_TEST}")
          # Build a regex that will only match this one specific test.
          set(REGEX_TEST_NAME "^\\Q${TEST_NAMESPACE}\\E$/^\\Q${SUB_TEST_REGEX}\\E$")
          set(FRIENDLY_TEST_NAME "${TEST_NAMESPACE}/${SUB_TEST}")
          set(TEST_OUTPUT_FILE_NAME "${TEST_NAMESPACE}-${SUB_TEST}")
        else()
          # If there is only a single part, then build a regex for only that test.
          set(REGEX_TEST_NAME "^\\Q${TEST}\\E$")
          set(FRIENDLY_TEST_NAME "${TEST}")
          set(TEST_OUTPUT_FILE_NAME "${TEST}")
        endif()
        string(REPLACE "/" "-" TEST_OUTPUT_FILE_NAME "${TEST_OUTPUT_FILE_NAME}")

        # TODO gotestsum is not included by default, test2json might not be either. So these need to be provisioned
        # before they can be used properly. This is basically just messing around.
        # Conditionally configure gotestsum and its arguments.
        set(GOTESTSUM_MAYBE "")
        if(TEST_USE_GOTESTSUM)
          list(APPEND GOTESTSUM_MAYBE "gotestsum")

          # When we are using gotestsum we want to produce json output files
          file(MAKE_DIRECTORY "${PACKAGE_TEST_DIRECTORY}/json")
          list(APPEND GOTESTSUM_MAYBE "--jsonfile=${PACKAGE_TEST_DIRECTORY}/json/${TEST_OUTPUT_FILE_NAME}.json")

          if(TEST_JUNIT)
            # When we are using gotestsum and we want to output junit files, they should go in their own directory
            file(MAKE_DIRECTORY "${PACKAGE_TEST_DIRECTORY}/junit")
            list(APPEND GOTESTSUM_MAYBE "--junitfile=${PACKAGE_TEST_DIRECTORY}/junit/${TEST_OUTPUT_FILE_NAME}.xml")
          endif()

          # Since we are executing a raw test binary using gotestsum, we need to pass it as a raw command.
          list(APPEND GOTESTSUM_MAYBE "--raw-command" "--")

          # When we are using gotestsum, we need to also format test output for it to be consumed easily.
          list(APPEND GOTESTSUM_MAYBE "${CMAKE_Go_COMPILER}" "tool" "test2json" "-t")
        endif()

        # Now we need to actually build our test arguments. We want to have count=1 to make sure that tests are not
        # cached. This could maybe be turned into an option though? And we _need_ v=true in order to have any usable
        # output for gotestsum or if a test fails in general.
        set(TEST_ARGS "-test.count=1" "-test.v")
        set(COVERAGE_FILE ${PACKAGE_COVERAGE_DIRECTORY}/${TEST_OUTPUT_FILE_NAME}.txt)
        if(TEST_COVERAGE)
          # Since we arent using the goofy binary format this might not be necessary?
          list(APPEND TEST_ARGS -test.coverprofile=${COVERAGE_FILE})
        endif()

        add_test(
          NAME ${PACKAGE}/${FRIENDLY_TEST_NAME}
          COMMAND ${GOTESTSUM_MAYBE} ${PACKAGE_TEST_BINARY} ${TEST_ARGS} -test.run ${REGEX_TEST_NAME}
          WORKING_DIRECTORY ${CURRENT_SOURCE_DIR}
        )
        set_tests_properties(
          ${PACKAGE}/${FRIENDLY_TEST_NAME}
          PROPERTIES
          FIXTURES_REQUIRED "DB;${PACKAGE};go.mocks"
          ENVIRONMENT "GOCOVERDIR=${PACKAGE_COVERAGE_DIRECTORY}"
          SKIP_REGULAR_EXPRESSION " SKIP: "
        )
      endforeach()
    endif()
  endif()
endmacro()
