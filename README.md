# Go Backend Starter Kit

Repositori ini adalah boilerplate backend menggunakan **Golang** dan **Gin Framework** yang dirancang dengan fokus pada skalabilitas, keamanan, dan *observability*. Proyek ini mengimplementasikan *Layered Architecture* dan *Prometheus metrics* (Opsional).

## System Architecture

![System Architecture](https://raw.githubusercontent.com/OfrenDialsa/go-backend-starter-kit/refs/heads/dev/diagram/go-starter-SysArch.png)

### Komponen Utama:

1.  **Auth API (Go/Gin):**
    * **Router Layer:** Menangani rute HTTP dan mengekspos endpoint `/metrics` untuk monitoring.
    * **Middleware Stack:** Terdiri dari Prometheus Metrics (untuk tracking hit & latency), Role Validation, Rate Limit dll.
    * **Handler, Service, & Repository:** Mengikuti *Separation of Concerns* untuk memudahkan testing dan scalability.

2.  **Monitoring Stack (Observability):**
    * **Prometheus:** Melakukan *scraping* data metrik dari API dan PostgreSQL secara berkala.
    * **Grafana:** Memvisualisasikan data metrik dalam dashboard yang interaktif untuk memantau performa sistem secara *real-time*.

3.  **Database & Storage:**
    * **PostgreSQL:** Database utama untuk menyimpan data pengguna, sesi dan log audit.

4.  **Unit Test:**
    * **Mockery:** Generate Unit test .
    * **Testify:** Jalankan test untuk memastikan fitur berjalan dengan baik.

## Fitur Utama
- **High Performance:** Dibangun dengan Golang yang efisien dalam penggunaan resource.
- **Scalable:** Siap dideploy dengan layered architecture yang matang.
- **Secure:** Dilengkapi dengan rate limiting dan validasi berlapis di tingkat middleware.
- **Observable:** Visibilitas penuh terhadap latency API dan kesehatan database.

## Tech Stack
- **Language:** Golang
- **Web Framework:** Gin Gonic
- **Database:** PostgreSQL
- **Monitoring (Opsional):** Prometheus & Grafana
- **Email Provider:** Mailgun / Mailtrap atau provider lainnya

## Cara Penggunaan
1. Clone repositori ini.
2. Jalankan `docker-compose up -d` untuk menjalankan PostgreSQL.
3. Jalankan aplikasi dengan `go run main.go`.

---
Dikembangkan oleh [Ofren Dialsa](https://github.com/OfrenDialsa)