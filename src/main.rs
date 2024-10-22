use std::str;
use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use serde_json::Result;
use lazy_static::lazy_static;
// use anyhow::Result;
// use reqwest_wasi::Client;

mod guest;

// #![feature(repr128)]
//
// // Exemple de structure pour les données JSON
// #[derive(Debug, Serialize, Deserialize)]
// struct ApiResponse {
//     status: String,
//     data: Option<serde_json::Value>,
// }

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

      // Créer un client HTTP
//             let client = Client::new();


    let conf: &Config = &*CONFIG;


//
//         // Exemple d'appel à une API externe
//         match call_external_api(&client).await {
//             Ok(api_response) => {
//                 // Ajouter les données de l'API comme header
//                 if let Some(status) = api_response.status.as_str() {
//                     guest::add_header(guest::REQUEST_HEADER, "X-API-Status", status);
//                 }
//
//                 // Si on a des données, les ajouter au body
//                 if let Some(data) = api_response.data {
//                     if let Ok(json_string) = serde_json::to_string(&data) {
//                         req.set_body(Some(HttpBody::new(json_string.into_bytes())));
//                     }
//                 }
//             }
//             Err(e) => {
//                 guest::send_log(guest::ERROR, &e.to_string();
//             }
//         }

    for (k, v) in &conf.headers {
        guest::add_header(guest::REQUEST_HEADER, &k, &v);
    }
    guest::send_log(guest::DEBUG, format!("{:?}", guest::get_addr()).as_str());
    return 1 as u64;
}

#[export_name="handle_response"]
fn http_response(_req_ctx: i32, _is_error: i32) {
}

// async fn call_external_api(client: &Client) -> Result<ApiResponse> {
//     // Exemple d'appel à une API JSON
//     let response = client
//         .get("http://worldtimeapi.org/api/timezone/europe/paris")
//         .send()
//         .await?
//         .json::<ApiResponse>()
//         .await?;
//
//     Ok(response)
// }
