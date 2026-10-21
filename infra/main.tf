# =====================================================================
# TP8 — Infraestructura como Código (plan C de la cátedra, §3.7)
# Entorno NUEVO "preprod": red + volumen + MySQL + API + front, como
# contenedores Docker locales. QA y PROD (compose del TP6) NO se tocan.
# =====================================================================

terraform {
  required_version = ">= 1.5"

  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 4.0" # fijada; el .terraform.lock.hcl congela la versión exacta y su hash
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
  # Sin bloque backend: el estado queda en infra/terraform.tfstate, en MI máquina,
  # y el .gitignore lo deja afuera del repo (tiene secretos adentro).
}

provider "docker" {
  host = var.docker_host # Docker Desktop en Windows
}

# --- Secretos: no se escriben en ningún archivo, los GENERA Terraform ---------
resource "random_password" "db" {
  length  = 20
  special = false
}

resource "random_password" "jwt" {
  length  = 32
  special = false
}

# --- Red y volumen propios de preprod -----------------------------------------
resource "docker_network" "preprod" {
  name = "gimnasio-preprod"
}

resource "docker_volume" "db_data" {
  name = "gimnasio-preprod-mysql"
}

# --- Imágenes: la MISMA imagen del TP7 (sha-<commit>), no se construye nada ----
resource "docker_image" "mysql" {
  name         = "mysql:8.0"
  keep_locally = true # QA y PROD también la usan: destroy no la borra
}

resource "docker_image" "api" {
  name         = "${var.image_repo_api}:${var.image_tag}"
  keep_locally = true
}

resource "docker_image" "front" {
  name         = "${var.image_repo_front}:${var.image_tag}"
  keep_locally = true
}

# --- Base de datos -------------------------------------------------------------
resource "docker_container" "db" {
  name    = "gimnasio-preprod-db"
  image   = docker_image.mysql.image_id
  restart = "unless-stopped"

  env = [
    "MYSQL_ROOT_PASSWORD=${random_password.db.result}",
    "MYSQL_DATABASE=backend",
    "MYSQL_USER=gimnasio_user",
    "MYSQL_PASSWORD=${random_password.db.result}",
  ]

  networks_advanced {
    name = docker_network.preprod.name
  }

  volumes {
    volume_name    = docker_volume.db_data.name
    container_path = "/var/lib/mysql"
  }

  # "Arrancó" no es "está lista" (TP2): la API se crea recién cuando MySQL
  # acepta conexiones TCP. Es el equivalente del service_healthy de compose.
  healthcheck {
    test         = ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 --protocol=tcp -uroot -p$MYSQL_ROOT_PASSWORD --silent"]
    interval     = "5s"
    timeout      = "5s"
    start_period = "20s"
    retries      = 30
  }
  wait         = true
  wait_timeout = 240
}

# --- API (Go) ------------------------------------------------------------------
resource "docker_container" "api" {
  name    = "gimnasio-preprod-api"
  image   = docker_image.api.image_id
  restart = "unless-stopped"

  env = [
    "DB_HOST=${docker_container.db.name}", # DERIVADO: el nombre del contenedor de SU base
    "DB_PORT=3306",
    "DB_USER=gimnasio_user",
    "DB_PASSWORD=${random_password.db.result}",
    "DB_NAME=backend",
    "JWT_SECRET=${random_password.jwt.result}",
    "APP_COMMIT=${trimprefix(var.image_tag, "sha-")}", # /health dice qué imagen corre
  ]

  networks_advanced {
    name = docker_network.preprod.name
  }

  ports {
    internal = 8080
    external = var.api_port
  }
}

# --- Front (nginx) -------------------------------------------------------------
resource "docker_container" "front" {
  name    = "gimnasio-preprod-front"
  image   = docker_image.front.image_id
  restart = "unless-stopped"

  env = [
    "BACKEND_URL=http://${docker_container.api.name}:8080", # DERIVADO: la API de ESTE entorno
    "DNS_RESOLVER=127.0.0.11",                              # DNS interno de Docker
  ]

  networks_advanced {
    name = docker_network.preprod.name
  }

  ports {
    internal = 80
    external = var.front_port
  }
}
