/**
 * Lighter WebSocket Example in C++
 *
 * This example demonstrates how to connect to the Lighter WebSocket API
 * and subscribe to real-time order book updates.
 *
 * Dependencies:
 *   - Boost.Beast (for WebSocket)
 *   - Boost.Asio (for networking)
 *   - OpenSSL (for TLS)
 *   - nlohmann/json (for JSON parsing)
 *
 * Build (macOS with Homebrew):
 *   brew install boost openssl nlohmann-json
 *   clang++ -std=c++17 -o websocket_example websocket_example.cpp \
 *     -I/opt/homebrew/include \
 *     -L/opt/homebrew/lib \
 *     -lssl -lcrypto -pthread
 *
 * Build (Linux):
 *   sudo apt-get install libboost-all-dev libssl-dev nlohmann-json3-dev
 *   g++ -std=c++17 -o websocket_example websocket_example.cpp \
 *     -lssl -lcrypto -pthread
 */

#include <boost/beast/core.hpp>
#include <boost/beast/ssl.hpp>
#include <boost/beast/websocket.hpp>
#include <boost/beast/websocket/ssl.hpp>
#include <boost/asio/connect.hpp>
#include <boost/asio/ip/tcp.hpp>
#include <boost/asio/ssl/stream.hpp>
#include <nlohmann/json.hpp>
#include <cstdlib>
#include <iostream>
#include <string>
#include <thread>
#include <atomic>
#include <csignal>

namespace beast = boost::beast;
namespace http = beast::http;
namespace websocket = beast::websocket;
namespace net = boost::asio;
namespace ssl = boost::asio::ssl;
using tcp = boost::asio::ip::tcp;
using json = nlohmann::json;

// Global flag for graceful shutdown
std::atomic<bool> g_running{true};

void signal_handler(int) {
    std::cout << "\nShutting down..." << std::endl;
    g_running = false;
}

class LighterWebSocket {
public:
    LighterWebSocket(const std::string& host, const std::string& port, const std::string& path)
        : host_(host), port_(port), path_(path),
          resolver_(ioc_), ctx_(ssl::context::tlsv12_client),
          ws_(ioc_, ctx_) {

        // Configure SSL context
        ctx_.set_default_verify_paths();
        ctx_.set_verify_mode(ssl::verify_peer);
    }

    bool connect() {
        try {
            // Resolve the host
            auto const results = resolver_.resolve(host_, port_);

            // Connect to the server
            auto ep = net::connect(get_lowest_layer(ws_), results);

            // Set SNI hostname
            if (!SSL_set_tlsext_host_name(ws_.next_layer().native_handle(), host_.c_str())) {
                throw beast::system_error(
                    beast::error_code(static_cast<int>(::ERR_get_error()),
                    net::error::get_ssl_category()),
                    "Failed to set SNI hostname");
            }

            // Perform SSL handshake
            ws_.next_layer().handshake(ssl::stream_base::client);

            // Set WebSocket options
            ws_.set_option(websocket::stream_base::decorator(
                [](websocket::request_type& req) {
                    req.set(http::field::user_agent, "lighter-cpp-client/1.0");
                }));

            // Perform WebSocket handshake
            std::string handshake_host = host_ + ":" + port_;
            ws_.handshake(handshake_host, path_);

            std::cout << "Connected to " << host_ << path_ << std::endl;
            return true;

        } catch (const std::exception& e) {
            std::cerr << "Connection error: " << e.what() << std::endl;
            return false;
        }
    }

    void subscribe_orderbook(int market_index) {
        json subscribe_msg = {
            {"type", "subscribe"},
            {"channel", "order_book/" + std::to_string(market_index)}
        };

        ws_.write(net::buffer(subscribe_msg.dump()));
        std::cout << "Subscribed to order_book/" << market_index << std::endl;
    }

    void subscribe_trades(int market_index) {
        json subscribe_msg = {
            {"type", "subscribe"},
            {"channel", "trade/" + std::to_string(market_index)}
        };

        ws_.write(net::buffer(subscribe_msg.dump()));
        std::cout << "Subscribed to trade/" << market_index << std::endl;
    }

    void subscribe_market_stats(int market_index) {
        json subscribe_msg = {
            {"type", "subscribe"},
            {"channel", "market_stats/" + std::to_string(market_index)}
        };

        ws_.write(net::buffer(subscribe_msg.dump()));
        std::cout << "Subscribed to market_stats/" << market_index << std::endl;
    }

    void run() {
        beast::flat_buffer buffer;

        while (g_running) {
            try {
                // Set a timeout for the read operation
                ws_.control_callback([](websocket::frame_type kind, beast::string_view payload) {
                    if (kind == websocket::frame_type::pong) {
                        // Handle pong if needed
                    }
                });

                buffer.clear();
                ws_.read(buffer);

                std::string msg = beast::buffers_to_string(buffer.data());
                handle_message(msg);

            } catch (const beast::system_error& se) {
                if (se.code() == websocket::error::closed) {
                    std::cout << "WebSocket closed" << std::endl;
                    break;
                }
                std::cerr << "Read error: " << se.what() << std::endl;
                break;
            }
        }
    }

    void close() {
        try {
            ws_.close(websocket::close_code::normal);
        } catch (...) {
            // Ignore close errors
        }
    }

private:
    void handle_message(const std::string& msg) {
        try {
            json j = json::parse(msg);

            std::string type = j.value("type", "");
            std::string channel = j.value("channel", "");

            if (type == "connected") {
                std::cout << "Received connected message" << std::endl;
                return;
            }

            if (type == "ping") {
                // Send pong response
                json pong = {{"type", "pong"}};
                ws_.write(net::buffer(pong.dump()));
                return;
            }

            if (type == "subscribed/order_book" || channel.find("order_book") == 0) {
                handle_orderbook(j);
                return;
            }

            if (type == "update/order_book") {
                handle_orderbook_update(j);
                return;
            }

            if (type == "subscribed/trade" || type == "update/trade") {
                handle_trade(j);
                return;
            }

            if (type == "subscribed/market_stats" || type == "update/market_stats") {
                handle_market_stats(j);
                return;
            }

            if (type == "error") {
                auto data = j.value("data", json::object());
                std::cerr << "Error: " << data.value("message", "Unknown error") << std::endl;
                return;
            }

            // Unknown message type - print for debugging
            if (!type.empty()) {
                std::cout << "Unknown message type: " << type << std::endl;
            }

        } catch (const json::exception& e) {
            std::cerr << "JSON parse error: " << e.what() << std::endl;
        }
    }

    void handle_orderbook(const json& j) {
        auto order_book = j.value("order_book", json::object());
        auto bids = order_book.value("bids", json::array());
        auto asks = order_book.value("asks", json::array());

        std::cout << "Order Book Snapshot: "
                  << bids.size() << " bids, "
                  << asks.size() << " asks" << std::endl;

        // Print best bid/ask
        if (!bids.empty() && !asks.empty()) {
            auto best_bid = bids[0];
            auto best_ask = asks[0];
            std::cout << "  Best Bid: " << best_bid.value("size", "0")
                      << " @ " << best_bid.value("price", "0")
                      << " | Best Ask: " << best_ask.value("size", "0")
                      << " @ " << best_ask.value("price", "0") << std::endl;
        }
    }

    void handle_orderbook_update(const json& j) {
        auto order_book = j.value("order_book", json::object());
        auto bids = order_book.value("bids", json::array());
        auto asks = order_book.value("asks", json::array());

        std::cout << "Order Book Update: "
                  << bids.size() << " bid updates, "
                  << asks.size() << " ask updates" << std::endl;
    }

    void handle_trade(const json& j) {
        auto data = j.value("data", json::object());

        if (data.is_array()) {
            for (const auto& trade : data) {
                std::cout << "Trade: "
                          << trade.value("size", "0") << " @ "
                          << trade.value("price", "0")
                          << " (" << trade.value("side", "unknown") << ")" << std::endl;
            }
        } else if (data.is_object()) {
            std::cout << "Trade: "
                      << data.value("size", "0") << " @ "
                      << data.value("price", "0")
                      << " (" << data.value("side", "unknown") << ")" << std::endl;
        }
    }

    void handle_market_stats(const json& j) {
        auto data = j.value("data", json::object());

        std::cout << "Market Stats:"
                  << " Last: " << data.value("last_price", "N/A")
                  << " Mark: " << data.value("mark_price", "N/A")
                  << " 24h Vol: " << data.value("volume_24h", "N/A")
                  << std::endl;
    }

    std::string host_;
    std::string port_;
    std::string path_;
    net::io_context ioc_;
    tcp::resolver resolver_;
    ssl::context ctx_;
    websocket::stream<beast::ssl_stream<tcp::socket>> ws_;
};

int main(int argc, char** argv) {
    // Set up signal handler for graceful shutdown
    std::signal(SIGINT, signal_handler);
    std::signal(SIGTERM, signal_handler);

    // Configuration
    std::string host = "mainnet.zklighter.elliot.ai";
    std::string port = "443";
    std::string path = "/stream";
    int market_index = 0;  // ETH-USD

    // Allow override from environment
    if (const char* env_host = std::getenv("LIGHTER_WS_HOST")) {
        host = env_host;
    }

    std::cout << "Lighter WebSocket C++ Example" << std::endl;
    std::cout << "Connecting to wss://" << host << path << std::endl;
    std::cout << "Press Ctrl+C to exit" << std::endl;
    std::cout << std::endl;

    try {
        LighterWebSocket ws(host, port, path);

        if (!ws.connect()) {
            return 1;
        }

        // Subscribe to order book for market 0 (ETH-USD)
        ws.subscribe_orderbook(market_index);

        // Run the message loop
        ws.run();

        ws.close();

    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }

    std::cout << "Disconnected" << std::endl;
    return 0;
}
