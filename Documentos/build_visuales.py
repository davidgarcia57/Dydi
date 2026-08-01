# -*- coding: utf-8 -*-
"""Ensambla los entregables visuales autocontenidos de la Actividad 4.2.

Toma las plantillas *.src.html y sustituye cada marcador FIG:<archivo.png> por
la imagen incrustada en base64 (data URI), de modo que el HTML resultante no
dependa de la red ni de rutas locales: se puede proyectar, imprimir o enviar
por correo como un solo archivo.

Las figuras NO se editan a mano: salen de load-tests/analyze_results.py. Si se
regeneran, basta con volver a correr este script.

Uso (desde la raíz del repositorio, sin instalar nada):
    python3 Documentos/build_visuales.py

Ademas copia a Documentos/figuras/ las tres figuras que el informe (articulo.md)
inserta como Figuras 1, 2 y 3. load-tests/analysis/ esta en .gitignore, asi que
el documento no puede depender de esa ruta.

Entradas : Documentos/**/ *.src.html  ·  load-tests/analysis/fig_*.png
Salidas  : Documentos/presentacion/dydi-defensa.html
           Documentos/divulgacion/dydi-infografia.html
           Documentos/figuras/fig{1,2,3}-*.png
"""
import base64
import os
import re
import shutil
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
FIGS = os.path.join(ROOT, "load-tests", "analysis")

TARGETS = [
    os.path.join(HERE, "presentacion", "dydi-defensa.src.html"),
    os.path.join(HERE, "divulgacion", "dydi-infografia.src.html"),
]

MARKER = re.compile(r"FIG:([A-Za-z0-9_.-]+\.png)")

# Tipografias de la marca. Van embebidas porque las piezas se imprimen a PDF en
# maquinas que no las tienen instaladas: sin esto, Newsreader cae a Georgia o a
# Liberation Serif y la pieza pierde la identidad. Son subconjuntos latinos de
# las fuentes variables de Google Fonts (licencia OFL, que permite incrustar).
FUENTES = os.path.join(HERE, "fuentes")
MARKER_FONT = re.compile(r"FONT:([A-Za-z0-9_.-]+\.woff2)")

# Figuras que articulo.md inserta en §5.2, con el nombre que usa el documento.
FIGURAS_INFORME = {
    "fig_h1_barras.png": "fig1-caidas-ws-por-nivel.png",
    "fig_h2_dumbbell.png": "fig2-ram-por-servicio.png",
    "fig_h3_linea.png": "fig3-ram-en-el-tiempo.png",
}


def data_uri(name: str) -> str:
    path = os.path.join(FIGS, name)
    with open(path, "rb") as fh:
        return "data:image/png;base64," + base64.b64encode(fh.read()).decode("ascii")


def font_uri(name: str) -> str:
    path = os.path.join(FUENTES, name)
    with open(path, "rb") as fh:
        return "data:font/woff2;base64," + base64.b64encode(fh.read()).decode("ascii")


def build(src: str) -> str:
    with open(src, encoding="utf-8") as fh:
        html = fh.read()

    missing = [n for n in set(MARKER.findall(html))
               if not os.path.exists(os.path.join(FIGS, n))]
    if missing:
        raise SystemExit(
            f"Faltan figuras para {os.path.basename(src)}: {', '.join(sorted(missing))}.\n"
            "Regenéralas con load-tests/analyze_results.py antes de ensamblar."
        )

    faltan_f = [n for n in set(MARKER_FONT.findall(html))
                if not os.path.exists(os.path.join(FUENTES, n))]
    if faltan_f:
        raise SystemExit(
            f"Faltan tipografias para {os.path.basename(src)}: {', '.join(sorted(faltan_f))}.\n"
            "Viven en Documentos/fuentes/ y estan versionadas."
        )

    cache: dict[str, str] = {}

    def repl(m: re.Match) -> str:
        name = m.group(1)
        if name not in cache:
            cache[name] = data_uri(name)
        return cache[name]

    def repl_font(m: re.Match) -> str:
        name = m.group(1)
        key = "font:" + name
        if key not in cache:
            cache[key] = font_uri(name)
        return cache[key]

    out = MARKER_FONT.sub(repl_font, MARKER.sub(repl, html))
    dst = src.replace(".src.html", ".html")
    with open(dst, "w", encoding="utf-8") as fh:
        fh.write(out)
    return dst


def sync_figuras_informe() -> None:
    """Refresca las copias versionadas que cita articulo.md."""
    dst_dir = os.path.join(HERE, "figuras")
    os.makedirs(dst_dir, exist_ok=True)
    for src_name, dst_name in FIGURAS_INFORME.items():
        src = os.path.join(FIGS, src_name)
        if not os.path.exists(src):
            print(f"[aviso] falta {src_name}; no se actualizo {dst_name}")
            continue
        shutil.copyfile(src, os.path.join(dst_dir, dst_name))
    print(f"[ok] Documentos/figuras/ ({len(FIGURAS_INFORME)} figuras del informe)")


def main() -> int:
    sync_figuras_informe()
    for src in TARGETS:
        if not os.path.exists(src):
            print(f"[omitido] no existe {os.path.relpath(src, ROOT)}")
            continue
        dst = build(src)
        size_kb = os.path.getsize(dst) / 1024
        print(f"[ok] {os.path.relpath(dst, ROOT)} ({size_kb:.0f} KB, autocontenido)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
