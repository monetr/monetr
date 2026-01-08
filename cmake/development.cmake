###############################################################################
# This entire file contains a ton of steps that are used to create the local
# development environment for monetr. monetr has several different options for
# how it can be run, and as a result it needs a local development environment
# that can be somewhat customized based on what you are working on.
# The scripts below prepare this customization
###############################################################################

message(STATUS "Local development is available via docker compose")
include(GolangUtils)

# These two directories are used for the files generated as part of local dev.
# The compose directory will contain a few docker compose files that have had
# some values replaced via `configure_file`. All of the compose files in this
# directory are eventually passed to the docker compose up command.
set(COMPOSE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/development)
file(MAKE_DIRECTORY ${COMPOSE_OUTPUT_DIRECTORY})
# The nginx directory is a child of the development directory. This directory
# contains stubs of templated nginx configuration as well as the final nginx
# configuration file.
set(NGINX_DIRECTORY ${COMPOSE_OUTPUT_DIRECTORY}/nginx)
file(MAKE_DIRECTORY ${NGINX_DIRECTORY})

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

# If the developer has not customized the local domain in their cmake user
# presets file or otherwise then we fall back to `monetr.local`. This is the
# base name though and is no longer used independently. Instead it is prefixed
# by a subdomain for each service.
if (NOT MONETR_LOCAL_DOMAIN)
  set(MONETR_LOCAL_DOMAIN "monetr.local")
endif()
# The certificate we will create though will always be a wildcard certificate.
set(WILDCARD_DOMAIN "*.${MONETR_LOCAL_DOMAIN}")
set(LOCAL_PROTOCOL "https")
set(CLOUD_MAGIC OFF)

# If we are in gitpod or github codespaces then we have some basic handling for
# that here. But it probably does not even work anymore.
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
set(NGINX_CONFIG_FILE "${COMPOSE_OUTPUT_DIRECTORY}/nginx/nginx.conf")
if (CLOUD_MAGIC)
  set(NGINX_PORT "80")
  set(NGINX_CONFIG_FILE "${CMAKE_SOURCE_DIR}/compose/nginx-cloud.conf")
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

# The two subdomains that we will be creating in the hosts file. Others might
# get added to this list as we process the developer's config, but these two
# will always be present.
set(LOCAL_DOMAINS "my.${MONETR_LOCAL_DOMAIN}" "mail.${MONETR_LOCAL_DOMAIN}")

# Create a custom command and target for provisioning the custom TLS
# certificate that monetr will use for local development.
add_custom_command(
  OUTPUT ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT}
  BYPRODUCTS ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT}
  COMMAND ${SUDO_EXECUTABLE} ${MKCERT_EXECUTABLE} -install
  COMMAND ${MKCERT_EXECUTABLE} -key-file ${LOCAL_CERTIFICATE_KEY} -cert-file ${LOCAL_CERTIFICATE_CERT} ${WILDCARD_DOMAIN}
  COMMENT "Setting up local development TLS certificate (*.${MONETR_LOCAL_DOMAIN}), this is required for OAuth2. You may be prompted for a password"
  DEPENDS ${MKCERT_EXECUTABLE}
)
add_custom_target(development.certificates DEPENDS ${LOCAL_CERTIFICATE_KEY} ${LOCAL_CERTIFICATE_CERT})

########################################################################################################################
# This section determines which compose files will be used when the development environment is started. Compose files
# are "merged" by docker at runtime, so this is a simple way of providing some customizability to local development.
########################################################################################################################

# Will be a list of template files that we are using to build our final nginx config.
# This will always contain at least the mail config.
set(NGINX_CONFIG_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/mail.nginx.conf.in)

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

  # If we have ngrok enabled then include it's nginx config in our final config.
  list(APPEND NGINX_CONFIG_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/ngrok.nginx.conf.in)
  # and add the domain name.
  list(APPEND LOCAL_DOMAINS "ngrok.${MONETR_LOCAL_DOMAIN}")
else()
  message(STATUS "No ngrok credentials detected, webhooks will not be enabled for local development.")
endif()

# Setup the KMS backend based on the config.
if("${MONETR_KMS_PROVIDER}" STREQUAL "aws")
  message(STATUS "AWS KMS (Local) will be used for local development as the KMS provider")
  list(APPEND COMPOSE_FILE_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/docker-compose.aws-kms.yaml.in)
elseif("${MONETR_KMS_PROVIDER}" STREQUAL "vault")
  message(STATUS "Vault Transit (Local) will be used for local development as the KMS provider")
  # If we are using the vault KMS provider, then make a vault directory in our build tree. And take the vault config
  # file and template it into that directory for later.
  file(MAKE_DIRECTORY ${COMPOSE_OUTPUT_DIRECTORY}/vault)
  set(VAULT_TOKEN_FILE ${COMPOSE_OUTPUT_DIRECTORY}/vault/token.txt)
  if(NOT EXISTS ${VAULT_TOKEN_FILE})
    string(RANDOM LENGTH 24 ALPHABET abcdefghijklmnopqrstuvwxyz1234567890 VAULT_ROOT_TOKEN)
    set(VAULT_ROOT_TOKEN "dev-${VAULT_ROOT_TOKEN}")
    message(STATUS "  Writing vault token file: ${VAULT_ROOT_TOKEN}")
    file(WRITE ${VAULT_TOKEN_FILE} "${VAULT_ROOT_TOKEN}")
  else()
    file(READ ${VAULT_TOKEN_FILE} VAULT_ROOT_TOKEN)
    message(STATUS "  Using existing vault token file: ${VAULT_ROOT_TOKEN}")
  endif()
  file(READ ${VAULT_TOKEN_FILE} VAULT_ROOT_TOKEN)
  configure_file("${CMAKE_SOURCE_DIR}/compose/vault-config.toml.in" "${COMPOSE_OUTPUT_DIRECTORY}/vault/config.toml" @ONLY)
  # And then add our vault container to our compose list.
  list(APPEND COMPOSE_FILE_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/docker-compose.vault-kms.yaml.in)
  list(APPEND NGINX_CONFIG_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/vault.nginx.conf.in)
  list(APPEND LOCAL_DOMAINS "vault.${MONETR_LOCAL_DOMAIN}")
elseif("${MONETR_KMS_PROVIDER}" STREQUAL "openbao")
  message(STATUS "OpenBao Transit (Local) will be used for local development as the KMS provider")
  # If we are using the OpenBao KMS provider, then make a OpenBao directory in
  # our build tree. And take the config file and template it into that
  # directory for later.
  file(MAKE_DIRECTORY ${COMPOSE_OUTPUT_DIRECTORY}/openbao)
  set(VAULT_TOKEN_FILE ${COMPOSE_OUTPUT_DIRECTORY}/openbao/token.txt)
  if(NOT EXISTS ${VAULT_TOKEN_FILE})
    string(RANDOM LENGTH 24 ALPHABET abcdefghijklmnopqrstuvwxyz1234567890 VAULT_ROOT_TOKEN)
    set(VAULT_ROOT_TOKEN "dev-${VAULT_ROOT_TOKEN}")
    message(STATUS "  Writing openbao token file: ${VAULT_ROOT_TOKEN}")
    file(WRITE ${VAULT_TOKEN_FILE} "${VAULT_ROOT_TOKEN}")
  else()
    file(READ ${VAULT_TOKEN_FILE} VAULT_ROOT_TOKEN)
    message(STATUS "  Using existing openbao token file: ${VAULT_ROOT_TOKEN}")
  endif()
  file(READ ${VAULT_TOKEN_FILE} VAULT_ROOT_TOKEN)
  configure_file("${CMAKE_SOURCE_DIR}/compose/openbao-config.toml.in" "${COMPOSE_OUTPUT_DIRECTORY}/openbao/config.toml" @ONLY)
  # And then add our openbao container to our compose list.
  list(APPEND COMPOSE_FILE_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/docker-compose.openbao-kms.yaml.in)
  list(APPEND NGINX_CONFIG_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/openbao.nginx.conf.in)
  list(APPEND LOCAL_DOMAINS "openbao.${MONETR_LOCAL_DOMAIN}")
elseif("${MONETR_KMS_PROVIDER}" STREQUAL "")
  set(MONETR_KMS_PROVIDER "plaintext")
elseif("${MONETR_KMS_PROVIDER}" STREQUAL "plaintext")
  set(MONETR_KMS_PROVIDER "plaintext")
else()
  message(FATAL "Invalid KMS provider specified, MONETR_KMS_PROVIDER=${MONETR_KMS_PROVIDER}\nValid options are: aws, vault, plaintext")
endif()

# Setup the storage backend based on the config.
if("${MONETR_STORAGE_PROVIDER}" STREQUAL "s3")
  set(MONETR_STORAGE_ENABLED "true")
  message(STATUS "S3 storage will be used for local development")
  list(APPEND COMPOSE_FILE_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/docker-compose.s3-storage.yaml.in)
  list(APPEND NGINX_CONFIG_TEMPLATES ${CMAKE_SOURCE_DIR}/compose/s3.nginx.conf.in)
  list(APPEND LOCAL_DOMAINS "s3.${MONETR_LOCAL_DOMAIN}")
elseif("${MONETR_STORAGE_PROVIDER}" STREQUAL "filesystem")
  set(MONETR_STORAGE_ENABLED "true")
  file(MAKE_DIRECTORY ${COMPOSE_OUTPUT_DIRECTORY}/storage)
  message(STATUS "Filesystem storage will be used for local development")
  message(STATUS "  Uploaded files will be available at: ${COMPOSE_OUTPUT_DIRECTORY}/storage")
elseif("${MONETR_STORAGE_PROVIDER}" STREQUAL "")
  set(MONETR_STORAGE_ENABLED "false")
endif()

# Once the list of compose file templates has been built, actually generate the template files and build our arguments
# for docker compose.

message(DEBUG "  Compose Files: ${COMPOSE_FILE_TEMPLATES}")

set(COMPOSE_FILES)
foreach(COMPOSE_FILE_TEMPLATE ${COMPOSE_FILE_TEMPLATES})
  set(COMPOSE_FILE_OUTPUT "${COMPOSE_FILE_TEMPLATE}")
  string(REPLACE ".in" "" COMPOSE_FILE_OUTPUT "${COMPOSE_FILE_OUTPUT}")
  string(REPLACE "${CMAKE_SOURCE_DIR}/compose" "${COMPOSE_OUTPUT_DIRECTORY}" COMPOSE_FILE_OUTPUT "${COMPOSE_FILE_OUTPUT}")
  configure_file("${COMPOSE_FILE_TEMPLATE}" "${COMPOSE_FILE_OUTPUT}" @ONLY)
  list(APPEND COMPOSE_FILES "-f" "${COMPOSE_FILE_OUTPUT}")
endforeach()


# Take all of the nginx configs that we are using and template them out.
foreach(NGINX_CONFIG_TEMPLATE ${NGINX_CONFIG_TEMPLATES})
  get_filename_component(NGINX_CONFIG_TEMPLATE_OUTPUT "${NGINX_CONFIG_TEMPLATE}" NAME_WLE)
  message(DEBUG "  Using nginx config part: ${NGINX_CONFIG_TEMPLATE_OUTPUT}")
  configure_file("${NGINX_CONFIG_TEMPLATE}" "${NGINX_DIRECTORY}/${NGINX_CONFIG_TEMPLATE_OUTPUT}" @ONLY)
endforeach()

# And then find all of the nginx configs that we generated.
set(S3_NGINX_CONFIG_FILE "${NGINX_DIRECTORY}/s3.nginx.conf")
set(OPENBAO_NGINX_CONFIG_FILE "${NGINX_DIRECTORY}/openbao.nginx.conf")
set(VAULT_NGINX_CONFIG_FILE "${NGINX_DIRECTORY}/vault.nginx.conf")
set(NGROK_NGINX_CONFIG_FILE "${NGINX_DIRECTORY}/ngrok.nginx.conf")
set(MAIL_NGINX_CONFIG_FILE "${NGINX_DIRECTORY}/mail.nginx.conf")

if(EXISTS "${S3_NGINX_CONFIG_FILE}")
  file(READ "${S3_NGINX_CONFIG_FILE}" S3_NGINX_CONFIG)
endif()

if(EXISTS "${OPENBAO_NGINX_CONFIG_FILE}")
  file(READ "${OPENBAO_NGINX_CONFIG_FILE}" OPENBAO_NGINX_CONFIG)
endif()

if(EXISTS "${VAULT_NGINX_CONFIG_FILE}")
  file(READ "${VAULT_NGINX_CONFIG_FILE}" VAULT_NGINX_CONFIG)
endif()

if(EXISTS "${NGROK_NGINX_CONFIG_FILE}")
  file(READ "${NGROK_NGINX_CONFIG_FILE}" NGROK_NGINX_CONFIG)
endif()

if(EXISTS "${MAIL_NGINX_CONFIG_FILE}")
  file(READ "${MAIL_NGINX_CONFIG_FILE}" MAIL_NGINX_CONFIG)
endif()

# And template them into the final config.
configure_file("${CMAKE_SOURCE_DIR}/compose/nginx.conf.in" "${NGINX_CONFIG_FILE}" @ONLY)

########################################################################################################################

add_custom_target(
  development.hostsfile
  COMMAND ${CMAKE_COMMAND} -E true
  COMMENT "Setup local domains in /etc/hosts"
)

foreach(LOCAL_DOMAIN ${LOCAL_DOMAINS})
  add_custom_target(
    development.hostsfile.${LOCAL_DOMAIN}
    BYPRODUCTS ${LOCAL_HOSTS_MARKER}-${LOCAL_DOMAIN}
    COMMAND ${SUDO_EXECUTABLE} ${HOSTESS_EXECUTABLE} add ${LOCAL_DOMAIN} 127.0.0.1
    COMMAND ${CMAKE_COMMAND} -E touch ${LOCAL_HOSTS_MARKER}-${LOCAL_DOMAIN}
    COMMENT "Setting up ${LOCAL_DOMAIN} domain with your /etc/hosts file. You may be prompted for a password"
    DEPENDS ${HOSTESS_EXECUTABLE}
  )
  add_dependencies(development.hostsfile development.hostsfile.${LOCAL_DOMAIN})
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
  COMMAND ${CMAKE_COMMAND} -E env LOCAL_PROTOCOL=${LOCAL_PROTOCOL} MONETR_LOCAL_DOMAIN=${MONETR_LOCAL_DOMAIN} ${DOCKER_EXECUTABLE} compose ${DEVELOPMENT_COMPOSE_ARGS} up --wait --remove-orphans --build
  COMMAND ${CMAKE_COMMAND} -E echo "-- ========================================================================="
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- monetr is now running locally."
  COMMAND ${CMAKE_COMMAND} -E echo "-- You can access monetr via ${LOCAL_PROTOCOL}://my.${MONETR_LOCAL_DOMAIN}"
  COMMAND ${CMAKE_COMMAND} -E echo "-- Emails sent during development can be seen at ${LOCAL_PROTOCOL}://mail.${MONETR_LOCAL_DOMAIN}"
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- Optional Services:"
  COMMAND ${CMAKE_COMMAND} -E echo "--  ${LOCAL_PROTOCOL}://ngrok.${MONETR_LOCAL_DOMAIN} External: https://${NGROK_HOSTNAME}"
  COMMAND ${CMAKE_COMMAND} -E echo "--  ${LOCAL_PROTOCOL}://s3.${MONETR_LOCAL_DOMAIN} User: monetr Pass: password"
  COMMAND ${CMAKE_COMMAND} -E echo "--  ${LOCAL_PROTOCOL}://openbao.${MONETR_LOCAL_DOMAIN} Token: ${VAULT_ROOT_TOKEN}"
  COMMAND ${CMAKE_COMMAND} -E echo "--  ${LOCAL_PROTOCOL}://vault.${MONETR_LOCAL_DOMAIN} Token: ${VAULT_ROOT_TOKEN}"
  COMMAND ${CMAKE_COMMAND} -E echo "--"
  COMMAND ${CMAKE_COMMAND} -E echo "-- Optional services might not all be available and depend on your personal configuration."
  COMMAND ${CMAKE_COMMAND} -E echo "-- More information here: https://monetr.app/documentation/development/local_development/"
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
    download.simple-icons
    dependencies.node_modules
    build.email
    ${NGINX_CONFIG_FILE}
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
    COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${DEVELOPMENT_COMPOSE_ARGS} exec monetr monetr -c /build/compose/monetr.yaml development clean:plaid || ${CMAKE_COMMAND} -E true
    COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${DEVELOPMENT_COMPOSE_ARGS} exec monetr monetr -c /build/compose/monetr.yaml development clean:stripe || ${CMAKE_COMMAND} -E true
    COMMAND ${DOCKER_EXECUTABLE} --log-level ERROR compose ${ALL_COMPOSE_ARGS} down --remove-orphans -v || ${CMAKE_COMMAND} -E true
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}/development || ${CMAKE_COMMAND} -E true
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
