use std::str;
use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use serde_json::Result;
use lazy_static::lazy_static;

mod guest;

lazy_static! {
    static ref CONFIG: Config = match serde_json::from_str(str::from_utf8(&guest::get_conf()).unwrap()) {
        Result::Ok(val) => {val},
        Result::Err(err) => {
            guest::send_log(guest::ERROR, &err.to_string());
            panic!("err {}", err)
        }
    };
}

fn main() {
    guest::send_log(guest::DEBUG, "Middleware started.")
}

#[derive(Debug, Serialize, Deserialize)]
struct Config {
   headers: HashMap<String, String>,
}

#[export_name="handle_request"]
pub fn http_request() -> u64 {
    let conf: &Config = &*CONFIG;

    for (k, v) in &conf.headers {
        guest::add_header(guest::REQUEST_HEADER, &k, &v);
    }
    guest::send_log(guest::DEBUG, format!("{:?}", guest::get_addr()).as_str());
    return 1 as u64;
}

#[export_name="handle_response"]
fn http_response(_req_ctx: i32, _is_error: i32) {
}
