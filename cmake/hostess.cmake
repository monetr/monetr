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
set(HOSTESS_linux_amd64   "v0.5.3" "96a7225c298f9546a25ae586693339e23a880fe55f990166bf5a0cb835eafebf")
set(HOSTESS_linux_arm64   "v0.5.3" "8b59014aa720ad7f8bf888333e0fa2616861b4abf18d8b73134853a5766e46ac")
set(HOSTESS_darwin_amd64  "v0.5.3" "947de867aed29bbc251b099acd2214fb87c3f87a103a7a29a843ab7e2df05652")
set(HOSTESS_darwin_arm64  "v0.5.3" "e6faefab87680b020f2d337f7c64a305ce495bed1a9b8c6c16f2c5628ed21f35")

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
