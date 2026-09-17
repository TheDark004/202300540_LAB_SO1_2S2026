#!/bin/bash
# Mantener siempre tres contenedores bajos y dos de alto consumo.
for i in 1 2 3; do
    docker run -d --label proyecto2=carga --label proyecto2_categoria=bajo alpine sleep 240
done

docker run -d --label proyecto2=carga --label proyecto2_categoria=alto roldyoran/go-client
docker run -d --label proyecto2=carga --label proyecto2_categoria=alto alpine sh -c "while true; do echo '2^20' | bc > /dev/null; sleep 2; done"

# Contenedor intruso ocasional
if (( RANDOM % 3 == 0 )); then
    docker run -d --label proyecto2=carga --label proyecto2_categoria=intruso intruso-202300540
fi