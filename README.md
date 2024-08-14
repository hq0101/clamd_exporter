# clamd_exporter

[![Go Reference](https://pkg.go.dev/badge/github.com/hq0101/clamd_exporter.svg)](https://pkg.go.dev/github.com/hq0101/clamd_exporter)

从clamd服务导出指标以供Prometheus使用


默认情况下，clamd_exporter在/metrics的端口0.0.0.0:8080上提供服务

参数：
- --listen: 监听地址
- --address：ClamAV server address /var/run/clamav/clamd.ctl or 127.0.0.1:3310
- --nettype: tcp or unix

```shell
./clamd_exporter-v0.1.0.linux-amd64 --listen :8181 --address 192.168.127.131:3310 --nettype tcp
or
./clamd_exporter-v0.1.0.linux-amd64 --listen :8181 --address /var/run/clamav/clamd.ctl --nettype tcp
```

- Prometheus示例配置

```yaml
scrape_configs:
  - job_name: "clamd_exporter"
    static_configs:
      - targets: ["192.168.1.10:8181"]
```



# Grafana Dashboard

![Dashboard](assets/dashboard.png)

Grafana Dashboard: https://github.com/hq0101/clamd_exporter/blob/main/assets/ClamdDashboard.json

