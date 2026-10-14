#!/usr/bin/env bash
# Gera relatorio.pdf a partir de relatorio.md
# Uso: bash relatorio/gerar_pdf.sh
set -e
cd "$(dirname "$0")"

pandoc relatorio.md \
  -o relatorio.pdf \
  --pdf-engine=xelatex \
  --template=template.tex \
  --highlight-style=monochrome

echo "-> relatorio/relatorio.pdf"