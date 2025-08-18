<!-- Top Banner -->
<p align="center" style="margin-bottom:0;">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:667eea,100:764ba2&height=220&section=header&text=Nexus&fontSize=75&fontColor=ffffff&width=100%"/>
</p>
<p align="center">
  <b>A modern, real-time chat platform built for communities</b>
</p>

---

## 🚀 What is Nexus?

Nexus is a **high-performance chat application** designed to bring people together through seamless real-time communication.  
It is built with **scalable, production-grade architecture** inspired by modern messaging systems like Slack and Discord.

### ✨ Highlights

- ⚡ **Real-time messaging** with WebSocket connections
- 🏠 **Server-based communities** with organized channels
- 🔒 **JWT-based authentication** and secure user management
- 📡 **Event-driven architecture** with NATS
- 🔍 **Message search** with Elasticsearch
- 📊 **Monitored & scalable** with Kubernetes, Prometheus, and Grafana

---

## 🛠️ Tech Stack

<p align="center">
  <img src="https://skillicons.dev/icons?i=go,postgres,redis,docker,kubernetes,githubactions&theme=dark" />
</p>

- **Backend:** Go (Golang), Gin, Gorilla WebSocket, gRPC
- **Database:** PostgreSQL (message storage, metadata)
- **Caching & Presence:** Redis (sessions, pub/sub for online status)
- **Messaging Bus:** NATS (event-driven communication)
- **Search:** Elasticsearch (full-text message search)
- **Authentication:** JWT tokens
- **DevOps & Infrastructure:** Docker, Kubernetes, GitHub Actions (CI/CD)
- **Monitoring:** Prometheus + Grafana for metrics, logs, and alerting

---

## 📊 System Architecture

```mermaid
flowchart TD
    subgraph Client
        A1[Web/Mobile App]
    end

    subgraph API
        B1[API Gateway]
        B2[Auth Service<br/>JWT Tokens]
        B3[Chat Service<br/>Go + WebSockets]
        B4[gRPC Internal Services]
    end

    subgraph Infra
        C1[(PostgreSQL)]
        C2[(Redis Cache & Presence)]
        C3[(NATS Message Bus)]
        C4[(Elasticsearch)]
    end

    subgraph DevOps
        D1[Docker]
        D2[Kubernetes]
        D3[Prometheus + Grafana]
        D4[GitHub Actions CI/CD]
    end

    %% Connections
    A1 <--> B1
    B1 --> B2
    B1 --> B3
    B3 --> C1
    B3 --> C2
    B3 --> C3
    B3 --> C4
    B2 --> C1
    D1 --> D2
    D2 --> B1
    D2 --> B2
    D2 --> B3
    D2 --> B4
    D3 --> D2
    D4 --> D2
```

---

## 🤝 Contributing

We welcome contributions! Once the core system reaches stability, we'll provide detailed contribution guidelines.

For now, ⭐ star the repo and watch for updates!

---

## 📄 License

This project is licensed under the MIT License.

<!-- Footer Banner -->
<p align="center" style="margin-top:0;">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:764ba2,100:667eea&height=120&section=footer&width=100%"/>
</p>