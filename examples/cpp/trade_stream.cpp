/**
 * Lighter Trade Stream Example in C++
 *
 * This example demonstrates how to stream real-time trade data
 * from the Lighter WebSocket API.
 *
 * Build (macOS with Homebrew):
 *   brew install boost openssl nlohmann-json
 *   clang++ -std=c++17 -o trade_stream trade_stream.cpp \
 *     -I/opt/homebrew/include \
 *     -L/opt/homebrew/lib \
 *     -lssl -lcrypto -pthread
 *
 * Build (Linux):
 *   sudo apt-get install libboost-all-dev libssl-dev nlohmann-json3-dev
 *   g++ -std=c++17 -o trade_stream trade_stream.cpp \
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
#include <atomic>
#include <csignal>
#include <iomanip>

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

class TradeStreamClient {
public:
    TradeStreamClient(const std::string& host, const std::string& port, const std::string& path)
        : host_(host), port_(port), path_(path),
          resolver_(ioc_), ctx_(ssl::context::tlsv12_client),
          ws_(ioc_, ctx_) {
        ctx_.set_default_verify_paths();
        ctx_.set_verify_mode(ssl::verify_peer);
    }

    bool connect() {
        try {
            auto const results = resolver_.resolve(host_, port_);
            auto ep = net::connect(get_lowest_layer(ws_), results);

            if (!SSL_set_tlsext_host_name(ws_.next_layer().native_handle(), host_.c_str())) {
                throw beast::system_error(
                    beast::error_code(static_cast<int>(::ERR_get_error()),
                    net::error::get_ssl_category()),
                    "Failed to set SNI hostname");
            }

            ws_.next_layer().handshake(ssl::stream_base::client);

            ws_.set_option(websocket::stream_base::decorator(
                [](websocket::request_type& req) {
                    req.set(http::field::user_agent, "lighter-cpp-trade-stream/1.0");
                }));

            std::string handshake_host = host_ + ":" + port_;
            ws_.handshake(handshake_host, path_);

            std::cout << "Connected to " << host_ << path_ << std::endl;
            return true;

        } catch (const std::exception& e) {
            std::cerr << "Connection error: " << e.what() << std::endl;
            return false;
        }
    }

    void subscribe_trades(int market_index) {
        json subscribe_msg = {
            {"type", "subscribe"},
            {"channel", "trade/" + std::to_string(market_index)}
        };

        ws_.write(net::buffer(subscribe_msg.dump()));
        std::cout << "Subscribed to trade/" << market_index << std::endl;
    }

    void run() {
        beast::flat_buffer buffer;

        while (g_running) {
            try {
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
                json pong = {{"type", "pong"}};
                ws_.write(net::buffer(pong.dump()));
                return;
            }

            if (type == "subscribed/trade" || channel.find("trade") == 0) {
                handle_trade_snapshot(j);
                return;
            }

            if (type == "update/trade") {
                handle_trade_update(j);
                return;
            }

            if (type == "error") {
                auto data = j.value("data", json::object());
                std::cerr << "Error: " << data.value("message", "Unknown error") << std::endl;
                return;
            }

        } catch (const json::exception& e) {
            std::cerr << "JSON parse error: " << e.what() << std::endl;
        }
    }

    void handle_trade_snapshot(const json& j) {
        auto data = j.value("data", json::object());

        if (data.is_array()) {
            std::cout << "Trade snapshot: " << data.size() << " recent trades" << std::endl;
            print_trade_header();
            for (const auto& trade : data) {
                print_trade(trade);
            }
        }
    }

    void handle_trade_update(const json& j) {
        auto data = j.value("data", json::object());

        if (data.is_array()) {
            for (const auto& trade : data) {
                print_trade(trade);
            }
        } else if (data.is_object()) {
            print_trade(data);
        }
    }

    void print_trade_header() {
        std::cout << std::setw(10) << "Side"
                  << std::setw(15) << "Price"
                  << std::setw(15) << "Size"
                  << std::endl;
        std::cout << std::string(40, '-') << std::endl;
    }

    void print_trade(const json& trade) {
        std::string side = trade.value("side", "unknown");
        std::string price = trade.value("price", "0");
        std::string size = trade.value("size", "0");

        // Color output: green for buy, red for sell
        if (side == "buy") {
            std::cout << "\033[32m"; // Green
        } else {
            std::cout << "\033[31m"; // Red
        }

        std::cout << std::setw(10) << side
                  << std::setw(15) << price
                  << std::setw(15) << size
                  << "\033[0m" // Reset color
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
    std::signal(SIGINT, signal_handler);
    std::signal(SIGTERM, signal_handler);

    std::string host = "mainnet.zklighter.elliot.ai";
    std::string port = "443";
    std::string path = "/stream";
    int market_index = 0;

    if (const char* env_host = std::getenv("LIGHTER_WS_HOST")) {
        host = env_host;
    }

    std::cout << "Lighter Trade Stream C++ Example" << std::endl;
    std::cout << "Connecting to wss://" << host << path << std::endl;
    std::cout << "Press Ctrl+C to exit" << std::endl;
    std::cout << std::endl;

    try {
        TradeStreamClient client(host, port, path);

        if (!client.connect()) {
            return 1;
        }

        // Subscribe to trades for market 0 (ETH-USD)
        client.subscribe_trades(market_index);

        std::cout << "Waiting for trades..." << std::endl << std::endl;

        client.run();
        client.close();

    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }

    std::cout << "Disconnected" << std::endl;
    return 0;
}
