use std::str;
use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use lazy_static::lazy_static;
use http_req::request;

mod guest;

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse {
    status: String,
    data: Option<serde_json::Value>,
}

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

#[derive(Debug, Serialize, Deserialize)]
struct WorldTime {
    currentDateTime: String,
}

#[export_name = "handle_request"]
pub fn http_request() -> u64 {
    let conf: &Config = &*CONFIG;

    let mut writer = Vec::new(); //container for body of a response
    // let res = request::get("http://worldtimeapi.org/api/timezone/europe/paris", &mut writer);
         let res = request::get("http://worldclockapi.com/api/json/est/now", &mut writer);
    let response = match res {
        Ok(response) => { response        }
        Err(e) => {
            guest::send_log(guest::ERROR, &e.to_string());
            guest::set_code(503);
            guest::writebody(guest::RESPONSE_BODY, &e.to_string());
            return 0u64
        }
    };

    let resp: WorldTime = match serde_json::from_str(format!("{}", String::from_utf8_lossy(&writer)).as_str()) {
        Result::Ok(val) => { val }
        Result::Err(e) => {
            guest::send_log(guest::ERROR, &e.to_string());
            guest::set_code(503);
            guest::writebody(guest::RESPONSE_BODY, &e.to_string());
            return 0u64
        }
    };

    guest::add_header(guest::REQUEST_HEADER, "X-Time", &resp.currentDateTime);


    for (k, v) in &conf.headers {
        guest::add_header(guest::REQUEST_HEADER, &k, &v);
    }
    return 1u64;
}

#[export_name = "handle_response"]
fn http_response(_req_ctx: i32, _is_error: i32) {}
