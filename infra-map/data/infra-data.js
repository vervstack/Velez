window.INFRA_DATA = {
 "generated_at": "2026-11-04 22:39:53 +0300",
 "nodes": [
  {
   "id": "sc",
   "ssh": "sc",
   "name": "sc",
   "title": "Home server (Raspberry Pi 5)",
   "role": "home / edge",
   "location": "Home, UTC+3",
   "lat": 55.7558,
   "lon": 37.6173,
   "public_ip": "",
   "tailscale": true,
   "note": "Raspberry Pi 5, NVMe 1TB, WLAN. Coordinates are a placeholder for home — edit lat/lon here.",
   "offline": false,
   "hostname": "rp",
   "os": "Ubuntu 25.10, kernel 6.17.0-1011-raspi (aarch64)",
   "cpu_model": "ARM Cortex-A76 (BCM2712)",
   "cpu_cores": 4,
   "ram_total_mb": 7934,
   "ram_avail_mb": 6032,
   "disk_total_gb": 938.7,
   "disk_used_gb": 323.6,
   "load": [
    "1.81",
    "1.52",
    "1.23"
   ],
   "uptime_days": 1.3,
   "docker_version": "29.4.1",
   "tailnet_ip": "100.64.0.37",
   "containers_running": 28,
   "containers_total": 31,
   "containers": [
    {
     "name": "gitlab_runner_artel",
     "image": "gitlab/gitlab-runner:latest",
     "status": "Up About an hour",
     "state": "running",
     "project": "gitlab_runner_artel",
     "networks": [
      "arteldind-net",
      "bridge"
     ],
     "sidecar": false,
     "velez_service": "gitlab_runner_artel"
    },
    {
     "name": "velez",
     "image": "vervstack/velez:v0.1.120-rc",
     "status": "Up 32 hours",
     "state": "running",
     "project": "",
     "networks": [
      "bridge",
      "verv"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "webserver",
     "image": "nginx:1.31.2",
     "status": "Up 32 hours",
     "state": "running",
     "project": "web",
     "networks": [],
     "sidecar": false,
     "velez_service": "webserver"
    },
    {
     "name": "webserver-ts-sidecar",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "web",
     "networks": [
      "web_default"
     ],
     "sidecar": true,
     "velez_service": "webserver"
    },
    {
     "name": "zpotify-prod",
     "image": "redsockruf/zpotify:v1.2.12",
     "status": "Up 32 hours",
     "state": "running",
     "project": "zpotify",
     "networks": [],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "arteldind",
     "image": "docker:dind",
     "status": "Up 32 hours",
     "state": "running",
     "project": "arteldind",
     "networks": [
      "arteldind-net",
      "bridge"
     ],
     "sidecar": false,
     "velez_service": "arteldind"
    },
    {
     "name": "s3_spb_web_ui",
     "image": "khairul169/garage-webui:1.1.0",
     "status": "Up 32 hours (healthy)",
     "state": "running",
     "project": "s3_spb",
     "networks": [],
     "sidecar": true,
     "velez_service": "s3_spb"
    },
    {
     "name": "s3_spb",
     "image": "dxflrs/garage:v2.4.0",
     "status": "Up 32 hours (healthy)",
     "state": "running",
     "project": "s3_spb",
     "networks": [
      "bridge",
      "s3-spb",
      "verv"
     ],
     "sidecar": false,
     "velez_service": "s3_spb"
    },
    {
     "name": "pgaas_fintech",
     "image": "postgres:18",
     "status": "Up 32 hours (healthy)",
     "state": "running",
     "project": "pgaas_fintech",
     "networks": [
      "bridge",
      "verv"
     ],
     "sidecar": false,
     "velez_service": "pgaas_fintech"
    },
    {
     "name": "flaresolverr",
     "image": "ghcr.io/flaresolverr/flaresolverr",
     "status": "Up 32 hours",
     "state": "running",
     "project": "zpotify",
     "networks": [],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "hysteria-client",
     "image": "tobyxdd/hysteria:v2.13.0",
     "status": "Up 32 hours",
     "state": "running",
     "project": "hyst",
     "networks": [
      "hyst_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "prowlarr",
     "image": "51431c0b26c9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "zpotify",
     "networks": [],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "postgres-db-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "postgres",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "postgres-db",
     "image": "postgres:18",
     "status": "Up 32 hours",
     "state": "running",
     "project": "postgres",
     "networks": [
      "postgres_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "s3_tst_web_ui",
     "image": "khairul169/garage-webui:1.1.0",
     "status": "Exited (255) 13 days ago",
     "state": "exited",
     "project": "s3_tst",
     "networks": [],
     "sidecar": true,
     "velez_service": "s3_tst"
    },
    {
     "name": "zpotify_webserver",
     "image": "nginx:1.31.2",
     "status": "Up 32 hours",
     "state": "running",
     "project": "zpotify",
     "networks": [],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "zpotify-prod-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "zpotify",
     "networks": [
      "zpotify_default"
     ],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "grafana-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "loki-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "victoriametrics-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Exited (255) 13 days ago",
     "state": "exited",
     "project": "grafana",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "promtail-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "vmagent-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "grafana",
     "image": "grafana/grafana:12.3",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [
      "grafana_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "promtail",
     "image": "grafana/promtail:3.2.1",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [
      "grafana_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "vmagent",
     "image": "victoriametrics/vmagent:v1.115.0",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [
      "metrics"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "node-exporter",
     "image": "prom/node-exporter:v1.9.1",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [
      "metrics"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "cadvisor",
     "image": "gcr.io/cadvisor/cadvisor:v0.52.1",
     "status": "Up 32 hours (healthy)",
     "state": "running",
     "project": "grafana",
     "networks": [
      "metrics"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "loki",
     "image": "grafana/loki",
     "status": "Up 32 hours",
     "state": "running",
     "project": "grafana",
     "networks": [
      "grafana_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "victoriametrics",
     "image": "victoriametrics/victoria-metrics:v1.115.0",
     "status": "Exited (0) 4 weeks ago",
     "state": "exited",
     "project": "grafana",
     "networks": [
      "grafana_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "obsidian-couchdb-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 32 hours",
     "state": "running",
     "project": "obsidian",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "obsidian-couchdb",
     "image": "couchdb:latest",
     "status": "Up 32 hours",
     "state": "running",
     "project": "obsidian",
     "networks": [
      "obsidian_default"
     ],
     "sidecar": false,
     "velez_service": ""
    }
   ],
   "networks": [
    {
     "name": "artel-net",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "arteldind2-net",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "arteldind-net",
     "driver": "bridge",
     "containers": [
      "arteldind",
      "gitlab_runner_artel"
     ]
    },
    {
     "name": "bridge",
     "driver": "bridge",
     "containers": [
      "s3_spb",
      "arteldind",
      "velez",
      "gitlab_runner_artel",
      "pgaas_fintech"
     ]
    },
    {
     "name": "deployer_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "gitlab_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "grafana_default",
     "driver": "bridge",
     "containers": [
      "grafana",
      "promtail",
      "loki"
     ]
    },
    {
     "name": "host",
     "driver": "host",
     "containers": []
    },
    {
     "name": "hyst_default",
     "driver": "bridge",
     "containers": [
      "hysteria-client"
     ]
    },
    {
     "name": "metrics",
     "driver": "bridge",
     "containers": [
      "vmagent",
      "cadvisor",
      "node-exporter"
     ]
    },
    {
     "name": "none",
     "driver": "null",
     "containers": []
    },
    {
     "name": "obsidian_default",
     "driver": "bridge",
     "containers": [
      "obsidian-couchdb"
     ]
    },
    {
     "name": "pi-hole_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "postgres_default",
     "driver": "bridge",
     "containers": [
      "postgres-db"
     ]
    },
    {
     "name": "s3-spb",
     "driver": "bridge",
     "containers": [
      "s3_spb"
     ]
    },
    {
     "name": "s3-teldwad",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "s3-tst",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "verv",
     "driver": "bridge",
     "containers": [
      "s3_spb",
      "velez",
      "pgaas_fintech"
     ]
    },
    {
     "name": "vervpn_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "web_default",
     "driver": "bridge",
     "containers": [
      "webserver-ts-sidecar"
     ]
    },
    {
     "name": "workbench-net",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "zpotify_default",
     "driver": "bridge",
     "containers": [
      "zpotify-prod-ts"
     ]
    }
   ]
  },
  {
   "id": "germ",
   "ssh": "germ",
   "name": "germ",
   "title": "Hostkey Germany",
   "role": "control plane / prod",
   "location": "Frankfurt am Main, DE",
   "lat": 50.1155,
   "lon": 8.6842,
   "public_ip": "151.243.171.247",
   "tailscale": false,
   "headscale_control": true,
   "note": "Headscale control plane, artel prod, hysteria server, minecraft.",
   "offline": false,
   "hostname": "root",
   "os": "Ubuntu 24.04.3 LTS, kernel 6.8.0-90-generic (x86_64)",
   "cpu_model": "AMD EPYC 7542 32-Core Processor",
   "cpu_cores": 4,
   "ram_total_mb": 7918,
   "ram_avail_mb": 2574,
   "disk_total_gb": 117.0,
   "disk_used_gb": 46.3,
   "load": [
    "0.57",
    "0.44",
    "0.44"
   ],
   "uptime_days": 319.1,
   "docker_version": "29.1.1",
   "tailnet_ip": "",
   "containers_running": 28,
   "containers_total": 30,
   "containers": [
    {
     "name": "artel-prod",
     "image": "redsockruf/artel:v1.3.0",
     "status": "Up 2 hours",
     "state": "running",
     "project": "artel",
     "networks": [],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "artel-prod-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 2 weeks",
     "state": "running",
     "project": "artel",
     "networks": [
      "redsockru"
     ],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "velez",
     "image": "vervstack/velez:v0.1.119-rc",
     "status": "Up 2 weeks",
     "state": "running",
     "project": "",
     "networks": [
      "bridge",
      "verv"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "hysteria-server",
     "image": "tobyxdd/hysteria:v2.13.0",
     "status": "Up 2 weeks",
     "state": "running",
     "project": "hysteria",
     "networks": [
      "hysteria_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "webserver",
     "image": "nginx:1.31.2",
     "status": "Up 3 weeks",
     "state": "running",
     "project": "web",
     "networks": [
      "xray-bridge",
      "redsockru"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "web-certbot-1",
     "image": "certbot/certbot:latest",
     "status": "Exited (1) 3 weeks ago",
     "state": "exited",
     "project": "web",
     "networks": [
      "redsockru"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "webserver-ts-sidecar",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 3 weeks",
     "state": "running",
     "project": "web",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "remnanode",
     "image": "remnawave/node:latest",
     "status": "Up 3 weeks",
     "state": "running",
     "project": "remnawavenode",
     "networks": [
      "host"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "headscale",
     "image": "headscale/headscale:0.27.2-rc.1",
     "status": "Up 3 months (healthy)",
     "state": "running",
     "project": "vervpn",
     "networks": [
      "redsockru"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "ollama-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 5 months",
     "state": "running",
     "project": "openclaw",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "openclaw-gateway",
     "image": "ghcr.io/openclaw/openclaw:latest",
     "status": "Up 5 months (healthy)",
     "state": "running",
     "project": "openclaw",
     "networks": [
      "openclaw_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "obsidian-mcp",
     "image": "ghcr.io/es617/obsidian-sync-mcp:latest",
     "status": "Up 6 months",
     "state": "running",
     "project": "obsidian-mcp",
     "networks": [
      "redsockru"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "mtproto-proxy-2",
     "image": "telegrammessenger/proxy:1.4",
     "status": "Up 7 months",
     "state": "running",
     "project": "mtproxy",
     "networks": [
      "mtproxy_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "mtproto-proxy",
     "image": "telegrammessenger/proxy:1.4",
     "status": "Up 7 months",
     "state": "running",
     "project": "mtproxy",
     "networks": [
      "mtproxy_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "mead-prod",
     "image": "redsockdev/mead:v0.0.18",
     "status": "Up 7 months",
     "state": "running",
     "project": "mead-prod",
     "networks": [
      "mead-prod_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "mead-prod-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 7 months",
     "state": "running",
     "project": "mead-prod",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "mead-stg",
     "image": "redsockdev/mead:v0.0.18",
     "status": "Up 7 months",
     "state": "running",
     "project": "mead",
     "networks": [
      "mead_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "mead-stg-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 7 months",
     "state": "running",
     "project": "mead",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "main-pg",
     "image": "postgres:18",
     "status": "Up 7 months",
     "state": "running",
     "project": "postgres",
     "networks": [
      "postgres_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "main-pg-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 7 months",
     "state": "running",
     "project": "postgres",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "vmagent-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 5 months",
     "state": "running",
     "project": "log",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "vmagent",
     "image": "victoriametrics/vmagent:v1.115.0",
     "status": "Up 5 months",
     "state": "running",
     "project": "log",
     "networks": [
      "metrics"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "node-exporter",
     "image": "prom/node-exporter:v1.9.1",
     "status": "Up 5 months",
     "state": "running",
     "project": "log",
     "networks": [
      "metrics"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "cadvisor",
     "image": "gcr.io/cadvisor/cadvisor:v0.52.1",
     "status": "Up 5 months (healthy)",
     "state": "running",
     "project": "log",
     "networks": [
      "metrics"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "promtail-ts",
     "image": "tailscale/tailscale:v1.90.9",
     "status": "Up 5 months",
     "state": "running",
     "project": "log",
     "networks": [],
     "sidecar": true,
     "velez_service": ""
    },
    {
     "name": "promtail",
     "image": "grafana/promtail:3.2.1",
     "status": "Up 5 months",
     "state": "running",
     "project": "log",
     "networks": [
      "log_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "minecraft",
     "image": "itzg/minecraft-server:latest",
     "status": "Up 2 months (healthy)",
     "state": "running",
     "project": "minecraft",
     "networks": [
      "redsockru"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "socks5",
     "image": "dante-socks5",
     "status": "Exited (0) 7 months ago",
     "state": "exited",
     "project": "dante",
     "networks": [
      "dante_default"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "marzban",
     "image": "gozargah/marzban:latest",
     "status": "Up 10 months",
     "state": "running",
     "project": "marzban",
     "networks": [
      "redsockru"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "s3",
     "image": "minio/minio",
     "status": "Up 10 months",
     "state": "running",
     "project": "s3",
     "networks": [
      "redsockru"
     ],
     "sidecar": false,
     "velez_service": ""
    }
   ],
   "networks": [
    {
     "name": "bridge",
     "driver": "bridge",
     "containers": [
      "velez"
     ]
    },
    {
     "name": "dante_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "host",
     "driver": "host",
     "containers": [
      "remnanode"
     ]
    },
    {
     "name": "hysteria_default",
     "driver": "bridge",
     "containers": [
      "hysteria-server"
     ]
    },
    {
     "name": "log_default",
     "driver": "bridge",
     "containers": [
      "promtail"
     ]
    },
    {
     "name": "mead-prod_default",
     "driver": "bridge",
     "containers": [
      "mead-prod"
     ]
    },
    {
     "name": "mead_default",
     "driver": "bridge",
     "containers": [
      "mead-stg"
     ]
    },
    {
     "name": "metrics",
     "driver": "bridge",
     "containers": [
      "cadvisor",
      "vmagent",
      "node-exporter"
     ]
    },
    {
     "name": "minecraft_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "mtproxy_default",
     "driver": "bridge",
     "containers": [
      "mtproto-proxy-2",
      "mtproto-proxy"
     ]
    },
    {
     "name": "none",
     "driver": "null",
     "containers": []
    },
    {
     "name": "openclaw_default",
     "driver": "bridge",
     "containers": [
      "openclaw-gateway"
     ]
    },
    {
     "name": "postgres_default",
     "driver": "bridge",
     "containers": [
      "main-pg"
     ]
    },
    {
     "name": "redsockru",
     "driver": "bridge",
     "containers": [
      "obsidian-mcp",
      "webserver",
      "s3",
      "minecraft",
      "headscale",
      "artel-prod-ts",
      "marzban"
     ]
    },
    {
     "name": "verv",
     "driver": "bridge",
     "containers": [
      "velez"
     ]
    },
    {
     "name": "vervpn-client_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "vervpn_default",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "xray-bridge",
     "driver": "bridge",
     "containers": [
      "webserver"
     ]
    }
   ]
  },
  {
   "id": "nl",
   "ssh": "nl",
   "name": "nl",
   "title": "Hostkey Netherlands",
   "role": "prod",
   "location": "Amsterdam, NL",
   "lat": 52.374,
   "lon": 4.8897,
   "public_ip": "146.0.79.91",
   "tailscale": false,
   "note": "Remnawave panel + node, artel DinD, zpotify-dev S3, pgaas.",
   "offline": false,
   "hostname": "NL16362",
   "os": "Ubuntu 24.04.5 LTS, kernel 6.8.0-39-generic (x86_64)",
   "cpu_model": "Intel(R) Xeon(R) CPU E5-2680 v2 @ 2.80GHz",
   "cpu_cores": 2,
   "ram_total_mb": 1893,
   "ram_avail_mb": 269,
   "disk_total_gb": 58.0,
   "disk_used_gb": 10.4,
   "load": [
    "0.69",
    "0.78",
    "1.34"
   ],
   "uptime_days": 26.2,
   "docker_version": "29.8.2",
   "tailnet_ip": "",
   "containers_running": 10,
   "containers_total": 10,
   "containers": [
    {
     "name": "velez",
     "image": "vervstack/velez:v0.1.111",
     "status": "Up 2 weeks",
     "state": "running",
     "project": "",
     "networks": [
      "bridge",
      "verv"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "s3_zpotify-dev_web_ui",
     "image": "khairul169/garage-webui:1.1.0",
     "status": "Up 2 weeks (healthy)",
     "state": "running",
     "project": "s3_zpotify-dev",
     "networks": [],
     "sidecar": true,
     "velez_service": "s3_zpotify-dev"
    },
    {
     "name": "s3_zpotify-dev",
     "image": "dxflrs/garage:v2.4.0",
     "status": "Up 2 weeks (healthy)",
     "state": "running",
     "project": "s3_zpotify-dev",
     "networks": [
      "bridge",
      "s3-zpotify-dev",
      "verv"
     ],
     "sidecar": false,
     "velez_service": "s3_zpotify-dev"
    },
    {
     "name": "artel",
     "image": "docker:dind",
     "status": "Up 2 weeks",
     "state": "running",
     "project": "artel",
     "networks": [
      "bridge",
      "artel-net"
     ],
     "sidecar": false,
     "velez_service": "artel"
    },
    {
     "name": "caddy",
     "image": "caddy:2.9",
     "status": "Up 2 weeks",
     "state": "running",
     "project": "caddy",
     "networks": [
      "remnawave-network",
      "xray-bridge"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "pgaas_artel-finance",
     "image": "postgres:18",
     "status": "Up 2 weeks (healthy)",
     "state": "running",
     "project": "pgaas_artel-finance",
     "networks": [
      "bridge",
      "verv"
     ],
     "sidecar": false,
     "velez_service": "pgaas_artel-finance"
    },
    {
     "name": "remnanode",
     "image": "remnawave/node:latest",
     "status": "Up 3 weeks",
     "state": "running",
     "project": "remnanode",
     "networks": [
      "host"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "remnawave",
     "image": "remnawave/backend:3",
     "status": "Up 3 weeks (healthy)",
     "state": "running",
     "project": "remnawave",
     "networks": [
      "remnawave-network"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "remnawave-db",
     "image": "postgres:18.4",
     "status": "Up 3 weeks (healthy)",
     "state": "running",
     "project": "remnawave",
     "networks": [
      "remnawave-network"
     ],
     "sidecar": false,
     "velez_service": ""
    },
    {
     "name": "remnawave-redis",
     "image": "valkey/valkey:9-alpine",
     "status": "Up 3 weeks (healthy)",
     "state": "running",
     "project": "remnawave",
     "networks": [
      "remnawave-network"
     ],
     "sidecar": false,
     "velez_service": ""
    }
   ],
   "networks": [
    {
     "name": "artel-net",
     "driver": "bridge",
     "containers": [
      "artel"
     ]
    },
    {
     "name": "arteldind-net",
     "driver": "bridge",
     "containers": []
    },
    {
     "name": "bridge",
     "driver": "bridge",
     "containers": [
      "velez",
      "pgaas_artel-finance",
      "s3_zpotify-dev",
      "artel"
     ]
    },
    {
     "name": "host",
     "driver": "host",
     "containers": [
      "remnanode"
     ]
    },
    {
     "name": "none",
     "driver": "null",
     "containers": []
    },
    {
     "name": "remnawave-network",
     "driver": "bridge",
     "containers": [
      "caddy",
      "remnawave",
      "remnawave-redis",
      "remnawave-db"
     ]
    },
    {
     "name": "s3-zpotify-dev",
     "driver": "bridge",
     "containers": [
      "s3_zpotify-dev"
     ]
    },
    {
     "name": "verv",
     "driver": "bridge",
     "containers": [
      "velez",
      "pgaas_artel-finance",
      "s3_zpotify-dev"
     ]
    },
    {
     "name": "xray-bridge",
     "driver": "bridge",
     "containers": [
      "caddy"
     ]
    }
   ]
  }
 ],
 "links": [
  {
   "a": "sc",
   "b": "germ",
   "kind": "tailnet",
   "label": "headscale tailnet"
  },
  {
   "a": "sc",
   "b": "germ",
   "kind": "docker-net",
   "label": "metrics"
  },
  {
   "a": "sc",
   "b": "germ",
   "kind": "docker-net",
   "label": "verv"
  },
  {
   "a": "sc",
   "b": "nl",
   "kind": "docker-net",
   "label": "artel-net"
  },
  {
   "a": "sc",
   "b": "nl",
   "kind": "docker-net",
   "label": "arteldind-net"
  },
  {
   "a": "sc",
   "b": "nl",
   "kind": "docker-net",
   "label": "verv"
  },
  {
   "a": "germ",
   "b": "nl",
   "kind": "docker-net",
   "label": "verv"
  },
  {
   "a": "germ",
   "b": "nl",
   "kind": "docker-net",
   "label": "xray-bridge"
  }
 ]
};
