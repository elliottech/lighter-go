use std::sync::Arc;
use std::thread;
use std::time::{Instant, SystemTime, UNIX_EPOCH};

use lighter_rust::LighterLib;

const CHAIN_ID: i32 = 304;
const ACCOUNT_INDEX: i64 = 100;
const MARKET_INDEX: i32 = 0; // ETH market
const BASE_AMOUNT: i64 = 10_000;
const PRICE: i32 = 400_000;
const ORDER_TYPE: i32 = 0; // limit
const TIME_IN_FORCE: i32 = 2; // post-only
const N_THREADS: usize = 5;
const N_ORDERS: usize = 100;

fn now_ms() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis() as i64
}

fn run_example(lib: &LighterLib, api_key_index: i32) {
    // Generate a fresh API key pair
    let api_resp = lib.generate_api_key();
    let (private_key, public_key) = match api_resp.check() {
        Ok(v) => v,
        Err(e) => {
            eprintln!("[{}] GenerateAPIKey error: {}", api_key_index, e);
            return;
        }
    };
    println!("[{}] publicKey={}", api_key_index, public_key);

    // Create a client bound to the generated key
    if let Some(err) = lib.create_client(None, &private_key, CHAIN_ID, api_key_index, ACCOUNT_INDEX)
    {
        eprintln!("[{}] CreateClient error: {}", api_key_index, err);
        return;
    }

    // Auth token valid for 7 hours
    let token_deadline = now_ms() + 7 * 60 * 60 * 1000;
    let auth_token = match lib
        .create_auth_token(token_deadline, api_key_index, ACCOUNT_INDEX)
        .unwrap_value()
    {
        Ok(t) => t,
        Err(e) => {
            eprintln!("[{}] CreateAuthToken error: {}", api_key_index, e);
            return;
        }
    };
    println!("[{}] authToken={}", api_key_index, auth_token);

    let mut nonce: i64 = 1;
    let start = Instant::now();

    for i in 1..=N_ORDERS {
        let order_expiry = now_ms() + 60 * 60 * 1000; // 60 min from now

        // Sign a limit post-only ask order
        let create = lib.sign_create_order(
            MARKET_INDEX,
            i as i64,        // client_order_index
            BASE_AMOUNT,
            PRICE,
            1,               // is_ask
            ORDER_TYPE,
            TIME_IN_FORCE,
            0,               // reduce_only
            0,               // trigger_price
            order_expiry,
            0,               // integrator_account_index
            0,               // integrator_taker_fee
            0,               // integrator_maker_fee
            0,               // skip_nonce
            nonce,
            api_key_index,
            ACCOUNT_INDEX,
        );
        nonce += 1;

        if let Some(e) = create.err {
            eprintln!("[{}] SignCreateOrder({}) error: {}", api_key_index, i, e);
        }

        // Cancel the same order by client order index
        let cancel = lib.sign_cancel_order(
            MARKET_INDEX,
            i as i64,
            0, // skip_nonce
            nonce,
            api_key_index,
            ACCOUNT_INDEX,
        );
        nonce += 1;

        if let Some(e) = cancel.err {
            eprintln!("[{}] SignCancelOrder({}) error: {}", api_key_index, i, e);
        }
    }

    let elapsed = start.elapsed();
    println!(
        "[{}] {} create+cancel pairs in {:.2} ms",
        api_key_index,
        N_ORDERS,
        elapsed.as_secs_f64() * 1000.0
    );
}

fn main() {
    let lib = Arc::new(
        LighterLib::load_from_dir("../sharedlib").expect("failed to load lighter shared library"),
    );

    let handles: Vec<_> = (0..N_THREADS)
        .map(|i| {
            let lib = Arc::clone(&lib);
            thread::spawn(move || run_example(&lib, i as i32))
        })
        .collect();

    for h in handles {
        h.join().unwrap();
    }
}
