use anyhow::{Context, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
struct Route {
    #[serde(rename = "@id")]
    id: String,
    #[serde(rename = "match")]
    matchs: Vec<Match>,
    handle: Vec<Handle>,
}

#[derive(Serialize, Deserialize)]
struct Match {
    host: Vec<String>,
}

#[derive(Serialize, Deserialize)]
struct Handle {
    handler: String,
    upstreams: Vec<Upstream>,
}

#[derive(Serialize, Deserialize)]
struct Upstream {
    dial: String,
}

const ENDPOINT: &str = "http://127.0.0.1:2019";

pub async fn add_route(target_port: u32) -> Result<()> {
    let route_id = format!("route-{}", target_port);
    let target = format!("127.0.0.1:{}", target_port);

    let route = Route {
        id: route_id.clone(),
        matchs: vec![Match {
            host: vec!["127.0.0.1".to_string()],
        }],
        handle: vec![Handle {
            handler: "reverse_proxy".to_string(),
            upstreams: vec![Upstream {
                dial: format!(":{}", target_port),
            }],
        }],
    };

    let client = Client::new();
    let res = client
        .post(&format!(
            "{}/config/apps/http/servers/srv0/routes",
            ENDPOINT
        ))
        .json(&route)
        .send()
        .await
        .context("route creation failed")?;

    log::info!("route created successfully, response: {:?}", res);

    let res = client
        .post(&format!(
            "{}/config/apps/tls/automation/policies/0/subjects",
            ENDPOINT
        ))
        .json(&target)
        .send()
        .await
        .context("route creation failed")?;

    log::info!(
        "route tls subjects created successfully, response: {:?}",
        res
    );

    Ok(())
}

pub async fn del_route(route_id: &str) -> Result<()> {
    let client = Client::new();

    log::info!("cleaning up route, id: {}", route_id);

    let _res = client
        .delete(&format!("{}/id/{}", ENDPOINT, route_id))
        .header("Content-Type", "application/json")
        .send()
        .await
        .context("route deletion failed")?;

    log::info!("route deleted successfully, id: {}", route_id);
    Ok(())
}
