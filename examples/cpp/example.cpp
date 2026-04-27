#include <iostream>
#include <cstdint>
#include <thread>
#include <vector>
#include <chrono>
#if defined(__APPLE__)
  #include "../../build/lighter-signer-darwin-arm64.h"
#elif defined(__linux__)
  #include "../../build/lighter-signer-linux.h"
#elif defined(_WIN32)
  #include "../../build/lighter-signer-windows.h"
#endif
using namespace std;


uint64_t now_us() {
    using namespace std::chrono;
    return duration_cast<microseconds>(
        system_clock::now().time_since_epoch()
    ).count();
}

uint64_t now_ms() {
    using namespace std::chrono;
    return duration_cast<milliseconds>(
        system_clock::now().time_since_epoch()
    ).count();
}

void run_example(int apiKeyIndex) {
    // Example: generate an API key
    ApiKeyResponse apiResp = GenerateAPIKey();

    if (apiResp.err != nullptr) {
        Free(apiResp.err);
        return;
    }

    auto clientErr = CreateClient(nullptr, apiResp.privateKey, 304, apiKeyIndex, 100);
    Free(apiResp.privateKey);
    Free(apiResp.publicKey);
    if (clientErr != nullptr) {
        cerr << "CreateClient" << '\t' << clientErr << '\n';
        Free(clientErr);
        return;
    }

    long long accountIndex = 100;

    // create an auth token with expiry 7 hours in the future
    StrOrErr tokenResp = CreateAuthToken(0 , apiKeyIndex, accountIndex);
    if (tokenResp.err != nullptr) {
        Free(tokenResp.err);
        return;
    }
    Free(tokenResp.str);

    long long nonce = 1;

    auto start = now_us();
    for (int i = 1; i <= 100; i += 1) {
        // create an order to sell 1 ETH @ 4000 w/ a deadline 60 mins in the future
        // limit post only order
        auto create = SignCreateOrder(
            0, i, 10000, 400000, true,
            /* cOrderType */ 0, /* cTimeInForce */ 2, /* cReduceOnly */ 0, /* cTriggerPrice */ 0,
            now_ms() + 60 * 60 * 1000,
            /* cIntegratorAccountIndex */ 0,
            /* cIntegratorTakerFee */ 0,
            /* cIntegratorMakerFee */ 0,
            /* cSkipNonce */ 0,
            nonce, apiKeyIndex, accountIndex);
        nonce += 1;

        if (create.err != nullptr) {
            cerr << "create" << '\t' << create.err << '\n';
            Free(create.err);
        }
        if (create.txInfo != nullptr) Free(create.txInfo);
        if (create.txHash != nullptr) Free(create.txHash);
        if (create.messageToSign != nullptr) Free(create.messageToSign);

        // cancel order with client order id i on ETH market (market ID 0)
        auto cancel = SignCancelOrder(0, i, /* cSkipNonce */ 0, nonce, apiKeyIndex, accountIndex);
        nonce += 1;

        if (cancel.err != nullptr) {
            cerr << "cancel" << '\t' << cancel.err << '\n';
            Free(cancel.err);
        }
        if (cancel.txInfo != nullptr) Free(cancel.txInfo);
        if (cancel.txHash != nullptr) Free(cancel.txHash);
        if (cancel.messageToSign != nullptr) Free(cancel.messageToSign);
    }
    auto end = now_us();
    cout << "elapsed" << '\t' << float(end - start) / 1000 << "ms" << '\n';
}

int main() {
    vector<thread> runners;
    for (int i = 0; i < 5; i += 1) {
        runners.emplace_back(run_example, i);
    }

    for (auto& t: runners) {
        t.join();
    }

    return 0;
}
