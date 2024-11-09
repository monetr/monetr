set(OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/images/output)
file(MAKE_DIRECTORY ${OUTPUT_DIRECTORY})

find_package(ImageMagick COMPONENTS convert)
if (ImageMagick_FOUND)
  set(BASE_LOGO ${CMAKE_CURRENT_SOURCE_DIR}/logo.png)

  # Taken from https://developer.apple.com/design/human-interface-guidelines/layout/#Specifications
  set(APPLE_SPLASH_SIZES
    "2048x2732"
    "2732x2048"
    "1668x2388"
    "2388x1668"
    "1536x2048"
    "2048x1536"
    "1488x2266"
    "2266x1488"
    "1640x2360"
    "2360x1640"
    "1668x2224"
    "2224x1668"
    "1620x2160"
    "2160x1620"
    "1320x2868"
    "2868x1320"
    "1206x2622"
    "2622x1206"
    "1290x2796"
    "2796x1290"
    "1179x2556"
    "2556x1179"
    "1284x2778"
    "2778x1284"
    "1170x2532"
    "2532x1170"
    "1125x2436"
    "2436x1125"
    "1242x2688"
    "2688x1242"
    "828x1792"
    "1792x828"
    "1242x2208"
    "2208x1242"
    "750x1334"
    "1334x750"
    "640x1136"
    "1136x640"
  )
  set(APPLE_IMAGES)
  foreach(APPLE_SIZE IN ITEMS ${APPLE_SPLASH_SIZES})
    # Extract the width and height from the current image size.
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\1" WIDTH "${APPLE_SIZE}")
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\2" HEIGHT "${APPLE_SIZE}")
    # Pad the inner icon to be at most half the size of the total image.
    math(EXPR LOGO_HEIGHT "${HEIGHT} / 3" OUTPUT_FORMAT DECIMAL)
    math(EXPR LOGO_WIDTH "${WIDTH} / 3" OUTPUT_FORMAT DECIMAL)
    # Generate that image.
    set(OUTPUT_IMAGE ${OUTPUT_DIRECTORY}/apple-splash-${WIDTH}-${HEIGHT}.jpg)
    add_custom_command(
      OUTPUT ${OUTPUT_IMAGE}
      COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize ${LOGO_WIDTH}x${LOGO_HEIGHT} -background "\#19161f" -extent "${WIDTH}x${HEIGHT}" "${OUTPUT_IMAGE}" > /dev/null 2>&1
      COMMENT "Generating ${WIDTH}x${HEIGHT} Apple splash screen image"
      DEPENDS ${BASE_LOGO}
      VERBATIM
    )
    list(APPEND APPLE_IMAGES ${OUTPUT_IMAGE})
  endforeach()
  add_custom_target(
    images.apple-splash-screen
    COMMENT "Generated all Apple splash screen images"
    DEPENDS ${APPLE_IMAGES}
  )

  add_custom_target(
    images.apple-touch-icon
    COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize 150x150 -background "\#19161f" -extent "180x180" "${OUTPUT_DIRECTORY}/apple-touch-icon.jpg" > /dev/null 2>&1
    VERBATIM
    COMMENT "Generating Apple touch icon"
  )

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
    COMMENT "Generating favicon from sub-icons"
    DEPENDS ${FAVICON_IMAGES}
    COMMAND_EXPAND_LISTS
    VERBATIM
  )

  set(MSTILE_SIZES 
    "128x128" 
    "270x270" 
    "558x270" 
    "558x558"
  )
  set(MSTILE_IMAGES)
  foreach(MSTILE_SIZE IN ITEMS ${MSTILE_SIZES})
    # Extract the width and height from the current image size.
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\1" WIDTH "${MSTILE_SIZE}")
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\2" HEIGHT "${MSTILE_SIZE}")
    # 90% aproximation, this will make the icon itself ~90% of the size of the total icon.
    math(EXPR LOGO_HEIGHT "(${HEIGHT} / 10) * 9" OUTPUT_FORMAT DECIMAL)
    math(EXPR LOGO_WIDTH "(${WIDTH} / 10) * 9" OUTPUT_FORMAT DECIMAL)
    # Generate that image.
    set(IMAGE_OUTPUT ${OUTPUT_DIRECTORY}/mstile-icon-${WIDTH}-${HEIGHT}.png)
    add_custom_command(
      OUTPUT ${IMAGE_OUTPUT}
      COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize ${LOGO_WIDTH}x${LOGO_HEIGHT} -background "\#19161f" -extent "${WIDTH}x${HEIGHT}" "${IMAGE_OUTPUT}" > /dev/null 2>&1
      COMMENT "Generating ${WIDTH}x${HEIGHT} Windows static tile icon"
      DEPENDS ${BASE_LOGO}
      VERBATIM
    )
    list(APPEND MSTILE_IMAGES ${IMAGE_OUTPUT})
  endforeach()

  add_custom_target(
    images.mstile
    COMMENT "Generated Windows static tile icons"
    DEPENDS ${MSTILE_IMAGES}
  )

  set(ANDROID_CHROME_SIZES
    "192x192"
    "512x512"
  )
  set(ANDROID_CHROME_IMAGES)
  foreach(ANDROID_CHROME_SIZE IN ITEMS ${ANDROID_CHROME_SIZES})
    # Extract the width and height from the current image size.
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\1" WIDTH "${ANDROID_CHROME_SIZE}")
    string(REGEX REPLACE "([0-9]+)[x]([0-9]+)" "\\2" HEIGHT "${ANDROID_CHROME_SIZE}")
    # 90% aproximation, this will make the icon itself ~90% of the size of the total icon.
    math(EXPR LOGO_HEIGHT "(${HEIGHT} / 10) * 9" OUTPUT_FORMAT DECIMAL)
    math(EXPR LOGO_WIDTH "(${WIDTH} / 10) * 9" OUTPUT_FORMAT DECIMAL)

    # Generate and image that is maskable, so it has a solid background.
    set(MASKABLE_OUTPUT ${OUTPUT_DIRECTORY}/android-chrome-${WIDTH}-${HEIGHT}_maskable.png)
    add_custom_command(
      OUTPUT ${MASKABLE_OUTPUT}
      COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize ${LOGO_WIDTH}x${LOGO_HEIGHT} -background "\#19161f" -extent "${WIDTH}x${HEIGHT}" "${MASKABLE_OUTPUT}" > /dev/null 2>&1
      COMMENT "Generating ${WIDTH}x${HEIGHT} Android Chrome maskable icon"
      DEPENDS ${BASE_LOGO}
      VERBATIM
    )
    list(APPEND ANDROID_CHROME_IMAGES ${MASKABLE_OUTPUT})

    # Generate the transparent image
    set(TRANSPARENT_OUTPUT ${OUTPUT_DIRECTORY}/android-chrome-${WIDTH}-${HEIGHT}.png)
    add_custom_command(
      OUTPUT ${TRANSPARENT_OUTPUT}
      COMMAND ${IMAGEMAGICK_CONVERT_EXECUTABLE} ${BASE_LOGO} -gravity center -resize ${LOGO_WIDTH}x${LOGO_HEIGHT} -background transparent -extent "${WIDTH}x${HEIGHT}" "${TRANSPARENT_OUTPUT}" > /dev/null 2>&1
      COMMENT "Generating ${WIDTH}x${HEIGHT} Android Chrome maskable icon"
      DEPENDS ${BASE_LOGO}
      VERBATIM
    )
    list(APPEND ANDROID_CHROME_IMAGES ${TRANSPARENT_OUTPUT})
  endforeach()
  add_custom_target(
    images.android-chrome
    COMMENT "Generated Android/Chrome icons"
    DEPENDS ${ANDROID_CHROME_IMAGES}
  )

  add_custom_target(images DEPENDS images.apple-touch-icon images.apple-splash-screen images.favicons images.mstile images.android-chrome)
else()
  message(AUTHOR_WARNING "ImageMagick was not detected, generating image assets is not available.")
  add_custom_target(
    images 
    BYPRODUCTS ${CMAKE_BINARY_DIR}/bogus-marker.txt
    COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --yellow "ImageMagick is not detected, unable to regenerate image assets"
    COMMENT "Install ImageMagick in order to generate image assets"
  )
endif()

