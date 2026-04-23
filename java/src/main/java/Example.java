import java.util.ArrayList;
import java.util.List;

/**
 * Mirrors examples/example.cpp — generates API keys, creates a client,
 * gets an auth token, then runs 5 threads each signing 100 create+cancel pairs.
 *
 * Build & run:
 *   javac -cp jna-5.14.0.jar:. LighterLib.java Example.java
 *   java --enable-native-access=ALL-UNNAMED -cp jna-5.14.0.jar:. Example
 */
public class Example {

    static final int  CHAIN_ID      = 304;
    static final long ACCOUNT_INDEX = 100L;
    static final int  MARKET_INDEX  = 0;       // ETH market
    static final long BASE_AMOUNT   = 10_000L;
    static final int  PRICE         = 400_000;
    static final int  ORDER_TYPE    = 0;       // limit
    static final int  TIME_IN_FORCE = 2;       // post-only
    static final int  N_THREADS     = 5;
    static final int  N_ORDERS      = 100;

    static long nowMs() { return System.currentTimeMillis(); }
    static long nowUs() { return System.nanoTime() / 1_000; }

    static void runExample(LighterLib.Lib lib, int apiKeyIndex) {
        // Generate a fresh API key pair
        LighterLib.ApiKeyResponse.ByValue apiResp = lib.GenerateAPIKey();
        if (apiResp.err != null) {
            System.err.println("[" + apiKeyIndex + "] GenerateAPIKey error: " + apiResp.err);
            return;
        }
        System.out.println("[" + apiKeyIndex + "] publicKey=" + apiResp.publicKey);

        // Create a client bound to the generated key
        String err = lib.CreateClient(null, apiResp.privateKey, CHAIN_ID, apiKeyIndex, ACCOUNT_INDEX);
        if (err != null) {
            System.err.println("[" + apiKeyIndex + "] CreateClient error: " + err);
            return;
        }

        // Auth token valid for 7 hours
        long tokenDeadline = nowMs() + 7L * 60 * 60 * 1000;
        LighterLib.StrOrErr.ByValue tokenResp = lib.CreateAuthToken(tokenDeadline, apiKeyIndex, ACCOUNT_INDEX);
        if (tokenResp.err != null) {
            System.err.println("[" + apiKeyIndex + "] CreateAuthToken error: " + tokenResp.err);
            return;
        }
        System.out.println("[" + apiKeyIndex + "] authToken=" + tokenResp.str);

        long nonce = 1L;
        long start = nowUs();

        for (int i = 1; i <= N_ORDERS; i++) {
            long orderExpiry = nowMs() + 60L * 60 * 1000; // 60 min from now

            // Sign a limit post-only ask order
            LighterLib.SignedTxResponse.ByValue create = lib.SignCreateOrder(
                MARKET_INDEX,
                (long) i,          // clientOrderIndex
                BASE_AMOUNT,
                PRICE,
                /* isAsk */        1,
                ORDER_TYPE,
                TIME_IN_FORCE,
                /* reduceOnly */   0,
                /* triggerPrice */ 0,
                orderExpiry,
                /* integratorAccountIndex */  0L,
                /* integratorTakerFee */      0,
                /* integratorMakerFee */      0,
                /* skipNonce */ (byte) 0,
                nonce,
                apiKeyIndex,
                ACCOUNT_INDEX
            );
            nonce++;

            if (create.err != null) {
                System.err.println("[" + apiKeyIndex + "] SignCreateOrder(" + i + ") error: " + create.err);
            }

            // Cancel the same order by client order index
            LighterLib.SignedTxResponse.ByValue cancel = lib.SignCancelOrder(
                MARKET_INDEX,
                (long) i,
                /* skipNonce */ (byte) 0,
                nonce,
                apiKeyIndex,
                ACCOUNT_INDEX
            );
            nonce++;

            if (cancel.err != null) {
                System.err.println("[" + apiKeyIndex + "] SignCancelOrder(" + i + ") error: " + cancel.err);
            }
        }

        long elapsed = nowUs() - start;
        System.out.printf("[%d] %d create+cancel pairs in %.2f ms%n",
            apiKeyIndex, N_ORDERS, elapsed / 1000.0);
    }

    public static void main(String[] args) throws InterruptedException {
        LighterLib.Lib lib = LighterLib.loadFromDir("../sharedlib");

        List<Thread> threads = new ArrayList<>();
        for (int i = 0; i < N_THREADS; i++) {
            final int apiKeyIndex = i;
            Thread t = new Thread(() -> runExample(lib, apiKeyIndex));
            t.start();
            threads.add(t);
        }

        for (Thread t : threads) {
            t.join();
        }
    }
}
