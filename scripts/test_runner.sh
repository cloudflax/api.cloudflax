#!/bin/bash

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
RESET='\033[0m'

# Ejecutar tests y capturar salida
OUTPUT=$(go test -v ./... 2>&1)
EXIT_CODE=$?

# Procesar la salida con AWK
echo "$OUTPUT" | awk -v green="$GREEN" -v red="$RED" -v yellow="$YELLOW" -v reset="$RESET" '
function format_name(n) {
    sub(/^Test/, "", n);
    gsub(/_/, " ", n);
    res = "";
    len = length(n);
    for (i = 1; i <= len; i++) {
        c = substr(n, i, 1);
        if (i > 1 && c ~ /[A-Z]/ && substr(n, i-1, 1) ~ /[a-z0-9]/) {
            res = res " ";
        }
        res = res c;
    }
    res = tolower(res);
    gsub(/[ ]+/, " ", res);
    sub(/^[ ]+/, "", res);
    return toupper(substr(res, 1, 1)) substr(res, 2);
}

BEGIN { 
    test_count = 0; 
}

# Capturar resultado de test individual y guardarlo en un buffer
/^--- (PASS|FAIL):/ {
    status = $2; sub(/:/, "", status);
    color = (status == "PASS" ? green : red);
    name = $3;
    time = "0.000s";
    if (match($0, /\(([0-9.]+)s\)/)) {
        # Extraer el valor numérico y formatear a 3 decimales
        t_str = substr($0, RSTART+1, RLENGTH-2);
        sub(/s$/, "", t_str);
        time = sprintf("%.3fs", t_val = t_str + 0);
    }
    
    test_count++;
    test_buffer[test_count] = sprintf("  %s%s %-56s [%s]%s", color, status, format_name(name), time, reset);
    next;
}

# Capturar errores y guardarlos en el buffer del test actual
/^[[:space:]]+.*_test\.go:[0-9]+:/ || /^[[:space:]]+Error Trace:/ || /^[[:space:]]+Error:/ || /^[[:space:]]+Messages:/ || /^[[:space:]]+expected:/ || /^[[:space:]]+actual  :/ {
    if (test_count > 0) {
        line = $0; sub(/^[[:space:]]+/, "", line);
        test_buffer[test_count] = test_buffer[test_count] "\n" yellow "    " line reset;
    }
    next;
}

# Cuando termina un paquete (ok o FAIL)
/^(ok|FAIL)[[:space:]]+github.com/ {
    pkg_status = ($1 == "ok" ? "PASS" : "FAIL");
    pkg_color = ($1 == "ok" ? "" : red);
    pkg_name = $2;
    
    # Formatear tiempo del paquete a 3 decimales
    if ($0 ~ /cached/) {
        pkg_time = "0.000s";
    } else {
        t_val = $3; sub(/s$/, "", t_val);
        pkg_time = sprintf("%.3fs", t_val + 0);
    }
    
    # 1. Imprimir primero la línea del paquete
    printf "%s%s %-58s [%s]%s\n", pkg_color, pkg_status, pkg_name, pkg_time, reset;
    
    # 2. Imprimir todos los tests guardados para este paquete
    for (i = 1; i <= test_count; i++) {
        print test_buffer[i];
    }
    
    # 3. Limpiar el buffer para el siguiente paquete
    test_count = 0;
    delete test_buffer;
    next;
}

# Ignorar otras líneas
{ next; }
'

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    printf "${GREEN}✓ All tests passed${RESET}\n"
else
    printf "${RED}✗ Some tests failed${RESET}\n"
fi

exit $EXIT_CODE
