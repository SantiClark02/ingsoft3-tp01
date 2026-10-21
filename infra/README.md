# infra/ — entorno preprod declarado con Terraform (TP8)

Preprod es un entorno **nuevo** (red, volumen, MySQL, API y front) que nace y muere con un comando.
QA y PROD (los `compose.*.yml` del TP6) no se tocan. Requiere Terraform ≥ 1.5 y Docker Desktop prendido.

```powershell
cd infra
terraform init        # baja los providers (versiones fijadas en el .terraform.lock.hcl)
terraform plan        # qué va a pasar, sin hacerlo
terraform apply       # crea preprod
terraform output      # api_url y front_url
curl.exe -s http://localhost:8083/health       # la API de preprod, viva
curl.exe -s http://localhost:8083/activities   # ...y hablando con SU base
terraform destroy     # preprod se va entero
```

El estado (`terraform.tfstate`) vive sólo en esta máquina y NO se commitea: tiene las contraseñas generadas.
