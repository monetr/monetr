message(STATUS "Local development is available via docker compose")
include(GolangUtils)

# mkcert for local development
set(MKCERT_EXECUTABLE ${GO_BIN_DIR}/mkcert)
set(MKCERT_VERSION latest)
go_install(
  OUTPUT ${MKCERT_EXECUTABLE}
  PACKAGE "filippo.io/mkcert"
  VERSION ${MKCERT_VERSION}
)

# hostess for local development
set(HOSTESS_EXECUTABLE ${GO_BIN_DIR}/hostess)
set(HOSTESS_VERSION latest)
go_install(
  OUTPUT ${HOSTESS_EXECUTABLE}
  PACKAGE "github.com/cbednarski/hostess"
  VERSION ${HOSTESS_VERSION}
)

if (NOT MONETR_LOCAL_DOMAIN)
  set(MONETR_LOCAL_DOMAIN "monetr.local")
endif()
set(LOCAL_PROTOCOL "https")
set(CLOUD_MAGIC OFF)

if(DEFINED ENV{GITPOD_WORKSPACE_ID})
  message(STATUS "Detected GitPod workspace environment, some local development settings will be adjusted.")
  set(MONETR_LOCAL_DOMAIN "80-$ENV{GITPOD_WORKSPACE_ID}.$ENV{GITPOD_WORKSPACE_CLUSTER_HOST}")
  set(LOCAL_PROTOCOL "https")
  set(CLOUD_MAGIC ON)
elseif(DEFINED ENV{CODESPACE_NAME})
  message(STATUS "Detected GitHub Codespaces environment, some local development settings will be adjusted.")
  set(MONETR_LOCAL_DOMAIN "$ENV{CODESPACE_NAME}-80.githubpreview.dev")
  set(LOCAL_PROTOCOL "https")
  set(CLOUD_MAGIC ON)
else()
  set(LOCAL_PROTOCOL "https")
  set(CLOUD_MAGIC OFF)
endif()

# When we are running locally we want nginx to handle our TLS termination with a self-signed certificate. But if we are
# using something like GitPod or Github workspaces then they will handle TLS termination for us.
set(NGINX_PORT "443")
set(NGINX_CONFIG_FILE "${CMAKE_SOURCE_DIR}/compose/nginx.conf")
if (CLOUD_MAGIC)
  set(NGINX_PORT "80")
  set(NGINX_CONFIG_FILE "${CMAKE_SOURCE_DIR}/compose/nginx-cloud.conf")
endif()

if(TELLER_CERTIFICATE AND TELLER_PRIVATE_KEY)
  file(MAKE_DIRECTORY ${CMAKE_BINARY_DIR}/compose/teller)
  configure_file(${TELLER_CERTIFICATE} ${CMAKE_BINARY_DIR}/compose/teller/certificate.pem COPYONLY)
  configure_file(${TELLER_PRIVATE_KEY} ${CMAKE_BINARY_DIR}/compose/teller/private_key.pem COPYONLY)
  set(MONETR_TELLER_CERTIFICATE "/build/build/compose/teller/certificate.pem")
  set(MONETR_TELLER_PRIVATE_KEY "/build/build/compose/teller/private_key.pem")
  message(STATUS "Detected teller certificates, they will configured for local development")
else()
  # If we are not using certificates for teller then make sure these values are cleared for when we configure the docker
  # compose manifest.
  set(MONETR_TELLER_CERTIFICATE "")
  set(MONETR_TELLER_PRIVATE_KEY "")
endif()


set(LOCAL_CERTIFICATE_DIR ${CMAKE_BINARY_DIR}/certificates/${MONETR_LOCAL_DOMAIN})
set(LOCAL_CERTIFICATE_KEY ${LOCAL_CERTIFICATE_DIR}/key.pem)
set(LOCAL_CERTIFICATE_CERT ${LOCAL_CERTIFICATE_DIR}/cert.pem)
file(MAKE_DIRECTORY ${LOCAL_CERTIFICATE_DIR})

set(LOCAL_HOSTS_MARKER ${CMAKE_BINARY_DIR}/etc-hosts.marker)

set(SUDO_EXECUTABLE "")
if(NOT WIN32)
  set(SUDO_EXECUTABLE "sudo")
endif()

if(WIN32)
  message(AUTHOR_WARNING "Because you are running on Windows, TLS might not be able to be provisioned for the local development environment.")
  message(AUTHOR_WARNING "Because you are running on Windows, ${MONETR_LOCAL_DOMAIN} might not be able to be registered with the hosts file.")
endif()

add_custom_command(
  OUTPUT ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT}
  BYPRODUCTS ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT}
  COMMAND ${SUDO_EXECUTABLE} ${MKCERT_EXECUTABLE} -install
  COMMAND ${MKCERT_EXECUTABLE} -key-file ${LOCAL_CERTIFICATE_KEY} -cert-file ${LOCAL_CERTIFICATE_CERT} ${MONETR_LOCAL_DOMAIN}
  COMMENT "Setting up local development TLS certificate, this is required for OAuth2. You may be prompted for a password"
  DEPENDS ${MKCERT_EXECUTABLE}
)

add_custom_command(
  OUTPUT ${LOCAL_HOSTS_MARKER}
  BYPRODUCTS ${LOCAL_HOSTS_MARKER}
  COMMAND ${SUDO_EXECUTABLE} ${HOSTESS_EXECUTABLE} add ${MONETR_LOCAL_DOMAIN} 127.0.0.1
  COMMAND ${CMAKE_COMMAND} -E touch ${LOCAL_HOSTS_MARKER}
  COMMENT "Setting up ${MONETR_LOCAL_DOMAIN} domain with your /etc/hosts file. You may be prompted for a password"
  DEPENDS ${HOSTESS_EXECUTABLE}
)

add_custom_target(development.certificates DEPENDS ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT})
add_custom_target(development.hostsfile DEPENDS ${LOCAL_HOSTS_MARKER})

set(COMPOSE_FILE_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/docker-compose.monetr.yaml.in)
if (NGROK_AUTH OR DEFINED ENV{NGROK_AUTH} OR NGROK_ENABLED)
  set(NGROK_AUTH "${NGROK_AUTH}")
  if(NOT NGROK_AUTH)
    set(NGROK_AUTH "$ENV{NGROK_AUTH}")
  endif()

  set(NGROK_HOSTNAME "${NGROK_HOSTNAME}")
  if(NOT NGROK_HOSTNAME)
    set(NGROK_HOSTNAME "$ENV{NGROK_HOSTNAME}")
  endif()
  list(APPEND COMPOSE_FILE_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/docker-compose-ngrok.monetr.yaml.in)
  message(STATUS "Detected ngrok credentials, webhooks will be enabled for local development.")
  if(NGROK_HOSTNAME)
    message(STATUS "  Webhook domain: ${NGROK_HOSTNAME}")
  endif()
else()
  message(STATUS "No ngrok credentials detected, webhooks will not be enabled for local development.")
endif()

set(COMPOSE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/development)
file(MAKE_DIRECTORY ${COMPOSE_OUTPUT_DIRECTORY})

set(COMPOSE_FILES)
foreach(COMPOSE_FILE_TEMPLATE ${COMPOSE_FILE_TEMPLATES})
  set(COMPOSE_FILE_OUTPUT "${COMPOSE_FILE_TEMPLATE}")
  string(REPLACE ".in" "" COMPOSE_FILE_OUTPUT "${COMPOSE_FILE_OUTPUT}")
  string(REPLACE "${CMAKE_SOURCE_DIR}/compose" "${COMPOSE_OUTPUT_DIRECTORY}" COMPOSE_FILE_OUTPUT "${COMPOSE_FILE_OUTPUT}")
  configure_file("${COMPOSE_FILE_TEMPLATE}" "${COMPOSE_FILE_OUTPUT}" @ONLY)
  list(APPEND COMPOSE_FILES "-f" "${COMPOSE_FILE_OUTPUT}")
endforeach()


set(ENV{LOCAL_CERTIFICATE_DIR} ${LOCAL_CERTIFICATE_DIR})
set(BASE_ARGS "--project-directory" "${CMAKE_SOURCE_DIR}")

# Check to see if the current user has local settings configured for development.
if(EXISTS ${HOME}/.monetr/development.env)
  message(STATUS "Detected development.env file at: ${HOME}/.monetr/development.env")
  message(STATUS "  It will be used if you start the local development environment.")
  list(APPEND BASE_ARGS "--env-file=${HOME}/.monetr/development.env")
endif()

set(DEVELOPMENT_COMPOSE_ARGS ${COMPOSE_FILES} ${BASE_ARGS})
set(ALL_COMPOSE_ARGS ${COMPOSE_FILES} ${BASE_ARGS})

add_custom_target(
  development.monetr.up
  COMMENT "Starting monetr using Docker compose locally..."
  COMMAND ${CMAKE_COMMAND} -E env LOCAL_PROTOCOL=${LOCAL_PROTOCOL} MONETR_LOCAL_DOMAIN=${MONETR_LOCAL_DOMAIN} ${DOCKER_EXECUTABLE} compose ${DEVELOPMENT_COMPOSE_ARGS} up --wait --remove-orphans
  COMMAND ${CMAKE_COMMAND} -E echo "-- ========================================================================="
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- monetr is now running locally."
  COMMAND ${CMAKE_COMMAND} -E echo "-- You can access monetr via ${LOCAL_PROTOCOL}://${MONETR_LOCAL_DOMAIN}"
  COMMAND ${CMAKE_COMMAND} -E echo "-- Emails sent during development can be seen at ${LOCAL_PROTOCOL}://${MONETR_LOCAL_DOMAIN}/mail"
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- When you are done you can shutdown the local development environment using:"
  COMMAND ${CMAKE_COMMAND} -E echo "--   make shutdown"
  COMMAND ${CMAKE_COMMAND} -E echo "--     or:"
  COMMAND ${CMAKE_COMMAND} -E echo "--   make clean"
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- ========================================================================="
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
  DEPENDS
    ${SIMPLE_ICONS}
    dependencies.node_modules
    build.email
)

if(NOT CLOUD_MAGIC)
  add_dependencies(development.monetr.up development.certificates development.hostsfile)
endif()

add_custom_target(
  development.logs
  COMMENT "Tailing logs from monetr's local development environment"
  COMMAND ${DOCKER_EXECUTABLE} compose ${ALL_COMPOSE_ARGS} logs -f $(CONTAINER)
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)

if(DOCKER_SERVER) 
  add_custom_target(
    development.down
    COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${DEVELOPMENT_COMPOSE_ARGS} exec monetr monetr -c /build/compose/monetr.yaml development clean:plaid || exit 0
    COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${DEVELOPMENT_COMPOSE_ARGS} exec monetr monetr -c /build/compose/monetr.yaml development clean:stripe || exit 0
    COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} down --remove-orphans -v || exit 0
    COMMAND_EXPAND_LISTS
    USES_TERMINAL
  )
else()
  add_custom_target(
    development.down
    COMMAND ${CMAKE_COMMAND} -E echo "-- Docker server is not running, development.down is a no-op"
  )
endif()

add_custom_target(
  development.shell
  COMMENT "Spawning a shell in the specified CONTAINER."
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} exec $(CONTAINER) /bin/sh
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)

add_custom_target(
  development.restart
  COMMENT "Restarting the specified CONTAINER"
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} restart $(CONTAINER)
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)

add_custom_target(
  development.shell.sql
  COMMENT "Spawning a SQL shell inside of PostgreSQL."
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} exec postgres psql -U postgres
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)

add_custom_target(
  development.shell.redis
  COMMENT "Spawning a Redis shell."
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} exec redis redis-cli
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)
