###############################################################################
# hostess for local development. We used to `go install
# github.com/cbednarski/hostess`, but that pulls an unpinned latest and
# recompiles it from source on every fresh build tree. Instead we download a
# pre-built, version-pinned binary from monetr's fork
# (github.com/monetr/hostess) and verify it against a sha256 that renovate
# keeps current via the github-release-attachments datasource.
#
# This sets HOSTESS_EXECUTABLE and a `hostess` ExternalProject target that the
# hostsfile commands in development.cmake depend on.
###############################################################################

# Each line below is self contained: the release version plus the sha256 of
# that one architecture's binary. renovate figures out which asset a checksum
# belongs to by matching the existing sha against the current release, so the
# per-arch lines self select. The version is repeated on every line because the
# attachments datasource needs the version sitting right next to the sha it
# manages. If you build on an arch that is not pinned here, grab the checksum
# from https://github.com/monetr/hostess/releases and add a line for it.
include(ExternalProject)
set(HOSTESS_REPOSITORY "monetr/hostess")
set(HOSTESS_linux_amd64   "v0.6.0" "fffa9517b8c27ec7bd31e5b6956e0188e51f482cd353a3962ee3eda57bcb1999")
set(HOSTESS_linux_arm64   "v0.6.0" "ec5ac96b04d81e4d6bab20a435c3a5155f9224ad47420df39564b88b1a95c23a")
set(HOSTESS_darwin_amd64  "v0.6.0" "da2b55de8205bfea03da65022f80c35e5c9e252934e4bf5ad686c3cb9086c14a")
set(HOSTESS_darwin_arm64  "v0.6.0" "b356a6293a542122c59cd3e3c40b39270032f15287cf7afc9b613d0eed6dfb71")

# Figure out which pinned binary matches the host that cmake is running on. The
# release assets use Go's GOOS/GOARCH naming so we have to translate cmake's
# host names (x86_64, aarch64, Darwin, ...) into that.
string(TOLOWER "${CMAKE_HOST_SYSTEM_NAME}" HOSTESS_OS)
set(HOSTESS_ARCH "${CMAKE_HOST_SYSTEM_PROCESSOR}")
if(HOSTESS_ARCH STREQUAL "x86_64" OR HOSTESS_ARCH STREQUAL "AMD64")
  set(HOSTESS_ARCH "amd64")
elseif(HOSTESS_ARCH STREQUAL "aarch64" OR HOSTESS_ARCH STREQUAL "arm64")
  set(HOSTESS_ARCH "arm64")
endif()
set(HOSTESS_PLATFORM "${HOSTESS_OS}_${HOSTESS_ARCH}")
if(NOT DEFINED HOSTESS_${HOSTESS_PLATFORM})
  message(FATAL_ERROR
    "No pinned hostess binary for host platform '${HOSTESS_PLATFORM}'. Add a "
    "set(HOSTESS_${HOSTESS_PLATFORM} \"<version>\" \"<sha256>\") line in "
    "cmake/hostess.cmake using a checksum from "
    "https://github.com/${HOSTESS_REPOSITORY}/releases")
endif()
# Each pinned entry is a two element list: the version, then the sha256.
list(GET HOSTESS_${HOSTESS_PLATFORM} 0 HOSTESS_VERSION)
list(GET HOSTESS_${HOSTESS_PLATFORM} 1 HOSTESS_CHECKSUM)

set(HOSTESS_TOOLS_DIR ${CMAKE_BINARY_DIR}/tools/bin)
set(HOSTESS_EXECUTABLE ${HOSTESS_TOOLS_DIR}/hostess)
set(HOSTESS_ASSET "hostess-${HOSTESS_VERSION}-${HOSTESS_OS}-${HOSTESS_ARCH}")
set(HOSTESS_URL "https://github.com/${HOSTESS_REPOSITORY}/releases/download/${HOSTESS_VERSION}/${HOSTESS_ASSET}")

# The release assets are raw binaries, not archives, so DOWNLOAD_NO_EXTRACT
# keeps cmake from trying to untar them. URL_HASH does the actual verification
# for us, if the downloaded bytes do not match the pinned sha the configure
# fails. There is nothing to configure, build, or install, we just copy the
# verified binary into place and mark it executable (the release assets do not
# come with the executable bit set).
ExternalProject_Add(hostess
  URL ${HOSTESS_URL}
  URL_HASH SHA256=${HOSTESS_CHECKSUM}
  DOWNLOAD_NO_EXTRACT TRUE
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  TEST_COMMAND ""
  INSTALL_COMMAND
    ${CMAKE_COMMAND} -E copy <DOWNLOADED_FILE> ${HOSTESS_EXECUTABLE}
    COMMAND ${CMAKE_COMMAND} -DFILE=${HOSTESS_EXECUTABLE} -P ${CMAKE_SOURCE_DIR}/cmake/scripts/make_executable.cmake
  # Declaring the binary as a byproduct lets the hostsfile commands depend on
  # the file directly, and gets it cleaned up when the clean target runs.
  BUILD_BYPRODUCTS ${HOSTESS_EXECUTABLE}
)
