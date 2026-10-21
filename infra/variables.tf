# Nada de esto es secreto: una imagen pública, una etiqueta y puertos.

variable "image_repo_api" {
  description = "Imagen del backend en el registry, SIN etiqueta"
  type        = string
  default     = "ghcr.io/santiclark02/ingsoft3-tp01-backend"
}

variable "image_repo_front" {
  description = "Imagen del frontend en el registry, SIN etiqueta"
  type        = string
  default     = "ghcr.io/santiclark02/ingsoft3-tp01-frontend"
}

variable "image_tag" {
  description = "sha-<commit> de un merge que publicó imagen (TP7). Hoy: el de la release v7.0.0"
  type        = string
  default     = "sha-71e53f7134d13cc5ef4092c75ca3e647be3c2852"
}

variable "api_port" {
  description = "Puerto del host para la API (QA usa 8081 y PROD 8082)"
  type        = number
  default     = 8083
}

variable "front_port" {
  description = "Puerto del host para el front (QA usa 3001 y PROD 3002)"
  type        = number
  default     = 3003
}

variable "docker_host" {
  description = "Dónde escucha Docker. Docker Desktop en Windows: el named pipe"
  type        = string
  default     = "npipe:////./pipe/docker_engine"
}
