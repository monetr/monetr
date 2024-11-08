find_package(ImageMagick COMPONENTS convert)
if (ImageMagick_FOUNDD)
  message(STATUS "Found ImageMagick: ${IMAGEMAGICK_CONVERT_EXECUTABLE} (found version \"${ImageMagick_VERSION_STRING}\")")
else()
  message(AUTHOR_WARNING "ImageMagick was not detected, generating image assets is not available.")
endif()

set(BASE_LOGO ${CMAKE_CURRENT_SOURCE_DIR}/logo.png)

set(APPLE_SPLASH_SIZES 
  "1125x2436" 
  "1136x640" 
  "1170x2532" 
  "1242x2208"
  "1284x2778"
  "1334x750"
  "1536x2048"
  "1620x2160"
  "1668x2224"
  "1792x828"
  "2048x1536"
  "2048x2732"
  "2160x1620"
  "2208x1242"
  "2224x1668"
  "2388x1668"
  "2436x1125"
  "2532x1170"
  "2688x1242"
  "2732x2048"
  "2778x1284"
  "640x1136"
  "750x1334"
  "828x1792"
)
file(MAKE_DIRECTORY ${CMAKE_BINARY_DIR}/images/output)

set(APPLE_SPLASH_MARKER ${CMAKE_BINARY_DIR}/images/apple-splash-marker.txt)
add_custom_command(
  OUTPUT ${APPLE_SPLASH_MARKER}
  COMMENT "Finished generating Apple splash screens"
  COMMAND ${CMAKE_COMMAND} -E touch ${APPLE_SPLASH_MARKER}
)

add_custom_target(
  images.apple-splash-screen
  DEPENDS ${APPLE_SPLASH_MARKER}
)

foreach(APPLE_SIZE IN ITEMS ${APPLE_SPLASH_SIZES})
  # Extract the width and height from the current image size.
  string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\1" WIDTH "${APPLE_SIZE}")
  string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\2" HEIGHT "${APPLE_SIZE}")
  # Pad the inner icon to be at most half the size of the total image.
  math(EXPR LOGO_HEIGHT "${HEIGHT} / 3" OUTPUT_FORMAT DECIMAL)
  math(EXPR LOGO_WIDTH "${WIDTH} / 3" OUTPUT_FORMAT DECIMAL)
  # Generate that image.
  add_custom_command(
    OUTPUT ${CMAKE_BINARY_DIR}/images/output/apple-splash-${WIDTH}-${HEIGHT}.jpg
    BYPRODUCTS ${CMAKE_BINARY_DIR}/images/output/apple-splash-${WIDTH}-${HEIGHT}.jpg
    COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize ${LOGO_WIDTH}x${LOGO_HEIGHT} -background "\#19161f" -extent "${WIDTH}x${HEIGHT}" "${CMAKE_BINARY_DIR}/images/output/apple-splash-${WIDTH}-${HEIGHT}.jpg" > /dev/null 2>&1
    COMMENT "Generating ${WIDTH}x${HEIGHT} Apple splash image"
    DEPENDS ${BASE_LOGO}
    VERBATIM
  )
  add_custom_command(OUTPUT ${APPLE_SPLASH_MARKER} APPEND DEPENDS ${CMAKE_BINARY_DIR}/images/output/apple-splash-${WIDTH}-${HEIGHT}.jpg)
endforeach()

add_custom_target(
  images.apple-touch-icon
  BYPRODUCTS ${CMAKE_BINARY_DIR}/images/output/apple-touch-icon.jpg
  COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize 150x150 -background "\#19161f" -extent "180x180" "${CMAKE_BINARY_DIR}/images/output/apple-touch-icon.jpg" > /dev/null 2>&1
  VERBATIM
  COMMENT "Generating Apple touch icon"
)

set(FAVICON_SIZES 
  "16x16" 
  "32x32" 
  "64x64" 
)

set(FAVICON_MARKER ${CMAKE_BINARY_DIR}/images/favicon.txt)
add_custom_command(
  OUTPUT ${FAVICON_MARKER}
  COMMENT "Finished generating favicons"
  COMMAND ${CMAKE_COMMAND} -E touch ${FAVICON_MARKER}
)

add_custom_target(
  images.favicons
  DEPENDS ${FAVICON_MARKER}
)

foreach(FAVICON_SIZE IN ITEMS ${FAVICON_SIZES})
  # Extract the width and height from the current image size.
  string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\1" WIDTH "${FAVICON_SIZE}")
  string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\2" HEIGHT "${FAVICON_SIZE}")
  # Generate that image.
  add_custom_command(
    OUTPUT ${CMAKE_BINARY_DIR}/images/output/favicon-${WIDTH}x${HEIGHT}.png
    BYPRODUCTS ${CMAKE_BINARY_DIR}/images/output/favicon-${WIDTH}x${HEIGHT}.png
    COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize ${WIDTH}x${HEIGHT} -background transparent -extent "${WIDTH}x${HEIGHT}" "${CMAKE_BINARY_DIR}/images/output/favicon-${WIDTH}x${HEIGHT}.png" > /dev/null 2>&1
    COMMENT "Generating ${WIDTH}x${HEIGHT} favicon"
    DEPENDS ${BASE_LOGO}
    VERBATIM
  )
  add_custom_command(OUTPUT ${FAVICON_MARKER} APPEND DEPENDS ${CMAKE_BINARY_DIR}/images/output/favicon-${WIDTH}x${HEIGHT}.png)
endforeach()

add_custom_target(images DEPENDS images.apple-touch-icon images.apple-splash-screen images.favicons)

