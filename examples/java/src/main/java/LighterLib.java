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

    private static String readAndFree(Lib lib, Pointer p) {
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

        String                   CreateClient(String url, String privateKey, int chainId,
                                              int apiKeyIndex, long accountIndex);

        String                   CheckClient(int apiKeyIndex, long accountIndex);

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
}