#!/usr/bin/env sh
# Exporta el deck de la defensa y la infografia a PDF, con Chromium headless en
# Docker (no hay navegador ni toolchain local: regla del proyecto).
#
#   ./Documentos/export-visuales.sh
#
# Requisito previo: los .html tienen que estar armados desde las plantillas, o el
# PDF sale con los marcadores FIG: sin sustituir:
#
#   python3 Documentos/build_visuales.py
#
# El tamano de hoja NO se pasa por linea de comandos, porque cada pieza lo declara
# en su propio CSS y Chromium lo respeta: el deck usa @page 297x167 mm (16:9, una
# lamina por pagina) y la infografia @page A4 vertical. Si el PDF sale en Carta,
# lo que se rompio es el @page de la plantilla.
#
# Nota de tipografia: Newsreader y Hanken Grotesk no van embebidas en el HTML, son
# pilas con alternativas del sistema. Dentro del contenedor caen a Liberation
# Serif y Liberation Sans, que es exactamente lo que ya embebian los PDF
# anteriores del repositorio, asi que el resultado es comparable con ellos. Para
# obtener el render con las tipografias de la marca habria que incrustarlas como
# @font-face en base64 en los .src.html.
set -eu

DIR=$(cd "$(dirname "$0")" && pwd)
IMAGE=debian:12-slim
PIEZAS="presentacion/dydi-defensa divulgacion/dydi-infografia"

for p in $PIEZAS; do
	[ -f "$DIR/$p.html" ] || {
		echo "falta $p.html — corre antes: python3 Documentos/build_visuales.py" >&2
		exit 1
	}
	if grep -q 'src="FIG:' "$DIR/$p.html"; then
		echo "$p.html tiene marcadores FIG: sin sustituir — corre build_visuales.py" >&2
		exit 1
	fi
done

# apt necesita root dentro del contenedor, asi que el PDF nace de root: al final se
# le devuelve el dueno del host para que no haga falta sudo para volver a exportar.
docker run --rm -v "$DIR":/doc -w /doc \
	-e OWNER="$(id -u):$(id -g)" "$IMAGE" sh -eu -c '
  apt-get update -qq >/dev/null
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
    chromium fonts-liberation fonts-wqy-zenhei >/dev/null
  for p in '"$PIEZAS"'; do
    chromium --headless --no-sandbox --disable-gpu --hide-scrollbars \
      --no-pdf-header-footer --virtual-time-budget=10000 \
      --print-to-pdf="/doc/$p.pdf" "file:///doc/$p.html" 2>/dev/null
    chown "$OWNER" "/doc/$p.pdf"
    printf "  %s.pdf\n" "$p"
  done
'
