include(GolangTestUtils)

provision_golang_tests(${CMAKE_CURRENT_SOURCE_DIR})

# Apple has some weird DNS shit that doesn't always work for some reason with
# golang. In order to work around this if we are on an Apple machine then we
# need to change the DNS resolver to be the pure go one instead of the cgo one.
# This will make it so that when we run tests or database migrations we can
# actually find the server. This was added because without it monetr literally
# couldn't even resolve localhost.
# https://pkg.go.dev/net#hdr-Name_Resolution
set(DNS_FUCKERY)
if(APPLE)
  set(DNS_FUCKERY ${CMAKE_COMMAND} -E env GODEBUG=netdns=go --)
endif()

add_custom_target(
  development.migrate
  COMMAND ${DNS_FUCKERY} ${CMAKE_Go_COMPILER} run ${MONETR_CLI_PKG} database migrate
  COMMENT "Performing database migrations"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)
