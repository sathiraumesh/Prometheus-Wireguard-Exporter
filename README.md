# Prometheus Wireguard Exporter
A simple minimalistic wireguard connection stats exporter for prometheus.
 
## Usage
```wireguard_exporter -p 9011 -i=wg1,wg2,wg3```
| Flag | Descriptions  |  Specs                    |
| :-------- | :------- | :-------------------------------- |
| `-p` | exporter listning port| No(monitors all if not specifed)|
| `-i` | list of comma seperated interface names to monitor  | No(defaults to 9011)| 

# Exported metrics
- LatestHandshake 
- Bytes Recived
- Bytes Transmitted

## Deployment
Currently there are no binaries. To build from source run the following command in project repository. Note that this build is not the static bianry.

```bash
  make build
```

To build the static binary use following command
```bash
  make-static build
```

## Run Locally
This is a small setup created to simulate and show the exporter in action. I have created a environment with multiple containers who are communicating via wireguard VPN. The setup includes promotheus and grafana configured to showcase the metrics. To start setup clone the project and go to the project directory


Make sure docker, docker compose and make utility  is installed. Run the following command to create a setup 

```bash
  make build-image
```

Run the project in local setup
```bash
  make run
```

Monitor the metrics using Grafana Dashboard using default password and username 
```admin, admin```

```bash
http://localhost:3000/dashboards
```
