macro(provision_spelling_tests CURRENT_SOURCE_DIR)
  if(BUILD_TESTING)
    string(REPLACE "${CMAKE_SOURCE_DIR}/" "" PACKAGE "${CURRENT_SOURCE_DIR}")
    message(STATUS "Preparing tests for: ${PACKAGE}")

    set(PACKAGE_TEST_DIRECTORY ${CMAKE_BINARY_DIR}/tests/${PACKAGE})
    file(MAKE_DIRECTORY ${PACKAGE_TEST_DIRECTORY})

    file(GLOB SPEC_FILES RELATIVE ${CURRENT_SOURCE_DIR} "${CURRENT_SOURCE_DIR}/*.md*")

    foreach(SPEC_FILE IN LISTS SPEC_FILES)
      string(REGEX REPLACE "([a-zA-Z0-9_]+)\\.md.+" "\\1" SPEC_NAME "${SPEC_FILE}")

      set(TEST_ARGS 
        # Language is english US
        "-l" "en-US" 
        # We want all the default plugins AND frontmatter
        "-p" "spell" "indefinite-article" "repeated-words" "syntax-mentions" "syntax-urls" "frontmatter" 
        # Check the frontmatter description key
        "--frontmatter-keys" "description"
        # Too many technical words, dont make suggestions
        "--no-suggestions"
      )
      add_test(
        NAME ${PACKAGE}/${SPEC_NAME}
        COMMAND ${SPELLCHECKER_EXECUTABLE} ${TEST_ARGS} -f ${CURRENT_SOURCE_DIR}/${SPEC_FILE}
        WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
      )
      set_tests_properties(${PACKAGE}/${SPEC_NAME} PROPERTIES 
        FIXTURES_REQUIRED node_modules
        TIMEOUT 45
      )
      set_property(
        TEST ${PACKAGE}/${SPEC_NAME}
        PROPERTY LABELS "docs"
      )
    endforeach()
  endif()
endmacro()
