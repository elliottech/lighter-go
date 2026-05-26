import com.sun.jna.Library;
import com.sun.jna.Native;
import com.sun.jna.Pointer;
import com.sun.jna.Structure;
import com.sun.jna.Structure.FieldOrder;

import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;

/**
 * JNA bindings for the lighter-go shared library.
 *
 * Build the .dylib first:
 *   go build -buildmode=c-shared -o lighter.dylib .   (macOS)
 *   go build -buildmode=c-shared -o lighter.so .      (Linux)
 *
 * Load:
 *   Lib lib = LighterLib.load("/abs/path/to/lighter.dylib");
 *   Lib lib = LighterLib.loadFromDir("../sharedlib");
 */
public class LighterLib {

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    @FieldOrder({"str", "err"})
    public static class StrOrErr extends Structure {
        public Pointer str;
        public Pointer err;

        public static class ByValue extends StrOrErr implements Structure.ByValue {}

        public String unwrap(Lib lib) {
            String errStr = readAndFree(lib, err);
            String strStr = readAndFree(lib, str);
            if (errStr != null) throw new RuntimeException(errStr);
            return strStr;
        }
    }

    @FieldOrder({"privateKey", "publicKey", "err"})
    public static class ApiKeyResponse extends Structure {
        public Pointer privateKey;
        public Pointer publicKey;
        public Pointer err;

        public static class ByValue extends ApiKeyResponse implements Structure.ByValue {}

        /** Read all string fields and free the native pointers. */
        public String[] readAndFree(Lib lib) {
            String pk  = LighterLib.readAndFree(lib, privateKey);
            String pub_ = LighterLib.readAndFree(lib, publicKey);
            String e   = LighterLib.readAndFree(lib, err);
            if (e != null) throw new RuntimeException(e);
            return new String[]{pk, pub_};
        }
    }

    // uint8_t txType sits at offset 0; the next field is a pointer which requires
    // 8-byte alignment on 64-bit platforms, so 7 bytes of padding follow txType.
    @FieldOrder({"txType", "_pad", "txInfo", "txHash", "messageToSign", "err"})
    public static class SignedTxResponse extends Structure {
        public byte    txType;
        public byte[]  _pad = new byte[7];
        public Pointer txInfo;
        public Pointer txHash;
        public Pointer messageToSign;
        public Pointer err;

        public static class ByValue extends SignedTxResponse implements Structure.ByValue {}

        /** Read all string fields and free the native pointers. Returns {txInfo, txHash, messageToSign}. */
        public String[] readAndFree(Lib lib) {
            String info = LighterLib.readAndFree(lib, txInfo);
            String hash = LighterLib.readAndFree(lib, txHash);
            String msg  = LighterLib.readAndFree(lib, messageToSign);
            String e    = LighterLib.readAndFree(lib, err);
            if (e != null) throw new RuntimeException(e);
            return new String[]{info, hash, msg};
        }
    }

    @FieldOrder({"MarketIndex", "ClientOrderIndex", "BaseAmount", "Price",
                 "IsAsk", "Type", "TimeInForce", "ReduceOnly", "TriggerPrice", "OrderExpiry"})
    public static class CreateOrderTxReq extends Structure {
        public short MarketIndex;
        public long  ClientOrderIndex;
        public long  BaseAmount;
        public int   Price;
        public byte  IsAsk;
        public byte  Type;
        public byte  TimeInForce;
        public byte  ReduceOnly;
        public int   TriggerPrice;
        public long  OrderExpiry;

        public static CreateOrderTxReq[] allocateArray(int size) {
            return (CreateOrderTxReq[]) new CreateOrderTxReq().toArray(size);
        }
    }

    // -------------------------------------------------------------------------
    // Helper — read a C string from a Pointer and free the native memory
    // -------------------------------------------------------------------------

    public static String readAndFree(Lib lib, Pointer p) {
        if (p == null) return null;
        String s = p.getString(0);
        lib.Free(p);
        return s;
    }

    // -------------------------------------------------------------------------
    // JNA interface — all struct return values are ByValue
    // -------------------------------------------------------------------------

    public interface Lib extends Library {
        ApiKeyResponse.ByValue   GenerateAPIKey();

        Pointer                  CreateClient(String url, String privateKey, int chainId,
                                              int apiKeyIndex, long accountIndex);

        Pointer                  CheckClient(int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignChangePubKey(String pubKey, byte skipNonce, long nonce,
                                                 int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignCreateOrder(
                int marketIndex, long clientOrderIndex, long baseAmount,
                int price, int isAsk, int orderType, int timeInForce,
                int reduceOnly, int triggerPrice, long orderExpiry,
                long integratorAccountIndex, int integratorTakerFee, int integratorMakerFee,
                byte skipNonce, long nonce, int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignCreateGroupedOrders(
                byte groupingType, CreateOrderTxReq orders, int len,
                long integratorAccountIndex, int integratorTakerFee, int integratorMakerFee,
                byte skipNonce, long nonce, int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignCancelOrder(int marketIndex, long orderIndex,
                                                byte skipNonce, long nonce,
                                                int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignWithdraw(int assetIndex, int routeType, long amount,
                                             byte skipNonce, long nonce,
                                             int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignCreateSubAccount(byte skipNonce, long nonce,
                                                      int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignCancelAllOrders(int timeInForce, long time,
                                                     byte skipNonce, long nonce,
                                                     int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignModifyOrder(
                int marketIndex, long index, long baseAmount, long price, long triggerPrice,
                long integratorAccountIndex, int integratorTakerFee, int integratorMakerFee,
                byte skipNonce, long nonce, int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignTransfer(
                long toAccountIndex, short assetIndex, byte fromRouteType, byte toRouteType,
                long amount, long usdcFee, String memo,
                byte skipNonce, long nonce, int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignCreatePublicPool(long operatorFee, int initialTotalShares,
                                                      long minOperatorShareRate,
                                                      byte skipNonce, long nonce,
                                                      int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignUpdatePublicPool(long publicPoolIndex, int status,
                                                      long operatorFee, int minOperatorShareRate,
                                                      byte skipNonce, long nonce,
                                                      int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignMintShares(long publicPoolIndex, long shareAmount,
                                               byte skipNonce, long nonce,
                                               int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignBurnShares(long publicPoolIndex, long shareAmount,
                                               byte skipNonce, long nonce,
                                               int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignUpdateLeverage(int marketIndex, int initialMarginFraction,
                                                    int marginMode,
                                                    byte skipNonce, long nonce,
                                                    int apiKeyIndex, long accountIndex);

        StrOrErr.ByValue         CreateAuthToken(long deadline, int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignUpdateMargin(int marketIndex, long usdcAmount, int direction,
                                                 byte skipNonce, long nonce,
                                                 int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignStakeAssets(long stakingPoolIndex, long shareAmount,
                                                byte skipNonce, long nonce,
                                                int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignUnstakeAssets(long stakingPoolIndex, long shareAmount,
                                                  byte skipNonce, long nonce,
                                                  int apiKeyIndex, long accountIndex);

        SignedTxResponse.ByValue SignApproveIntegrator(
                long integratorIndex,
                int maxPerpsTakerFee, int maxPerpsMakerFee,
                int maxSpotTakerFee, int maxSpotMakerFee,
                long approvalExpiry,
                byte skipNonce, long nonce, int apiKeyIndex, long accountIndex);

        void Free(Pointer ptr);
    }

    // -------------------------------------------------------------------------
    // Loader helpers
    // -------------------------------------------------------------------------

    public static Lib load(String absolutePath) {
        return Native.load(absolutePath, Lib.class);
    }

    public static Lib loadFromDir(String relativeDir) {
        String ext = System.getProperty("os.name").toLowerCase().contains("mac") ? "dylib" : "so";
        Path lib = Paths.get(System.getProperty("user.dir"))
                        .resolve(relativeDir)
                        .resolve("lighter." + ext)
                        .toAbsolutePath();
        return load(lib.toString());
    }

    // -------------------------------------------------------------------------
    // BugDemo — demonstrates why every char* MUST be bound as Pointer.
    //
    // Run from examples/java/ after building the .so:
    //   mvn -B exec:java -Dexec.mainClass='LighterLib$BugDemo' -Dexec.args=fixed
    //   MALLOC_CHECK_=3 mvn -B exec:java \
    //     -Dexec.mainClass='LighterLib$BugDemo' -Dexec.args=broken
    // On macOS substitute  MallocScribble=1 MallocGuardEdges=1  for MALLOC_CHECK_=3.
    // -------------------------------------------------------------------------
    public static class BugDemo {

        private static final int ITERS = 200_000;

        /**
         * The broken pattern. Every char* return / struct field is declared as
         * {@code String} instead of {@code Pointer}, and {@code Free} takes a
         * {@code String}. Both are wrong:
         *
         *  - Struct field as String: JNA copies the bytes from the malloc'd
         *    C buffer into a fresh JVM String and discards the original
         *    pointer. The native buffer leaks forever.
         *  - Free(String): JNA must convert the Java String into a void* to
         *    satisfy the C signature, so it allocates a fresh com.sun.jna.Memory
         *    buffer, copies the bytes in, and passes that Memory's address to
         *    Go's Free. Go calls C.free on it; JNA's Memory finalizer later
         *    tries to free the same address again → double-free.
         */
        public interface BrokenLib extends Library {
            @FieldOrder({"privateKey", "publicKey", "err"})
            class ApiKeyResponse extends Structure {
                public String privateKey;  // WRONG — should be Pointer
                public String publicKey;   // WRONG
                public String err;         // WRONG
                public static class ByValue extends ApiKeyResponse implements Structure.ByValue {}
            }
            ApiKeyResponse.ByValue GenerateAPIKey();
            void Free(String ptr);         // WRONG — should be void Free(Pointer)
        }

        public static void main(String[] args) {
            String mode = args.length > 0 ? args[0] : "fixed";
            String soDir = "../../sharedlib";
            switch (mode) {
                case "fixed":  runFixed(soDir); break;
                case "broken": runBroken(soDir); break;
                default:
                    System.err.println("Usage: BugDemo (fixed|broken)");
                    System.exit(2);
            }
        }

        /** The correct pattern: Pointer fields + readAndFree. */
        public static void runFixed(String soDir) {
            Lib lib = LighterLib.loadFromDir(soDir);
            System.out.println("[fixed] running " + ITERS + " GenerateAPIKey iterations");
            long t0 = System.nanoTime();
            for (int i = 0; i < ITERS; i++) {
                ApiKeyResponse.ByValue r = lib.GenerateAPIKey();
                String[] keys = r.readAndFree(lib);  // reads + frees the original Pointers
                if (i == 0) {
                    System.out.println("[fixed]   privateKey[0..16]=" + keys[0].substring(0, Math.min(18, keys[0].length())));
                    System.out.println("[fixed]   publicKey [0..16]=" + keys[1].substring(0, Math.min(18, keys[1].length())));
                }
            }
            long ms = (System.nanoTime() - t0) / 1_000_000;
            System.out.printf("[fixed] done — %d iters in %d ms, no leak, no double-free%n", ITERS, ms);
        }

        /** Demonstrates the broken pattern. Expected: leak then a hard crash. */
        public static void runBroken(String soDir) {
            String ext = System.getProperty("os.name").toLowerCase().contains("mac") ? "dylib" : "so";
            Path libPath = Paths.get(System.getProperty("user.dir")).resolve(soDir).resolve("lighter." + ext).toAbsolutePath();
            BrokenLib lib = Native.load(libPath.toString(), BrokenLib.class);
            System.out.println("[broken] loading lib with String-typed bindings (this is the bug pattern)");
            System.out.println("[broken] expect a leak per call + a double-free crash within a few iters");
            for (int i = 0; i < ITERS; i++) {
                BrokenLib.ApiKeyResponse.ByValue r = lib.GenerateAPIKey();
                // The Java Strings exist, but the original C heap addresses are
                // GONE — JNA already discarded them when marshaling char* -> String.
                // Trying to "Free" them is undefined behavior.
                lib.Free(r.privateKey);
                lib.Free(r.publicKey);
                if (i % 50 == 0) System.out.println("[broken]   iter " + i + " (still alive — but leaking + corrupting heap)");
            }
            System.out.println("[broken] survived all iters (unlikely — usually crashes earlier)");
        }
    }
}
