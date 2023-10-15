if(DOCKER_EXECUTABLE)
  message(STATUS "Documentation can be built and run locally via Docker:")
  set(DOCUMENTATION_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/documentation")
  file(MAKE_DIRECTORY "${DOCUMENTATION_OUTPUT_DIRECTORY}")

  file(GLOB DOCUMENTATION_SRC "${CMAKE_SOURCE_DIR}/docs/*")

  if(NOT "$ENV{MKDOCS_IMAGE}" STREQUAL "")
    set(MKDOCS_IMAGE "$ENV{MKDOCS_IMAGE}")
  else()
    set(MKDOCS_IMAGE squidfunk/mkdocs-material:9.4.4)
  endif()

  set(SITE_URL "https://monetr.app/")
  set(DOCUMENTATION_DIST
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/404.html"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/assets/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/contact/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/custom/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/developing/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/help/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/img/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/index.html"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/screenshots/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/search/"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/sitemap.xml"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/sitemap.xml.gz"
    "${DOCUMENTATION_OUTPUT_DIRECTORY}/stylesheets/"
  )
  add_custom_command(
    OUTPUT ${DOCUMENTATION_DIST}
    BYPRODUCTS ${DOCUMENTATION_DIST}
    COMMAND ${DOCKER_EXECUTABLE} run -v ${CMAKE_SOURCE_DIR}:${CMAKE_SOURCE_DIR} -e SITE_URL=${SITE_URL} -w ${CMAKE_SOURCE_DIR} --rm ${MKDOCS_IMAGE} build
    COMMENT "Building monetr.app documentation site using Docker: ${MKDOCS_IMAGE}"
    WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
    DEPENDS
      ${DOCUMENTATION_SRC}
  )

  add_custom_target(
    build.docs
    DEPENDS ${DOCUMENTATION_DIST}
    COMMENT "Documentation has been built at: ${DOCUMENTATION_OUTPUT_DIRECTORY}"
  )
endif()
