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

set(LOCAL_DOMAIN "monetr.local")
set(LOCAL_PROTOCOL "https")
set(CLOUD_MAGIC OFF)

if(DEFINED ENV{GITPOD_WORKSPACE_ID})
  message(STATUS "Detected GitPod workspace environment, some local development settings will be adjusted.")
  set(LOCAL_DOMAIN "80-$ENV{GITPOD_WORKSPACE_ID}.$ENV{GITPOD_WORKSPACE_CLUSTER_HOST}")
  set(LOCAL_PROTOCOL "https")
  set(CLOUD_MAGIC ON)
elseif(DEFINED ENV{CODESPACE_NAME})
  message(STATUS "Detected GitHub Codespaces environment, some local development settings will be adjusted.")
  set(LOCAL_DOMAIN "$ENV{CODESPACE_NAME}-80.githubpreview.dev")
  set(LOCAL_PROTOCOL "https")
  set(CLOUD_MAGIC ON)
else()
  set(LOCAL_DOMAIN "monetr.local")
  set(LOCAL_PROTOCOL "https")
  set(CLOUD_MAGIC OFF)
endif()

set(LOCAL_CERTIFICATE_DIR ${CMAKE_BINARY_DIR}/certificates/${LOCAL_DOMAIN})
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
  message(AUTHOR_WARNING "Because you are running on Windows, ${LOCAL_DOMAIN} might not be able to be registered with the hosts file.")
endif()

add_custom_command(
  OUTPUT ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT}
  BYPRODUCTS ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT}
  COMMAND ${SUDO_EXECUTABLE} ${MKCERT_EXECUTABLE} -install
  COMMAND ${MKCERT_EXECUTABLE} -key-file ${LOCAL_CERTIFICATE_KEY} -cert-file ${LOCAL_CERTIFICATE_CERT} ${LOCAL_DOMAIN}
  COMMENT "Setting up local development TLS certificate, this is required for OAuth2. You may be prompted for a password"
  DEPENDS ${MKCERT_EXECUTABLE}
)

add_custom_command(
  OUTPUT ${LOCAL_HOSTS_MARKER}
  BYPRODUCTS ${LOCAL_HOSTS_MARKER}
  COMMAND ${SUDO_EXECUTABLE} ${HOSTESS_EXECUTABLE} add ${LOCAL_DOMAIN} 127.0.0.1
  COMMAND ${CMAKE_COMMAND} -E touch ${LOCAL_HOSTS_MARKER}
  COMMENT "Setting up ${LOCAL_DOMAIN} domain with your /etc/hosts file. You may be prompted for a password"
  DEPENDS ${HOSTESS_EXECUTABLE}
)

add_custom_target(development.certificates DEPENDS ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT})
add_custom_target(development.hostsfile DEPENDS ${LOCAL_HOSTS_MARKER})

set(LOCAL_COMPOSE_FILE ${CMAKE_SOURCE_DIR}/compose/docker-compose.monetr.yaml)
set(DOCUMENTATION_COMPOSE_FILE ${CMAKE_SOURCE_DIR}/compose/docker-compose.documentation.yaml)

set(ENV{LOCAL_CERTIFICATE_DIR} ${LOCAL_CERTIFICATE_DIR})
set(BASE_ARGS "--project-directory" "${CMAKE_SOURCE_DIR}")

# Check to see if the current user has local settings configured for development.
if(EXISTS ${HOME}/.monetr/development.env)
  message(STATUS "Detected development.env file at: ${HOME}/.monetr/development.env")
  message(STATUS "  It will be used if you start the local development environment.")
  list(APPEND BASE_ARGS "--env-file=${HOME}/.monetr/development.env")
endif()

set(DEVELOPMENT_COMPOSE_ARGS "-f" "${LOCAL_COMPOSE_FILE}" ${BASE_ARGS})
set(DOCUMENTATION_COMPOSE_ARGS "-f" "${DOCUMENTATION_COMPOSE_FILE}" ${BASE_ARGS})
set(ALL_COMPOSE_ARGS "-f" "${LOCAL_COMPOSE_FILE}" "-f" "${DOCUMENTATION_COMPOSE_FILE}" ${BASE_ARGS})

add_custom_target(
  development.monetr.up
  COMMENT "Starting monetr using Docker compose locally..."
  COMMAND ${CMAKE_COMMAND} -E env LOCAL_PROTOCOL=${LOCAL_PROTOCOL} LOCAL_DOMAIN=${LOCAL_DOMAIN} ${DOCKER_EXECUTABLE} compose ${DEVELOPMENT_COMPOSE_ARGS} up --wait --remove-orphans
  COMMAND ${CMAKE_COMMAND} -E echo "-- ========================================================================="
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- monetr is now running locally."
  COMMAND ${CMAKE_COMMAND} -E echo "-- You can access monetr via ${LOCAL_PROTOCOL}://${LOCAL_DOMAIN}"
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
    ${NODE_MODULES}
    ${HTML_EMAIL_TEMPLATES}
    ${TEXT_EMAIL_TEMPLATES}
)

if(NOT CLOUD_MAGIC)
  # If we are not in the cloud then we need to make sure we setup TLS stuff locally.
  # TODO Put the hosts file and tls stuff behind a feature flag.
  add_dependencies(development.monetr.up development.certificates development.hostsfile)
  # if(NOT DEFINED ENV{CI})
  #   add_dependencies(development.monetr.up development.certificates development.hostsfile)
  # else()
  #   add_dependencies(development.monetr.up development.certificates)
  # endif()
else()
  # If we are in the cloud we need to also make sure nginx uses the right config.
  set_target_properties(development.monetr.up PROPERTY ENVIRONMENT
    "NGINX_CONFIG_NAME=nginx-cloud.conf"
    "NGINX_PORT=80"
  )
endif()

add_custom_target(
  development.documentation.up
  COMMENT "Starting documentation using Docker compose locally..."
  COMMAND ${DOCKER_EXECUTABLE} compose ${DOCUMENTATION_COMPOSE_ARGS} up --wait --remove-orphans
  COMMAND ${CMAKE_COMMAND} -E echo "-- ========================================================================="
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- Documentation is now running locally."
  COMMAND ${CMAKE_COMMAND} -E echo "-- You can access monetr via http://localhost:8000/documentation"
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- Changes you make to the documentation will automatically hot reload in"
  COMMAND ${CMAKE_COMMAND} -E echo "-- your browser."
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- When you are done you can shutdown the local development environment using:"
  COMMAND ${CMAKE_COMMAND} -E echo "--   make shutdown"
  COMMAND ${CMAKE_COMMAND} -E echo "--     or:"
  COMMAND ${CMAKE_COMMAND} -E echo "--   make clean"
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- ========================================================================="
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)

add_custom_target(
  development.logs
  COMMENT "Tailing logs from monetr's local development environment"
  COMMAND ${DOCKER_EXECUTABLE} compose ${ALL_COMPOSE_ARGS} logs -f
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)

add_custom_target(
  development.down
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${DEVELOPMENT_COMPOSE_ARGS} exec monetr monetr development clean:plaid || exit 0
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${DEVELOPMENT_COMPOSE_ARGS} exec monetr monetr development clean:stripe || exit 0
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} down --remove-orphans -v || exit 0
  COMMAND_EXPAND_LISTS
  USES_TERMINAL
)

add_custom_target(
  development.shell
  COMMENT "Spawning a shell in the specified CONTAINER."
  COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} exec $(CONTAINER) /bin/sh
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
