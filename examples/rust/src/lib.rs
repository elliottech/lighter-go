use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use libloading::{Library, Symbol};

// -------------------------------------------------------------------------
// Raw C structs — repr(C) inserts the same padding the C/Go ABI uses
// -------------------------------------------------------------------------

/// Mirrors `StrOrErr` from lighter.h
#[repr(C)]
pub struct RawStrOrErr {
    pub str_: *mut c_char,
    pub err: *mut c_char,
}

/// Mirrors `ApiKeyResponse` from lighter.h
#[repr(C)]
pub struct RawApiKeyResponse {
    pub private_key: *mut c_char,
    pub public_key: *mut c_char,
    pub err: *mut c_char,
}

/// Mirrors `SignedTxResponse` from lighter.h.
///
/// txType (1 byte) is followed by 7 bytes of implicit padding on 64-bit
/// before the first pointer — repr(C) handles this automatically.
#[repr(C)]
pub struct RawSignedTxResponse {
    pub tx_type: u8,
    pub tx_info: *mut c_char,
    pub tx_hash: *mut c_char,
    pub message_to_sign: *mut c_char,
    pub err: *mut c_char,
}

/// Mirrors `CreateOrderTxReq` from lighter.h
#[repr(C)]
pub struct CreateOrderTxReq {
    pub market_index: i16,
    pub client_order_index: i64,
    pub base_amount: i64,
    pub price: u32,
    pub is_ask: u8,
    pub r#type: u8,
    pub time_in_force: u8,
    pub reduce_only: u8,
    pub trigger_price: u32,
    pub order_expiry: i64,
}

// -------------------------------------------------------------------------
// Idiomatic Rust wrappers
// -------------------------------------------------------------------------

#[derive(Debug)]
pub struct StrOrErr {
    pub value: Option<String>,
    pub err: Option<String>,
}

impl StrOrErr {
    pub fn unwrap_value(self) -> Result<String, String> {
        match self.err {
            Some(e) => Err(e),
            None => Ok(self.value.unwrap_or_default()),
        }
    }
}

#[derive(Debug)]
pub struct ApiKeyResponse {
    pub private_key: Option<String>,
    pub public_key: Option<String>,
    pub err: Option<String>,
}

impl ApiKeyResponse {
    pub fn check(self) -> Result<(String, String), String> {
        match self.err {
            Some(e) => Err(e),
            None => Ok((
                self.private_key.unwrap_or_default(),
                self.public_key.unwrap_or_default(),
            )),
        }
    }
}

#[derive(Debug)]
pub struct SignedTxResponse {
    pub tx_type: u8,
    pub tx_info: Option<String>,
    pub tx_hash: Option<String>,
    pub message_to_sign: Option<String>,
    pub err: Option<String>,
}

impl SignedTxResponse {
    pub fn check(self) -> Result<Self, String> {
        match self.err {
            Some(e) => Err(e),
            None => Ok(self),
        }
    }
}

// -------------------------------------------------------------------------
// Internal helpers
// -------------------------------------------------------------------------

unsafe fn ptr_to_string(ptr: *mut c_char) -> Option<String> {
    if ptr.is_null() {
        None
    } else {
        Some(CStr::from_ptr(ptr).to_string_lossy().into_owned())
    }
}

fn raw_to_signed_tx(raw: RawSignedTxResponse) -> SignedTxResponse {
    unsafe {
        SignedTxResponse {
            tx_type: raw.tx_type,
            tx_info: ptr_to_string(raw.tx_info),
            tx_hash: ptr_to_string(raw.tx_hash),
            message_to_sign: ptr_to_string(raw.message_to_sign),
            err: ptr_to_string(raw.err),
        }
    }
}

// -------------------------------------------------------------------------
// LighterLib — loads the shared library dynamically (mirrors Java JNA usage)
// -------------------------------------------------------------------------

pub struct LighterLib {
    lib: Library,
}

// The Go shared library uses its own goroutine scheduler and internal locking;
// all exported C functions are safe to call from multiple threads concurrently.
unsafe impl Send for LighterLib {}
unsafe impl Sync for LighterLib {}

impl LighterLib {
    /// Load by absolute path.
    pub fn load(path: &str) -> Result<Self, libloading::Error> {
        let lib = unsafe { Library::new(path)? };
        Ok(Self { lib })
    }

    /// Load `lighter.{dylib,so}` from a directory relative to the working directory.
    pub fn load_from_dir(dir: &str) -> Result<Self, libloading::Error> {
        let ext = if cfg!(target_os = "macos") { "dylib" } else { "so" };
        let mut path = std::env::current_dir().expect("current dir");
        path.push(dir);
        path.push(format!("lighter.{}", ext));
        Self::load(path.to_string_lossy().as_ref())
    }

    // -------------------------------------------------------------------------
    // Functions
    // -------------------------------------------------------------------------

    pub fn generate_api_key(&self) -> ApiKeyResponse {
        unsafe {
            let f: Symbol<unsafe extern "C" fn() -> RawApiKeyResponse> =
                self.lib.get(b"GenerateAPIKey\0").unwrap();
            let raw = f();
            ApiKeyResponse {
                private_key: ptr_to_string(raw.private_key),
                public_key: ptr_to_string(raw.public_key),
                err: ptr_to_string(raw.err),
            }
        }
    }

    /// Returns `None` on success, `Some(err_message)` on failure.
    pub fn create_client(
        &self,
        url: Option<&str>,
        private_key: &str,
        chain_id: i32,
        api_key_index: i32,
        account_index: i64,
    ) -> Option<String> {
        let url_c = url.map(|u| CString::new(u).unwrap());
        let pk_c = CString::new(private_key).unwrap();
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(*mut c_char, *mut c_char, i32, i32, i64) -> *mut c_char,
            > = self.lib.get(b"CreateClient\0").unwrap();
            let url_ptr = url_c
                .as_ref()
                .map_or(std::ptr::null_mut(), |c| c.as_ptr() as *mut c_char);
            ptr_to_string(f(
                url_ptr,
                pk_c.as_ptr() as *mut c_char,
                chain_id,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn check_client(&self, api_key_index: i32, account_index: i64) -> Option<String> {
        unsafe {
            let f: Symbol<unsafe extern "C" fn(i32, i64) -> *mut c_char> =
                self.lib.get(b"CheckClient\0").unwrap();
            ptr_to_string(f(api_key_index, account_index))
        }
    }

    pub fn sign_change_pub_key(
        &self,
        pub_key: &str,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        let pk_c = CString::new(pub_key).unwrap();
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(*mut c_char, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignChangePubKey\0").unwrap();
            raw_to_signed_tx(f(
                pk_c.as_ptr() as *mut c_char,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn sign_create_order(
        &self,
        market_index: i32,
        client_order_index: i64,
        base_amount: i64,
        price: i32,
        is_ask: i32,
        order_type: i32,
        time_in_force: i32,
        reduce_only: i32,
        trigger_price: i32,
        order_expiry: i64,
        integrator_account_index: i64,
        integrator_taker_fee: i32,
        integrator_maker_fee: i32,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(
                    i32, i64, i64, i32, i32, i32, i32, i32, i32, i64,
                    i64, i32, i32, u8, i64, i32, i64,
                ) -> RawSignedTxResponse,
            > = self.lib.get(b"SignCreateOrder\0").unwrap();
            raw_to_signed_tx(f(
                market_index,
                client_order_index,
                base_amount,
                price,
                is_ask,
                order_type,
                time_in_force,
                reduce_only,
                trigger_price,
                order_expiry,
                integrator_account_index,
                integrator_taker_fee,
                integrator_maker_fee,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn sign_create_grouped_orders(
        &self,
        grouping_type: u8,
        orders: &[CreateOrderTxReq],
        integrator_account_index: i64,
        integrator_taker_fee: i32,
        integrator_maker_fee: i32,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(
                    u8, *const CreateOrderTxReq, i32,
                    i64, i32, i32, u8, i64, i32, i64,
                ) -> RawSignedTxResponse,
            > = self.lib.get(b"SignCreateGroupedOrders\0").unwrap();
            raw_to_signed_tx(f(
                grouping_type,
                orders.as_ptr(),
                orders.len() as i32,
                integrator_account_index,
                integrator_taker_fee,
                integrator_maker_fee,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn sign_cancel_order(
        &self,
        market_index: i32,
        order_index: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i32, i64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignCancelOrder\0").unwrap();
            raw_to_signed_tx(f(
                market_index,
                order_index,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn sign_withdraw(
        &self,
        asset_index: i32,
        route_type: i32,
        amount: u64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i32, i32, u64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignWithdraw\0").unwrap();
            raw_to_signed_tx(f(
                asset_index, route_type, amount, skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    pub fn sign_create_sub_account(
        &self,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignCreateSubAccount\0").unwrap();
            raw_to_signed_tx(f(skip_nonce, nonce, api_key_index, account_index))
        }
    }

    pub fn sign_cancel_all_orders(
        &self,
        time_in_force: i32,
        time: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i32, i64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignCancelAllOrders\0").unwrap();
            raw_to_signed_tx(f(
                time_in_force, time, skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn sign_modify_order(
        &self,
        market_index: i32,
        index: i64,
        base_amount: i64,
        price: i64,
        trigger_price: i64,
        integrator_account_index: i64,
        integrator_taker_fee: i32,
        integrator_maker_fee: i32,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(
                    i32, i64, i64, i64, i64,
                    i64, i32, i32, u8, i64, i32, i64,
                ) -> RawSignedTxResponse,
            > = self.lib.get(b"SignModifyOrder\0").unwrap();
            raw_to_signed_tx(f(
                market_index, index, base_amount, price, trigger_price,
                integrator_account_index, integrator_taker_fee, integrator_maker_fee,
                skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn sign_transfer(
        &self,
        to_account_index: i64,
        asset_index: i16,
        from_route_type: u8,
        to_route_type: u8,
        amount: i64,
        usdc_fee: i64,
        memo: &str,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        let memo_c = CString::new(memo).unwrap();
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(
                    i64, i16, u8, u8, i64, i64, *mut c_char,
                    u8, i64, i32, i64,
                ) -> RawSignedTxResponse,
            > = self.lib.get(b"SignTransfer\0").unwrap();
            raw_to_signed_tx(f(
                to_account_index,
                asset_index,
                from_route_type,
                to_route_type,
                amount,
                usdc_fee,
                memo_c.as_ptr() as *mut c_char,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn sign_create_public_pool(
        &self,
        operator_fee: i64,
        initial_total_shares: i32,
        min_operator_share_rate: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i64, i32, i64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignCreatePublicPool\0").unwrap();
            raw_to_signed_tx(f(
                operator_fee,
                initial_total_shares,
                min_operator_share_rate,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn sign_update_public_pool(
        &self,
        public_pool_index: i64,
        status: i32,
        operator_fee: i64,
        min_operator_share_rate: i32,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i64, i32, i64, i32, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignUpdatePublicPool\0").unwrap();
            raw_to_signed_tx(f(
                public_pool_index,
                status,
                operator_fee,
                min_operator_share_rate,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn sign_mint_shares(
        &self,
        public_pool_index: i64,
        share_amount: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i64, i64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignMintShares\0").unwrap();
            raw_to_signed_tx(f(
                public_pool_index, share_amount, skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    pub fn sign_burn_shares(
        &self,
        public_pool_index: i64,
        share_amount: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i64, i64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignBurnShares\0").unwrap();
            raw_to_signed_tx(f(
                public_pool_index, share_amount, skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    pub fn sign_update_leverage(
        &self,
        market_index: i32,
        initial_margin_fraction: i32,
        margin_mode: i32,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i32, i32, i32, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignUpdateLeverage\0").unwrap();
            raw_to_signed_tx(f(
                market_index,
                initial_margin_fraction,
                margin_mode,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn create_auth_token(
        &self,
        deadline: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> StrOrErr {
        unsafe {
            let f: Symbol<unsafe extern "C" fn(i64, i32, i64) -> RawStrOrErr> =
                self.lib.get(b"CreateAuthToken\0").unwrap();
            let raw = f(deadline, api_key_index, account_index);
            StrOrErr {
                value: ptr_to_string(raw.str_),
                err: ptr_to_string(raw.err),
            }
        }
    }

    pub fn sign_update_margin(
        &self,
        market_index: i32,
        usdc_amount: i64,
        direction: i32,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i32, i64, i32, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignUpdateMargin\0").unwrap();
            raw_to_signed_tx(f(
                market_index, usdc_amount, direction, skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    pub fn sign_stake_assets(
        &self,
        staking_pool_index: i64,
        share_amount: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i64, i64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignStakeAssets\0").unwrap();
            raw_to_signed_tx(f(
                staking_pool_index, share_amount, skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    pub fn sign_unstake_assets(
        &self,
        staking_pool_index: i64,
        share_amount: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(i64, i64, u8, i64, i32, i64) -> RawSignedTxResponse,
            > = self.lib.get(b"SignUnstakeAssets\0").unwrap();
            raw_to_signed_tx(f(
                staking_pool_index, share_amount, skip_nonce, nonce, api_key_index, account_index,
            ))
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn sign_approve_integrator(
        &self,
        integrator_index: i64,
        max_perps_taker_fee: u32,
        max_perps_maker_fee: u32,
        max_spot_taker_fee: u32,
        max_spot_maker_fee: u32,
        approval_expiry: i64,
        skip_nonce: u8,
        nonce: i64,
        api_key_index: i32,
        account_index: i64,
    ) -> SignedTxResponse {
        unsafe {
            let f: Symbol<
                unsafe extern "C" fn(
                    i64, u32, u32, u32, u32, i64, u8, i64, i32, i64,
                ) -> RawSignedTxResponse,
            > = self.lib.get(b"SignApproveIntegrator\0").unwrap();
            raw_to_signed_tx(f(
                integrator_index,
                max_perps_taker_fee,
                max_perps_maker_fee,
                max_spot_taker_fee,
                max_spot_maker_fee,
                approval_expiry,
                skip_nonce,
                nonce,
                api_key_index,
                account_index,
            ))
        }
    }

    pub fn free(&self, ptr: *mut std::ffi::c_void) {
        unsafe {
            let f: Symbol<unsafe extern "C" fn(*mut std::ffi::c_void)> =
                self.lib.get(b"Free\0").unwrap();
            f(ptr);
        }
    }
}