# Prometheus Wireguard Exporter
A simple minimalistic wireguard connection stats exporter for Prometheus.
 
## Usage
```wireguard_exporter -p 9011 -i=wg1,wg2,wg3```
| Flag | Descriptions  |  Specs                    |
| :-------- | :------- | :-------------------------------- |
| `-p` | exporter listning port| No(monitors all if not specifed)|
| `-i` | list of comma seperated interface names to monitor  | No(defaults to 9011)| 

# Exported metrics
- LatestHandshake 
- Bytes Received
- Bytes Transmitted

<img width="1508" alt="Screenshot 2024-02-19 at 6 01 37 PM" src="https://github.com/sathiraumesh/Prometheus-Wireguard-Exporter/assets/28914919/83327a18-ff5b-426a-bce8-bcbdb6750606">

## Deployment
Currently, there are no binaries. To build from the source run the following command in the project repository. Just so you know, this build is not the static binary.

```bash
  make
```

## Test Localy
```bash
  make test
```

## Run Locally
This small setup was created to simulate and show the exporter in action. I have created an environment with multiple containers communicating via wireguard VPN. The setup includes Prometheus and Grafana configured to showcase the metrics. To start setup clone the project and go to the project directory


Make sure docker, docker-compose, and make utility is installed. Run the following command to create a setup 


Run the project in a local setup
```bash
  make run
```

Monitor the metrics using Grafana Dashboard using the default password and username 
```admin, admin```

Import the dashboard from path 
```setup/grafana-provisioning```
```bash
http://localhost:3000/dashboards
```
