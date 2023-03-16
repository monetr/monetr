cmake_minimum_required(VERSION 3.0)

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake")

project(monetr
        DESCRIPTION "monetr is an open source budgeting and financial planning application"
        HOMEPAGE_URL "https://monetr.app"
        LANGUAGES Go)

# Place targets in bin/ or lib/ for all build configurations
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
foreach(CFGNAME ${CMAKE_CONFIGURATION_TYPES})
  string(TOUPPER ${CFGNAME} CFGNAME)
  set(CMAKE_RUNTIME_OUTPUT_DIRECTORY_${CFGNAME} ${CMAKE_BINARY_DIR}/bin)
  set(CMAKE_LIBRARY_OUTPUT_DIRECTORY_${CFGNAME} ${CMAKE_BINARY_DIR}/lib)
  set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY_${CFGNAME} ${CMAKE_BINARY_DIR}/lib)
endforeach()

add_custom_command(OUTPUT "${CMAKE_CURRENT_SOURCE_DIR}/build/monetr"
                   COMMAND ${CMAKE_Go_COMPILER} build -o ${TARGET_FILE} github.com/monetr/monetr/pkg/cmd)

add_custom_target(monetr
                  COMMENT "Building monetr!"
                  DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/build/monetr")
