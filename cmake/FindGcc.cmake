find_program(GCC_EXECUTABLE NAMES gcc)

if(NOT GCC_EXECUTABLE)
  message(WARNING "Could not find gcc compiler, monetr binary builds may fail as they require cgo.")
endif()
