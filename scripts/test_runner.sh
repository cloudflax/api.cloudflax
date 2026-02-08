#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
RESET='\033[0m'

# Run tests and capture output
OUTPUT=$(go test -v ./... 2>&1)
EXIT_CODE=$?

# Process output with AWK
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
    total_time = 0;
}

# Capture individual test result and save to buffer
/^--- (PASS|FAIL):/ {
    status = $2; sub(/:/, "", status);
    color = (status == "PASS" ? green : red);
    name = $3;
    time = "0.000s";
    if (match($0, /\([0-9.]+s\)/)) {
        t_str = substr($0, RSTART+1, RLENGTH-2);
        sub(/s$/, "", t_str);
        time = sprintf("%.3fs", t_val = t_str + 0);
    }
    
    test_count++;
    test_buffer[test_count] = sprintf("  %s%s %-56s [%s]%s", color, status, format_name(name), time, reset);
    next;
}

# Capture error details and save to current test buffer
/^[[:space:]]+.*_test\.go:[0-9]+:/ || /^[[:space:]]+Error Trace:/ || /^[[:space:]]+Error:/ || /^[[:space:]]+Messages:/ || /^[[:space:]]+expected:/ || /^[[:space:]]+actual  :/ {
    if (test_count > 0) {
        line = $0; sub(/^[[:space:]]+/, "", line);
        test_buffer[test_count] = test_buffer[test_count] "\n" yellow "    " line reset;
    }
    next;
}

# When a package finishes (ok or FAIL)
/^(ok|FAIL)[[:space:]]+github.com/ {
    pkg_status = ($1 == "ok" ? "PASS" : "FAIL");
    pkg_color = ($1 == "ok" ? "" : red);
    pkg_full_name = $2;
    
    # Shorten package name: github.com/cloudflax/api.cloudflax/internal/user -> api.cloudflax/internal/user
    pkg_name = pkg_full_name;
    sub(/^github\.com\/cloudflax\//, "", pkg_name);
    
    if ($0 ~ /cached/) {
        t_val = 0;
        pkg_time = "0.000s";
    } else {
        t_val = $3; sub(/s$/, "", t_val);
        t_val = t_val + 0;
        pkg_time = sprintf("%.3fs", t_val);
    }
    
    total_time += t_val;
    
    # Print package header with shortened name and consistent padding
    printf "%s%s %-58s [%s]%s\n", pkg_color, pkg_status, pkg_name, pkg_time, reset;
    
    for (i = 1; i <= test_count; i++) {
        print test_buffer[i];
    }
    
    test_count = 0;
    delete test_buffer;
    next;
}

# End block to pass total time back to shell
END {
    printf "TOTAL_TIME:%.3fs\n", total_time;
}

# Ignore other lines
{ next; }
' > .test_output

# Extract total time and clean output
TOTAL_TIME=$(grep "TOTAL_TIME:" .test_output | cut -d: -f2)
sed -i '/TOTAL_TIME:/d' .test_output
cat .test_output
rm .test_output

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    printf "${GREEN}✓ All tests passed [${TOTAL_TIME}]${RESET}\n"
else
    printf "${RED}✗ Some tests failed [${TOTAL_TIME}]${RESET}\n"
fi

exit $EXIT_CODE
