#!/bin/sh
# Corre los tests con cobertura, arma los reportes y aplica el umbral.
# Lo ejecuta la etapa "test" del Dockerfile: misma receta en local y en CI.
UMBRAL="${UMBRAL:-15}"
OUT=/out
mkdir -p "$OUT"

# 1) Tests + cobertura de TODOS los paquetes (-coverpkg=./...)
go test ./... -coverpkg=./... -coverprofile="$OUT/coverage.raw.out" -v > "$OUT/tests.txt" 2>&1
TESTS=$?
cat "$OUT/tests.txt"

# 2) Qué NO entra en la cuenta (se excluye, así lo nuevo entra solo):
#    main.go (arranque), db/ y clients/ (conexión, migraciones, seed),
#    dao/ y domain/ (structs sin comportamiento).
grep -v -E '^backend/(main\.go|db/|clients/|dao/|domain/)' "$OUT/coverage.raw.out" > "$OUT/coverage.out"

# 3) Reportes: total por función + HTML navegable + resumen para GitHub
go tool cover -func="$OUT/coverage.out" > "$OUT/coverage.txt"
go tool cover -html="$OUT/coverage.out" -o "$OUT/coverage.html"
TOTAL=$(awk '/^total:/ {gsub("%","",$3); print $3}' "$OUT/coverage.txt")
{
  echo "### Coverage del backend (Go)"
  echo "| métrica | % | umbral |"
  echo "|---|---|---|"
  echo "| sentencias | ${TOTAL}% | ${UMBRAL}% |"
  echo "| ramas | no la mide Go (ver decisiones) | - |"
} > "$OUT/summary.md"
echo "Coverage total (sentencias): ${TOTAL}% | umbral: ${UMBRAL}%"

# 4) Gate: si fallaron tests, o la cobertura no llega al umbral, sale con error
[ "$TESTS" -ne 0 ] && echo "ERROR: hay tests en rojo" && exit 1
if awk -v t="$TOTAL" -v u="$UMBRAL" 'BEGIN { exit !(t < u) }'; then
  echo "ERROR: Coverage (statements) ${TOTAL}% does not meet threshold ${UMBRAL}%"
  exit 1
fi
