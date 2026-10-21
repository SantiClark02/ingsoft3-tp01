output "api_url" {
  value = "http://localhost:${docker_container.api.ports[0].external}"
}

output "front_url" {
  value = "http://localhost:${docker_container.front.ports[0].external}"
}

output "db_container" {
  value = docker_container.db.name
}

output "image" {
  value = docker_image.api.name
}
