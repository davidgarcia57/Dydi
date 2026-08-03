#!/usr/bin/env sh
# Exporta los .md de Documentos/ a .docx con pandoc.
#
# Va por Docker porque pandoc no esta instalado en WSL (regla del proyecto: nada
# de toolchains locales). Corre desde WSL:
#
#   ./Documentos/export.sh                 # los tres documentos
#   ./Documentos/export.sh articulo.md     # solo uno
#
# El contenedor trabaja con cwd=/doc, de modo que las rutas relativas de las
# imagenes (figuras/fig1-*.png) resuelven igual que al leer el .md.
#
# Formato de figuras y tablas: el orden APA 7 (numero en negritas, titulo en
# cursiva, contenido, *Nota.*) se escribe a mano en el .md con parrafos, y las
# imagenes van con alt VACIO -- ![](ruta.png) -- para que pandoc NO genere su
# propio caption debajo. Por lo mismo no se usa la sintaxis nativa de caption de
# tabla (": Titulo"): duplicaria el titulo, una vez como parrafo y otra con
# estilo TableCaption.
#
# reference.docx fija la maqueta (Carta, margenes de 2.5 cm) y los estilos. Es el
# reference.docx por defecto de pandoc con la geometria de pagina inyectada en su
# sectPr, para que el resultado no dependa de la configuracion del Word que abra
# el archivo. Sin el, pandoc no emite pgSz/pgMar y cada lector aplica su default.
set -eu

DIR=$(cd "$(dirname "$0")" && pwd)
IMAGE=pandoc/core:3.10
REF=reference.docx

# El articulo se entrega con el nombre del informe final, no con el del fuente.
out_name() {
	case "$1" in
	articulo.md) echo "Informe final - Dydi (articulo).docx" ;;
	*) echo "${1%.md}.docx" ;;
	esac
}

if [ "$#" -eq 0 ]; then
	set -- "articulo.md" \
		"Actividad 4.1 - Informe final y comunicacion visual.md" \
		"Actividad 4.2 - Presentacion oral y divulgacion.md"
fi

for md in "$@"; do
	[ -f "$DIR/$md" ] || {
		echo "no existe: $md" >&2
		exit 1
	}
	out=$(out_name "$md")
	printf '%s  ->  %s\n' "$md" "$out"
	docker run --rm -v "$DIR":/doc -w /doc "$IMAGE" \
		"$md" -o "$out" --reference-doc="$REF"
done
