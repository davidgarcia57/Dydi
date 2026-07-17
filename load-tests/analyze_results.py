# -*- coding: utf-8 -*-
"""Análisis de resultados del experimento — banco de corridas, descriptivos y figuras.

Genera desde los artefactos crudos de results/ (nunca a mano):
  - analysis/banco_corridas.csv  (1 fila por corrida, 22 variables; el "banco"
    de la Actividad 3.4 — la copia entregada vive en Documentos/)
  - analysis/stats.json          (descriptivos por nivel + frecuencias)
  - analysis/fig_*.png           (las 6 figuras: 2 alternativas por hallazgo)

Uso (sin instalar nada local, regla del proyecto — desde la raíz del repo):
  docker run --rm -v "$(pwd)/load-tests":/lt -w /lt python:3.12-slim \
    sh -c "pip install -q matplotlib && python analyze_results.py"

O con un Python del host que tenga matplotlib:
  python load-tests/analyze_results.py [results_dir] [out_dir]

La clasificación de corridas (válida/excluida y su causa) es la fuente de
verdad del dataset y se decidió por causa asignable ANTES del análisis
(matriz.log + protocolo v2 §6.2). Si se corren nuevas corridas, agrégalas a
RUNS con su clasificación.
"""
import csv, json, os, re, statistics, sys
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
from matplotlib.lines import Line2D

HERE = os.path.dirname(os.path.abspath(__file__))
RESULTS = sys.argv[1] if len(sys.argv) > 1 else os.path.join(HERE, "results")
OUT = sys.argv[2] if len(sys.argv) > 2 else os.path.join(HERE, "analysis")
CSV_OUT = os.path.join(OUT, "banco_corridas.csv")
os.makedirs(OUT, exist_ok=True)

# ---------- clasificación de corridas (matriz.log + protocolo v2 §6.2) ----------
RUNS = [
    ("20260702-192048-peak100-rep1", "PILOTO", 0,
     "Artefacto del instrumento: limite de tasa por usuario compartido (se midio el limitador, no la arquitectura)"),
    ("20260702-195005-peak100-rep1", "PILOTO", 1, ""),
    ("20260704-200841-peak100-rep1", "S1", 1, ""),
    ("20260704-202815-peak100-rep2", "S1", 1, ""),
    ("20260704-204754-peak100-rep3", "S1", 1, ""),
    ("20260704-210729-peak1000-rep1", "S1", 1, ""),
    ("20260704-212704-peak1000-rep2", "S1", 1, ""),
    ("20260704-214639-peak1000-rep3", "S1", 1, ""),
    ("20260713-190612-peak1000-rep1", "S2", 0,
     "Interferencia de configuracion: pooler de BD en modo sesion (5432) satura el pool de groups"),
    ("20260713-193625-peak1000-rep1", "S2", 0,
     "Fallo del inyector: error de I/O de Docker en el host; k6 nunca ejecuto (sin datos k6)"),
    ("20260713-195531-peak1000-rep1", "S2", 0,
     "Estado no representativo de la capa de datos: creditos de rafaga agotados (t3a.nano) tras la pausa del proyecto"),
]

SERVICES = ["gateway", "groups", "habits", "realtime"]
MIB = 1024 * 1024  # el articulo y Render reportan MiB binarios etiquetados como MB

def dur_ms(s):
    """'954ms' | '2.56s' | '1m3s' -> ms"""
    s = s.strip()
    m = re.fullmatch(r"(?:(\d+)m)?([\d.]+)(m?s)", s)
    if not m:
        return None
    mins = int(m.group(1) or 0)
    val = float(m.group(2))
    ms = val if m.group(3) == "ms" else val * 1000
    return mins * 60000 + ms

def parse_k6_text(path):
    """Extrae metricas del resumen textual de k6 (corridas sin summary.json)."""
    out = {}
    txt = open(path, encoding="utf-8", errors="replace").read()
    m = re.search(r"http_req_duration\.+:.*?med=(\S+).*?p\(95\)=(\S+)", txt)
    if m:
        out["http_p50_ms"] = dur_ms(m.group(1)); out["http_p95_ms"] = dur_ms(m.group(2))
    m = re.search(r"http_req_failed\.+: ([\d.]+)% (\d+) out of (\d+)", txt)
    if m:
        out["http_fail_pct"] = float(m.group(1)); out["http_reqs"] = int(m.group(3))
    m = re.search(r"ws_dropped_rate\.+: ([\d.]+)% (\d+) out of (\d+)", txt)
    if m:
        out["ws_drop_pct"] = float(m.group(1)); out["ws_sessions"] = int(m.group(3))
    m = re.search(r"ws_connect_time\.+:.*?p\(95\)=(\S+)", txt)
    if m:
        out["ws_conn_p95_ms"] = dur_ms(m.group(1))
    m = re.search(r"vus_max\.+: (\d+)", txt)
    if m:
        out["vus_max"] = int(m.group(1))
    m = re.search(r"data_received\.+: ([\d.]+) ([kMG]B)", txt)
    if m:
        mult = {"kB": 1e3, "MB": 1e6, "GB": 1e9}[m.group(2)]
        out["datos_recibidos_mb"] = float(m.group(1)) * mult / 1e6
    return out

def parse_summary(path):
    d = json.load(open(path))["metrics"]
    return {
        "http_reqs": d["http_reqs"]["count"],
        "http_fail_pct": round(d["http_req_failed"]["value"] * 100, 2),
        "http_p50_ms": round(d["http_req_duration"]["med"], 1),
        "http_p95_ms": round(d["http_req_duration"]["p(95)"], 1),
        "ws_sessions": d["ws_sessions"]["count"],
        "ws_drop_pct": round(d["ws_dropped_rate"]["value"] * 100, 2),
        "ws_conn_p95_ms": round(d["ws_connect_time"]["p(95)"], 1),
        "vus_max": d["vus_max"]["value"],
        "datos_recibidos_mb": round(d["data_received"]["count"] / 1e6, 1),
    }

def ram_peaks(path):
    """max process_resident_memory_bytes por servicio, en MB (MiB)."""
    peaks = {s: None for s in SERVICES}
    with open(path) as f:
        for row in csv.DictReader(f):
            if row["metric"] != "process_resident_memory_bytes":
                continue
            v = float(row["value"]) / MIB
            s = row["service"]
            if peaks.get(s) is None or v > peaks[s]:
                peaks[s] = v
    return {s: (round(v, 1) if v is not None else None) for s, v in peaks.items()}

def ram_series(path):
    """serie de tiempo RSS (MB) por servicio: [(min_desde_inicio, mb)]"""
    series = {s: [] for s in SERVICES}
    t0 = None
    with open(path) as f:
        for row in csv.DictReader(f):
            if row["metric"] != "process_resident_memory_bytes":
                continue
            t = int(row["unix_ts"])
            if t0 is None:
                t0 = t
            series[row["service"]].append(((t - t0) / 60.0, float(row["value"]) / MIB))
    return series

# ---------- construir banco ----------
bank = []
for run_id, sesion, valida, motivo in RUNS:
    d = os.path.join(RESULTS, run_id)
    meta = json.load(open(os.path.join(d, "metadata.json"))) if os.path.exists(os.path.join(d, "metadata.json")) else {}
    m = re.search(r"peak(\d+)-rep(\d+)", run_id)
    row = {
        "id_corrida": run_id,
        "fecha_hora_utc": meta.get("started_at", ""),
        "sesion": sesion,
        "commit": meta.get("commit", ""),
        "nivel_vus": int(m.group(1)),
        "repeticion": int(m.group(2)),
        "valida": valida,
        "motivo_exclusion": motivo,
    }
    sj = os.path.join(d, "summary.json")
    kt = os.path.join(d, "k6_output.txt")
    fuente = ""
    if os.path.exists(sj):
        row.update(parse_summary(sj)); fuente = "summary.json"
    elif os.path.exists(kt) and os.path.getsize(kt) > 1000:
        parsed = parse_k6_text(kt)
        if parsed:
            row.update(parsed); fuente = "k6_output.txt (texto)"
    row["fuente_k6"] = fuente or "sin datos k6"
    mc = os.path.join(d, "metrics.csv")
    if os.path.exists(mc) and os.path.getsize(mc) > 200:
        for s, v in ram_peaks(mc).items():
            row[f"ram_pico_{s}_mb"] = v
    bank.append(row)

COLS = ["id_corrida", "fecha_hora_utc", "sesion", "commit", "nivel_vus", "repeticion",
        "valida", "motivo_exclusion", "fuente_k6", "vus_max", "http_reqs", "http_fail_pct",
        "http_p50_ms", "http_p95_ms", "ws_sessions", "ws_drop_pct", "ws_conn_p95_ms",
        "datos_recibidos_mb", "ram_pico_gateway_mb", "ram_pico_groups_mb",
        "ram_pico_habits_mb", "ram_pico_realtime_mb"]
with open(CSV_OUT, "w", newline="", encoding="utf-8-sig") as f:
    w = csv.DictWriter(f, fieldnames=COLS, extrasaction="ignore")
    w.writeheader()
    for r in bank:
        w.writerow({k: ("NA" if r.get(k) in (None, "") and k not in ("motivo_exclusion",) else r.get(k, "NA")) for k in COLS})

# ---------- estadistica descriptiva (solo matriz S1 valida) ----------
s1 = [r for r in bank if r["sesion"] == "S1" and r["valida"] == 1]
NUMVARS = ["http_p95_ms", "ws_drop_pct", "ws_conn_p95_ms",
           "ram_pico_gateway_mb", "ram_pico_groups_mb", "ram_pico_habits_mb", "ram_pico_realtime_mb"]

def describe(vals):
    return {
        "n": len(vals),
        "media": round(statistics.mean(vals), 2),
        "mediana": round(statistics.median(vals), 2),
        "min": round(min(vals), 2),
        "max": round(max(vals), 2),
        "desv_est": round(statistics.stdev(vals), 2) if len(vals) > 1 else None,
    }

desc = {}
for nivel in (100, 1000):
    sub = [r for r in s1 if r["nivel_vus"] == nivel]
    desc[str(nivel)] = {v: describe([r[v] for r in sub if r.get(v) is not None]) for v in NUMVARS}

freq = {"por_nivel": {}, "por_sesion": {}, "por_estado": {}}
for r in bank:
    freq["por_nivel"][str(r["nivel_vus"])] = freq["por_nivel"].get(str(r["nivel_vus"]), 0) + 1
    freq["por_sesion"][r["sesion"]] = freq["por_sesion"].get(r["sesion"], 0) + 1
    k = "valida" if r["valida"] else "excluida"
    freq["por_estado"][k] = freq["por_estado"].get(k, 0) + 1

# ---------- graficos ----------
INK, SEC, MUTED = "#0b0b0b", "#52514e", "#898781"
GRID, BASE = "#e1e0d9", "#c3c2b7"
BLUE, BLUE_L, AQUA, DEEMPH = "#2a78d6", "#86b6ef", "#1baf7a", "#c3c2b7"
from matplotlib import font_manager
_fams = {f.name for f in font_manager.fontManager.ttflist}
FONT = ["Segoe UI", "DejaVu Sans"] if "Segoe UI" in _fams else ["DejaVu Sans"]
plt.rcParams.update({
    "font.family": FONT,
    "figure.facecolor": "white", "axes.facecolor": "white",
    "axes.edgecolor": BASE, "axes.labelcolor": SEC,
    "xtick.color": MUTED, "ytick.color": MUTED,
    "xtick.labelsize": 9, "ytick.labelsize": 9,
    "axes.titlesize": 12, "axes.titleweight": "semibold", "axes.titlecolor": INK,
    "axes.spines.top": False, "axes.spines.right": False,
    "svg.fonttype": "none",
})

def style_ax(ax, ygrid=True):
    if ygrid:
        ax.grid(axis="y", color=GRID, linewidth=0.8)
        ax.set_axisbelow(True)
    ax.spines["left"].set_visible(False)
    ax.spines["bottom"].set_color(BASE)
    ax.tick_params(length=0)

def miles(n):  # 4 634 con espacio fino, como el articulo
    return f"{n:,}".replace(",", " ")

drop_by_level = {n: sorted(r["ws_drop_pct"] for r in s1 if r["nivel_vus"] == n) for n in (100, 1000)}

# --- Fig 1A: barras ws_drop por nivel + umbral (GANADOR H1) ---
fig, ax = plt.subplots(figsize=(6.4, 3.7), dpi=200)
levels = [100, 1000]
med = [statistics.median(drop_by_level[n]) for n in levels]
x = [0, 1]
ax.bar(x, med, width=0.42, color=BLUE, zorder=3)
for xi, n in zip(x, levels):  # repeticiones individuales
    ax.scatter([xi + 0.29] * 3, drop_by_level[n], s=22, color=INK, zorder=4)
ax.axhline(10, color=MUTED, linewidth=1.2, linestyle=(0, (5, 4)), zorder=2)
ax.text(2.05, 10.5, "umbral QoS: 10 %", va="bottom", ha="right", fontsize=9, color=SEC)
for xi, v in zip(x, med):
    ax.text(xi, v + 0.7, f"{v:.2f} %", ha="center", fontsize=10.5, color=INK, fontweight="semibold")
ax.set_xticks(x); ax.set_xticklabels(["100 VUs", "1 000 VUs"], fontsize=10, color=SEC)
ax.set_xlim(-0.5, 2.1); ax.set_ylim(0, 28)
ax.set_ylabel("Conexiones WS caídas (%)", fontsize=9.5)
ax.set_title("Conexiones WebSocket caídas por nivel de carga", pad=24, loc="left")
ax.text(0, 1.03, "Barra = mediana de 3 repeticiones · puntos = repeticiones individuales (Sesión 1)",
        transform=ax.transAxes, fontsize=9, color=SEC)
style_ax(ax)
fig.tight_layout(); fig.savefig(os.path.join(OUT, "fig_h1_barras.png")); plt.close(fig)

# --- Fig 1B: pastel (alternativa descartada H1) ---
fig, ax = plt.subplots(figsize=(6.4, 3.7), dpi=200)
run = next(r for r in s1 if r["nivel_vus"] == 1000 and r["repeticion"] == 1)
caidas = round(run["ws_sessions"] * run["ws_drop_pct"] / 100)
ok = run["ws_sessions"] - caidas
wedges, _ = ax.pie([ok, caidas], colors=[BLUE, "#d03b3b"], startangle=90,
                   wedgeprops=dict(width=0.42, edgecolor="white", linewidth=2))
ax.text(0, 0, f"{miles(run['ws_sessions'])}\nsesiones", ha="center", va="center", fontsize=10, color=SEC)
ax.legend(wedges, [f"Completadas · {miles(ok)} ({100-run['ws_drop_pct']:.1f} %)",
                   f"Caídas · {miles(caidas)} ({run['ws_drop_pct']:.1f} %)"],
          loc="center left", bbox_to_anchor=(1.02, 0.5), frameon=False, fontsize=9.5, labelcolor=SEC)
ax.set_title("Composición de sesiones WS a 1 000 VUs (rep. 1)", pad=14, loc="left")
fig.tight_layout(); fig.savefig(os.path.join(OUT, "fig_h1_pastel.png")); plt.close(fig)

# --- datos H2: RAM mediana por servicio y nivel ---
ram_med = {}
for svc in SERVICES:
    ram_med[svc] = {n: statistics.median([r[f"ram_pico_{svc}_mb"] for r in s1 if r["nivel_vus"] == n]) for n in (100, 1000)}
order = ["gateway", "realtime", "groups", "habits"]
NAMES = {"gateway": "api-gateway", "realtime": "realtime", "groups": "groups", "habits": "habits"}

# --- Fig 2A: dumbbell RAM (GANADOR H2) ---
fig, ax = plt.subplots(figsize=(6.4, 3.9), dpi=200)
ys = list(range(len(order)))[::-1]
for y, svc in zip(ys, order):
    a, b = ram_med[svc][100], ram_med[svc][1000]
    ax.plot([a, b], [y, y], color=BASE, linewidth=2, zorder=2)
    ax.scatter([a], [y], s=64, color=BLUE_L, zorder=3, edgecolors="white", linewidths=1)
    ax.scatter([b], [y], s=64, color=BLUE, zorder=3, edgecolors="white", linewidths=1)
    ax.text(b + 12, y, f"{b:.0f} MB", va="center", fontsize=9.5, color=INK, fontweight="semibold")
    ax.text(min(a, b) - 12, y, NAMES[svc], va="center", ha="right", fontsize=10, color=SEC)
ax.axvline(512, color=MUTED, linewidth=1.2, linestyle=(0, (5, 4)), zorder=1)
ax.text(500, -0.42, "límite Render: 512 MB ", fontsize=9, color=SEC, ha="right", va="center")
ax.set_yticks([]); ax.set_xlim(-90, 560); ax.set_ylim(-0.6, len(order) - 0.1)
ax.set_xlabel("RAM pico (MB, mediana de 3 repeticiones)", fontsize=9.5)
ax.set_title("RAM pico por servicio: 100 → 1 000 VUs", pad=24, loc="left")
leg = [Line2D([0], [0], marker="o", linestyle="", markersize=8, markerfacecolor=BLUE_L, markeredgecolor="white", label="100 VUs"),
       Line2D([0], [0], marker="o", linestyle="", markersize=8, markerfacecolor=BLUE, markeredgecolor="white", label="1 000 VUs")]
ax.legend(handles=leg, loc="lower center", bbox_to_anchor=(0.58, 0.0), frameon=False, fontsize=9.5, labelcolor=SEC)
ax.grid(axis="x", color=GRID, linewidth=0.8); ax.set_axisbelow(True)
ax.spines["left"].set_visible(False); ax.spines["bottom"].set_color(BASE); ax.tick_params(length=0)
fig.tight_layout(); fig.savefig(os.path.join(OUT, "fig_h2_dumbbell.png")); plt.close(fig)

# --- Fig 2B: barras agrupadas RAM (alternativa H2) ---
fig, ax = plt.subplots(figsize=(6.4, 3.9), dpi=200)
xs = list(range(len(order))); w = 0.34
v100 = [ram_med[s][100] for s in order]; v1000 = [ram_med[s][1000] for s in order]
ax.bar([i - w / 2 for i in xs], v100, width=w - 0.03, color=BLUE_L, zorder=3, label="100 VUs")
ax.bar([i + w / 2 for i in xs], v1000, width=w - 0.03, color=BLUE, zorder=3, label="1 000 VUs")
for i, (a, b) in enumerate(zip(v100, v1000)):
    ax.text(i - w / 2, a + 8, f"{a:.0f}", ha="center", fontsize=8.5, color=SEC)
    ax.text(i + w / 2, b + 8, f"{b:.0f}", ha="center", fontsize=8.5, color=INK, fontweight="semibold")
ax.axhline(512, color=MUTED, linewidth=1.2, linestyle=(0, (5, 4)), zorder=2)
ax.text(-0.4, 522, "límite Render: 512 MB", fontsize=9, color=SEC, ha="left")
ax.set_xticks(xs); ax.set_xticklabels([NAMES[s] for s in order], fontsize=10, color=SEC)
ax.set_ylim(0, 580); ax.set_ylabel("RAM pico (MB)", fontsize=9.5)
ax.set_title("RAM pico por servicio y nivel de carga", pad=24, loc="left")
ax.text(0, 1.03, "Mediana de 3 repeticiones (Sesión 1)", transform=ax.transAxes, fontsize=9, color=SEC)
ax.legend(loc="upper right", bbox_to_anchor=(1, 0.82), frameon=False, fontsize=9.5, labelcolor=SEC)
style_ax(ax)
fig.tight_layout(); fig.savefig(os.path.join(OUT, "fig_h2_barras.png")); plt.close(fig)

# --- Fig 3A: linea serie de tiempo RAM (GANADOR H3) ---
series = ram_series(os.path.join(RESULTS, "20260704-210729-peak1000-rep1", "metrics.csv"))
fig, ax = plt.subplots(figsize=(6.4, 3.9), dpi=200)
colors = {"gateway": BLUE, "realtime": AQUA, "groups": DEEMPH, "habits": DEEMPH}
for svc in ["groups", "habits", "gateway", "realtime"]:
    pts = series[svc]
    ax.plot([p[0] for p in pts], [p[1] for p in pts], color=colors[svc], linewidth=2, zorder=3)
    lx, ly = pts[-1][0] + 0.15, pts[-1][1]
    if svc == "habits":
        ly -= 14
    ax.text(lx, ly, NAMES[svc], va="center", fontsize=9.5, fontweight="semibold",
            color={"gateway": BLUE, "realtime": "#0e8a5f", "groups": MUTED, "habits": MUTED}[svc])
ax.axhline(512, color=MUTED, linewidth=1.2, linestyle=(0, (5, 4)), zorder=2)
ax.text(0.1, 522, "límite Render: 512 MB", fontsize=9, color=SEC)
ax.set_ylim(0, 570); ax.set_xlim(0, max(p[0] for p in series["gateway"]) + 1.6)
ax.set_xlabel("Minutos desde el inicio de la corrida", fontsize=9.5)
ax.set_ylabel("RAM (MB)", fontsize=9.5)
ax.set_title("RAM durante la corrida de 1 000 VUs (rep. 1)", pad=24, loc="left")
ax.text(0, 1.03, "Muestreo cada 5 s vía /metrics · la rampa de conexiones dura ~9 min", transform=ax.transAxes, fontsize=9, color=SEC)
style_ax(ax)
fig.tight_layout(); fig.savefig(os.path.join(OUT, "fig_h3_linea.png")); plt.close(fig)

# --- Fig 3B: barras de picos (alternativa H3) ---
fig, ax = plt.subplots(figsize=(6.4, 3.9), dpi=200)
peaks = {s: max(p[1] for p in series[s]) for s in SERVICES}
xs = list(range(len(order)))
vals = [peaks[s] for s in order]
ax.bar(xs, vals, width=0.42, color=[colors[s] for s in order], zorder=3)
for i, v in enumerate(vals):
    ax.text(i, v + 8, f"{v:.0f}", ha="center", fontsize=9.5, color=INK, fontweight="semibold")
ax.axhline(512, color=MUTED, linewidth=1.2, linestyle=(0, (5, 4)), zorder=2)
ax.text(len(order) - 0.55, 520, "límite Render: 512 MB", fontsize=9, color=SEC, ha="right")
ax.set_xticks(xs); ax.set_xticklabels([NAMES[s] for s in order], fontsize=10, color=SEC)
ax.set_ylim(0, 580); ax.set_ylabel("RAM pico (MB)", fontsize=9.5)
ax.set_title("Solo el pico: RAM máxima en la corrida de 1 000 VUs (rep. 1)", pad=24, loc="left")
ax.text(0, 1.03, "La misma corrida que la figura de líneas, reducida a un punto por servicio", transform=ax.transAxes, fontsize=9, color=SEC)
style_ax(ax)
fig.tight_layout(); fig.savefig(os.path.join(OUT, "fig_h3_barras.png")); plt.close(fig)

# ---------- stats.json (insumo de tablas/documentos) ----------
json.dump({
    "bank": bank, "desc": desc, "freq": freq,
    "ram_med": ram_med,
    "drop_by_level": drop_by_level,
    "pastel_run": {"sesiones": run["ws_sessions"], "caidas": caidas, "ok": ok},
}, open(os.path.join(OUT, "stats.json"), "w", encoding="utf-8"), ensure_ascii=False, indent=1)

print("Banco:", CSV_OUT, f"({len(bank)} corridas, {len(COLS)} variables)")
print("Validas S1:", len(s1), "| Figuras:", sorted(f for f in os.listdir(OUT) if f.endswith(".png")))
