# Creates a new PostgreSQL migration file and opens it in $EDITOR.
# Usage: cmake -P cmake/scripts/NewPgMigration.cmake

set(migrationsDir "${CMAKE_SOURCE_DIR}/server/migrations/schema/pg")

# CMake has no way to read from stdin directly, so let a shell do the prompt.
execute_process(
  COMMAND sh -c "printf 'Migration name (no spaces, e.g. AddWidgets): ' >&2 && read -r name && printf '%s' \"$name\""
  OUTPUT_VARIABLE migrationName
  OUTPUT_STRIP_TRAILING_WHITESPACE
)

if(migrationName STREQUAL "")
  message(FATAL_ERROR "A migration name is required")
endif()
if(NOT migrationName MATCHES "^[A-Za-z0-9_]+$")
  message(FATAL_ERROR "Migration name \"${migrationName}\" may only contain letters, numbers and underscores")
endif()

string(TIMESTAMP today "%Y%m%d")

# Find the highest counter already used today and take the next one.
file(GLOB existingMigrations RELATIVE "${migrationsDir}" "${migrationsDir}/${today}[0-9][0-9]_*")
set(nextCounter 0)
foreach(existing IN LISTS existingMigrations)
  string(SUBSTRING "${existing}" 8 2 existingCounter)
  math(EXPR existingCounter "1${existingCounter} - 100")
  if(existingCounter GREATER_EQUAL nextCounter)
    math(EXPR nextCounter "${existingCounter} + 1")
  endif()
endforeach()
if(nextCounter GREATER 99)
  message(FATAL_ERROR "There are already 100 migrations for ${today}")
endif()

if(nextCounter LESS 10)
  set(nextCounter "0${nextCounter}")
endif()

set(migrationFile "${migrationsDir}/${today}${nextCounter}_${migrationName}.tx.up.sql")
file(TOUCH "${migrationFile}")
file(RELATIVE_PATH migrationFileRelative "${CMAKE_SOURCE_DIR}" "${migrationFile}")
message(STATUS "Created ${migrationFileRelative}")

if(NOT DEFINED ENV{EDITOR} OR "$ENV{EDITOR}" STREQUAL "")
  message(STATUS "EDITOR is not set, not opening the migration")
  return()
endif()

# execute_process pipes the child's stdout/stderr back through CMake, which
# breaks terminal editors, so hand the editor the real terminal instead. Going
# through sh also lets EDITOR include arguments, like "code --wait".
execute_process(
  COMMAND sh -c "exec $EDITOR \"$1\" </dev/tty >/dev/tty 2>/dev/tty" sh "${migrationFile}"
)
