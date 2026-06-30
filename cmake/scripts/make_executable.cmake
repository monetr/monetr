# Tiny helper invoked via `cmake -P` to mark a file as executable. cmake -E has
# no chmod equivalent, so when we download a raw binary it lands without the
# executable bit set and we need this to flip it. Pass the file in via
# `-DFILE=<path>`. file(CHMOD) needs cmake 3.19+ which we are well past.
if(NOT DEFINED FILE)
  message(FATAL_ERROR "FILE must be set when running make_executable.cmake, e.g. -DFILE=<path>")
endif()

file(CHMOD ${FILE}
  PERMISSIONS
    OWNER_READ OWNER_WRITE OWNER_EXECUTE
    GROUP_READ GROUP_EXECUTE
    WORLD_READ WORLD_EXECUTE
)
