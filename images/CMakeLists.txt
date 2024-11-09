set(OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/images/output)
file(MAKE_DIRECTORY ${OUTPUT_DIRECTORY})

find_package(ImageMagick COMPONENTS convert)
if (ImageMagick_FOUND)
  set(BASE_LOGO ${CMAKE_CURRENT_SOURCE_DIR}/logo.png)
  set(FAVICON_SIZES 
    "16x16" 
    "24x24"
    "32x32" 
    "64x64" 
  )
  set(FAVICON_IMAGES)
  foreach(FAVICON_SIZE IN ITEMS ${FAVICON_SIZES})
    # Extract the width and height from the current image size.
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\1" WIDTH "${FAVICON_SIZE}")
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\2" HEIGHT "${FAVICON_SIZE}")
    # Generate that image.
    add_custom_command(
      OUTPUT ${OUTPUT_DIRECTORY}/favicon-${WIDTH}x${HEIGHT}.png
      COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize ${WIDTH}x${HEIGHT} -background transparent -extent "${WIDTH}x${HEIGHT}" "${OUTPUT_DIRECTORY}/favicon-${WIDTH}x${HEIGHT}.png" > /dev/null 2>&1
      COMMENT "Generating ${WIDTH}x${HEIGHT} favicon"
      DEPENDS ${BASE_LOGO}
      VERBATIM
    )
    list(APPEND FAVICON_IMAGES ${OUTPUT_DIRECTORY}/favicon-${WIDTH}x${HEIGHT}.png)
  endforeach()
  add_custom_target(
    images.favicons
    COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${FAVICON_IMAGES} ${OUTPUT_DIRECTORY}/favicon.ico > /dev/null 2>&1
    COMMENT "Generating favicon from sub-icons: ${OUTPUT_DIRECTORY}/favicon.ico"
    DEPENDS ${FAVICON_IMAGES}
    COMMAND_EXPAND_LISTS
    VERBATIM
  )

  add_custom_target(images DEPENDS images.favicons)
else()
  add_custom_target(
    images 
    BYPRODUCTS ${CMAKE_BINARY_DIR}/bogus-marker.txt
    COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --yellow "ImageMagick is not detected, unable to regenerate image assets"
    COMMENT "Install ImageMagick in order to generate image assets"
  )
endif()

