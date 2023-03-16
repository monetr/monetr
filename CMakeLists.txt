cmake_minimum_required(VERSION 3.0)

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake")

project(monetr
        DESCRIPTION "monetr is an open source budgeting and financial planning application"
        HOMEPAGE_URL "https://monetr.app"
        LANGUAGES Go)

set(BUILD_DIR     "${PROJECT_SOURCE_DIR}/build")
set(MONETR_BINARY "${BUILD_DIR}/monetr")

add_custom_command(OUTPUT "${MONETR_BINARY}"
                   WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
                   COMMAND ${CMAKE_Go_COMPILER} build
                           -o ${MONETR_BINARY}
                           github.com/monetr/monetr/pkg/cmd)

add_custom_target(monetr
                  COMMENT "Building monetr!"
                  WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
                  DEPENDS "${MONETR_BINARY}")
