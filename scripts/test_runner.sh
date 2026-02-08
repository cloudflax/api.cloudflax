#!/bin/bash

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
RESET='\033[0m'

# Ejecutar tests y capturar salida
OUTPUT=$(go test -v ./... 2>&1)
EXIT_CODE=$?

echo "$OUTPUT" | awk -v green="$GREEN" -v red="$RED" -v yellow="$YELLOW" -v reset="$RESET" '
BEGIN { count = 0; pkg_sum = 0; }

# 1. Identificar el inicio de un test individual
/^[[:space:]]*=== RUN/ {
    name = $3;
    if (!(name in test_seen)) {
        count++;
        names[count] = name;
        test_seen[name] = count;
        statuses[count] = "PASS";
        colors[count] = green;
    }
    next;
}

# 2. Si detectamos un fallo explícito en un test
/^--- FAIL:/ {
    name = $3;
    idx = test_seen[name];
    statuses[idx] = "FAIL";
    colors[idx] = red;
    
    raw_time = $4; gsub(/[()s]/, "", raw_time);
    times[idx] = sprintf("%.2fs", raw_time);
    next;
}

# 3. Si detectamos un éxito explícito
/^--- PASS:/ {
    name = $3;
    idx = test_seen[name];
    statuses[idx] = "PASS";
    colors[idx] = green;
    
    raw_time = $4; gsub(/[()s]/, "", raw_time);
    times[idx] = sprintf("%.2fs", raw_time);
    next;
}

# 4. Capturar detalles de error
/^[[:space:]]+.*_test\.go:[0-9]+:/ || /^[[:space:]]+Error Trace:/ || /^[[:space:]]+Error:/ || /^[[:space:]]+Messages:/ || /^[[:space:]]+expected:/ || /^[[:space:]]+actual  :/ {
    sub(/^[[:space:]]+/, "");
    # AJUSTE: Ahora usamos 4 espacios para que el detalle se alinee con el inicio de FAIL/PASS
    details[count] = details[count] yellow "    " $0 reset "\n";
    next;
}

# 5. Resumen del paquete
/^ok[[:space:]]+/ || /^FAIL[[:space:]]+github.com/ {
    pkg = $2;
    is_cached = ($0 ~ /cached/);
    pkg_status = ($1 == "ok" ? "PASS" : "FAIL");
    
    if (!is_cached) {
        raw_pkg = $3; gsub(/s/, "", raw_pkg);
        pkg_time = sprintf("%.2fs", raw_pkg);
    } else {
        pkg_time = "0.00s";
    }

    # Cabecera de paquete (Sin sangría)
    printf "%s %-58s [%s]\n", pkg_status, pkg, pkg_time;
    
    suffix = is_cached ? "(cached)" : "";

    for (i = 1; i <= count; i++) {
        t = (times[i] == "" ? "0.00s" : times[i]);
        # Los tests hijos tienen 2 espacios de sangría inicial
        printf "  %s%s %-56s [%s] %s%s\n", colors[i], statuses[i], names[i], t, suffix, reset;
        if (details[i] != "") printf "%s", details[i];
    }

    delete names; delete times; delete statuses; delete colors; delete details; delete test_seen;
    count = 0;
    next;
}
'

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    printf "${GREEN}✓ All tests passed${RESET}\n"
else
    printf "${RED}✗ Some tests failed${RESET}\n"
fi

exit $EXIT_CODE