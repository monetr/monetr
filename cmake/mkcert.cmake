###############################################################################
# mkcert for local development. We used to `go install filippo.io/mkcert`, but
# that pulls an unpinned latest and recompiles it from source on every fresh
# build tree. Instead we download a pre-built, version-pinned binary from
# monetr's fork (github.com/monetr/mkcert) and verify it against a sha256 that
# renovate keeps current via the github-release-attachments datasource.
#
# This sets MKCERT_EXECUTABLE and an `mkcert` ExternalProject target that the
# certificates command in development.cmake depends on.
###############################################################################

# Each line below is self contained: the release version plus the sha256 of
# that one architecture's binary. renovate figures out which asset a checksum
# belongs to by matching the existing sha against the current release, so the
# per-arch lines self select. The version is repeated on every line because the
# attachments datasource needs the version sitting right next to the sha it
# manages. If you build on an arch that is not pinned here, grab the checksum
# from https://github.com/monetr/mkcert/releases and add a line for it.
include(ExternalProject)
set(MKCERT_REPOSITORY "monetr/mkcert")
set(MKCERT_linux_amd64   "v1.4.6" "139bb7138c942b8d4c00252c8da02d155fcc44ff2862038d8a5887483c5b2a33")
set(MKCERT_linux_arm64   "v1.4.6" "263fe8e3713891be2be3076434c5cde5f1cac6de8e3059a93aad2d2e81a36405")
set(MKCERT_darwin_amd64  "v1.4.6" "77c042aa6be8fcc0f349ef0ff73020b2ad2666940cabfa458b8df2aa56ad1c97")
set(MKCERT_darwin_arm64  "v1.4.6" "011ae2b83da1c6a14c4ff5e06b7f0342d714f48b008d240a2aa2707c66c0d14f")

# Figure out which pinned binary matches the host that cmake is running on. The
# release assets use Go's GOOS/GOARCH naming so we have to translate cmake's
# host names (x86_64, aarch64, Darwin, ...) into that.
string(TOLOWER "${CMAKE_HOST_SYSTEM_NAME}" MKCERT_OS)
set(MKCERT_ARCH "${CMAKE_HOST_SYSTEM_PROCESSOR}")
if(MKCERT_ARCH STREQUAL "x86_64" OR MKCERT_ARCH STREQUAL "AMD64")
  set(MKCERT_ARCH "amd64")
elseif(MKCERT_ARCH STREQUAL "aarch64" OR MKCERT_ARCH STREQUAL "arm64")
  set(MKCERT_ARCH "arm64")
endif()
set(MKCERT_PLATFORM "${MKCERT_OS}_${MKCERT_ARCH}")
if(NOT DEFINED MKCERT_${MKCERT_PLATFORM})
  message(FATAL_ERROR
    "No pinned mkcert binary for host platform '${MKCERT_PLATFORM}'. Add a "
    "set(MKCERT_${MKCERT_PLATFORM} \"<version>\" \"<sha256>\") line in "
    "cmake/mkcert.cmake using a checksum from "
    "https://github.com/${MKCERT_REPOSITORY}/releases")
endif()
# Each pinned entry is a two element list: the version, then the sha256.
list(GET MKCERT_${MKCERT_PLATFORM} 0 MKCERT_VERSION)
list(GET MKCERT_${MKCERT_PLATFORM} 1 MKCERT_CHECKSUM)

set(MKCERT_TOOLS_DIR ${CMAKE_BINARY_DIR}/tools/bin)
set(MKCERT_EXECUTABLE ${MKCERT_TOOLS_DIR}/mkcert)
set(MKCERT_ASSET "mkcert-${MKCERT_VERSION}-${MKCERT_OS}-${MKCERT_ARCH}")
set(MKCERT_URL "https://github.com/${MKCERT_REPOSITORY}/releases/download/${MKCERT_VERSION}/${MKCERT_ASSET}")

# The release assets are raw binaries, not archives, so DOWNLOAD_NO_EXTRACT
# keeps cmake from trying to untar them. URL_HASH does the actual verification
# for us, if the downloaded bytes do not match the pinned sha the configure
# fails. There is nothing to configure, build, or install, we just copy the
# verified binary into place and mark it executable (the release assets do not
# come with the executable bit set).
ExternalProject_Add(mkcert
  URL ${MKCERT_URL}
  URL_HASH SHA256=${MKCERT_CHECKSUM}
  DOWNLOAD_NO_EXTRACT TRUE
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  TEST_COMMAND ""
  INSTALL_COMMAND
    ${CMAKE_COMMAND} -E copy <DOWNLOADED_FILE> ${MKCERT_EXECUTABLE}
    COMMAND ${CMAKE_COMMAND} -DFILE=${MKCERT_EXECUTABLE} -P ${CMAKE_SOURCE_DIR}/cmake/scripts/make_executable.cmake
  # Declaring the binary as a byproduct lets the certificates command depend on
  # the file directly, and gets it cleaned up when the clean target runs.
  BUILD_BYPRODUCTS ${MKCERT_EXECUTABLE}
)
