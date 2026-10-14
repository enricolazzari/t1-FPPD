#!/usr/bin/env python3
"""Lê results/todas.csv e gera results/tabela.md + gráficos PNG.

Uso: python3 scripts/report.py   (precisa de matplotlib: pip3 install matplotlib)
Funciona com qualquer subconjunto de versões (ex.: só a V3).
"""
import csv
import os
from collections import defaultdict

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
RES = os.path.join(ROOT, "results")

# cor fixa por solução (a mesma em todos os gráficos)
COLORS = {"v0": "#2a78d6", "v1": "#e34948", "v2": "#eb6834",
          "v3/monitor": "#1baf7a", "v3/ordered": "#4a3aa7"}
INK, MUTED, GRID = "#0b0b0b", "#52514e", "#e4e3df"


def load():
    groups = defaultdict(list)
    with open(os.path.join(RES, "todas.csv")) as f:
        for r in csv.DictReader(f):
            key = r["versao"] if r["versao"] != "v3" else f"v3/{r['estrategia']}"
            groups[key].append(r)
    return dict(sorted(groups.items()))


def col(rows, name, conv=float):
    return [conv(r[name]) for r in rows]


def mean(xs):
    return sum(xs) / len(xs)


def table(groups):
    lines = [
        "| Solução | Estratégia | Execuções | Tempo total médio (s) | Tempo total máx. (s) "
        "| Espera média (s) | Maior espera (s) | Máx. comendo | Ciclos por filósofo |",
        "|---|---|---|---|---|---|---|---|---|",
    ]
    for key, rows in groups.items():
        t = col(rows, "tempo_total_ms")
        ciclos = sorted({tuple(r[f"ciclos_f{i}"] for i in range(5)) for r in rows})
        ciclos_txt = " / ".join("[" + ",".join(c) + "]" for c in ciclos)
        lines.append(
            f"| {key.split('/')[0].upper()} | {rows[0]['estrategia']} | {len(rows)} "
            f"| {mean(t) / 1000:.2f} | {max(t) / 1000:.2f} "
            f"| {mean(col(rows, 'espera_media_ms')) / 1000:.3f} "
            f"| {max(col(rows, 'espera_max_ms')) / 1000:.3f} "
            f"| {max(col(rows, 'max_comendo', int))} | {ciclos_txt} |"
        )
    lines.append("")
    lines.append("Médias e máximos sobre as execuções de cada solução. "
                 "\"Maior espera\" = maior valor observado em todas as execuções.")
    return "\n".join(lines) + "\n"


def style(ax, title, ylabel):
    ax.set_title(title, loc="left", color=INK, fontsize=11)
    ax.set_ylabel(ylabel, color=MUTED, fontsize=9)
    ax.tick_params(colors=MUTED, labelsize=9, length=0)
    ax.grid(axis="y", color=GRID, linewidth=0.8)
    ax.set_axisbelow(True)
    for s in ("top", "right", "left"):
        ax.spines[s].set_visible(False)
    ax.spines["bottom"].set_color(GRID)


def bars(ax, keys, values, fmt):
    colors = [COLORS.get(k, "#888888") for k in keys]
    b = ax.bar(keys, values, color=colors, width=0.6, edgecolor="white", linewidth=2)
    ax.bar_label(b, labels=[fmt(v) for v in values], padding=3, color=INK, fontsize=9)
    ax.margins(y=0.15)


def chart_total(groups, keys):
    fig, ax = plt.subplots(figsize=(6.4, 3.6), dpi=150)
    bars(ax, keys, [mean(col(groups[k], "tempo_total_ms")) / 1000 for k in keys],
         lambda v: f"{v:.1f} s")
    style(ax, "Tempo total de execução (média das execuções)", "segundos")
    fig.tight_layout()
    fig.savefig(os.path.join(RES, "tempo_total.png"))
    plt.close(fig)


def chart_wait(groups, keys):
    fig, (a1, a2) = plt.subplots(1, 2, figsize=(9, 3.6), dpi=150)
    bars(a1, keys, [mean(col(groups[k], "espera_media_ms")) / 1000 for k in keys],
         lambda v: f"{v:.2f} s")
    style(a1, "Espera média pelos garfos", "segundos")
    bars(a2, keys, [max(col(groups[k], "espera_max_ms")) / 1000 for k in keys],
         lambda v: f"{v:.2f} s")
    style(a2, "Maior espera observada", "segundos")
    fig.tight_layout()
    fig.savefig(os.path.join(RES, "espera.png"))
    plt.close(fig)


def chart_eating(groups, keys):
    fig, ax = plt.subplots(figsize=(6.4, 3.6), dpi=150)
    bars(ax, keys, [max(col(groups[k], "max_comendo", int)) for k in keys], str)
    ax.set_yticks(range(0, 4))
    style(ax, "Máximo de filósofos comendo ao mesmo tempo", "filósofos")
    fig.tight_layout()
    fig.savefig(os.path.join(RES, "max_comendo.png"))
    plt.close(fig)


def main():
    groups = load()
    if not groups:
        print("results/todas.csv está vazio — rode scripts/run_experiments.sh antes")
        return
    with open(os.path.join(RES, "tabela.md"), "w") as f:
        f.write(table(groups))
    keys = list(groups)
    chart_total(groups, keys)
    chart_wait(groups, keys)
    chart_eating(groups, keys)
    print("versões:", ", ".join(keys))
    print("-> results/tabela.md, tempo_total.png, espera.png, max_comendo.png")


if __name__ == "__main__":
    main()
