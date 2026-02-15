cmake_minimum_required(VERSION 3.14)
project(lighter_cpp_examples VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find required packages
find_package(Boost REQUIRED)
find_package(OpenSSL REQUIRED)
find_package(Threads REQUIRED)

# Try to find nlohmann_json, or use FetchContent as fallback
find_package(nlohmann_json QUIET)
if(NOT nlohmann_json_FOUND)
    include(FetchContent)
    FetchContent_Declare(
        json
        URL https://github.com/nlohmann/json/releases/download/v3.11.3/json.tar.xz
    )
    FetchContent_MakeAvailable(json)
endif()

# WebSocket example
add_executable(websocket_example websocket_example.cpp)
target_link_libraries(websocket_example
    PRIVATE
        OpenSSL::SSL
        OpenSSL::Crypto
        Threads::Threads
        nlohmann_json::nlohmann_json
)

# Add include directories for Boost Beast (header-only)
target_include_directories(websocket_example PRIVATE ${Boost_INCLUDE_DIRS})

# Signer example (uses the shared library)
if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/../build/lighter-darwin-arm64.h")
    add_executable(signer_example example.cpp)
    target_include_directories(signer_example PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/../build")

    # Link against the shared library
    if(APPLE)
        target_link_libraries(signer_example PRIVATE
            "${CMAKE_CURRENT_SOURCE_DIR}/../build/lighter-darwin-arm64.dylib"
        )
    elseif(UNIX)
        target_link_libraries(signer_example PRIVATE
            "${CMAKE_CURRENT_SOURCE_DIR}/../build/lighter-linux-amd64.so"
        )
    endif()
endif()

# Installation
install(TARGETS websocket_example RUNTIME DESTINATION bin)
