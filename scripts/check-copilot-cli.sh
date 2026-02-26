#!/usr/bin/env bash
set -euo pipefail

# Strictly match Copilot CLI agent processes only.
matches="$(ps -axo pid=,command= | awk '
  BEGIN { IGNORECASE = 1 }
  {
    cmd = $0

    # Exclude VS Code/TypeScript plugin processes.
    if (cmd ~ /tsserver/ || cmd ~ /copilot-typescript-server-plugin/ || cmd ~ /Code - Insiders Helper/ || cmd ~ /Code Helper/) {
      next
    }

    # Include only likely Copilot CLI agent entrypoints.
    if (cmd ~ /(^|[[:space:]])gh[[:space:]]+copilot([[:space:]]|$)/ ||
        cmd ~ /(^|[[:space:]\/])copilot-cli([[:space:]]|$)/ ||
        cmd ~ /(^|[[:space:]\/])copilot-agent([[:space:]]|$)/ ||
        cmd ~ /(^|[[:space:]\/])github-copilot-cli([[:space:]]|$)/) {
      print
    }
  }
' || true)"

if [[ -z "$matches" ]]; then
  echo "No running Copilot CLI agent processes found."
  exit 0
fi

echo "Running Copilot CLI agent processes:"
echo "$matches"
