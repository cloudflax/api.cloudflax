#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

# Run tests with coverage and capture output
# We use -cover to get coverage percentages
OUTPUT=$(go test -v -cover ./... 2>&1)
EXIT_CODE=$?

# Process output with AWK
echo "$OUTPUT" | awk -v green="$GREEN" -v red="$RED" -v yellow="$YELLOW" -v blue="$BLUE" -v reset="$RESET" '
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
    passed_count = 0;
    failed_count = 0;
}

# Capture individual test result
/^--- (PASS|FAIL):/ {
    status = $2; sub(/:/, "", status);
    name = $3;
    time_val = 0;
    time_str = "0.000s";
    
    if (match($0, /\([0-9.]+s\)/)) {
        t_str = substr($0, RSTART+1, RLENGTH-2);
        sub(/s$/, "", t_str);
        time_val = t_str + 0;
        time_str = sprintf("%.3fs", time_val);
    }
    
    # Highlight slow tests (> 0.5s)
    t_color = (time_val > 0.5) ? yellow : "";
    s_color = (status == "PASS" ? green : red);
    
    if (status == "PASS") passed_count++; else failed_count++;
    
    test_count++;
    test_buffer[test_count] = sprintf("  %s%s %-56s [%s%s%s]%s", s_color, status, format_name(name), t_color, time_str, s_color, reset);
    next;
}

# Capture error details
/^[[:space:]]+.*_test\.go:[0-9]+:/ || /^[[:space:]]+Error Trace:/ || /^[[:space:]]+Error:/ || /^[[:space:]]+Messages:/ || /^[[:space:]]+expected:/ || /^[[:space:]]+actual  :/ {
    if (test_count > 0) {
        line = $0; sub(/^[[:space:]]+/, "", line);
        test_buffer[test_count] = test_buffer[test_count] "\n" yellow "    " line reset;
    }
    next;
}

# Capture coverage
/^coverage: [0-9.]+% of statements/ {
    match($0, /[0-9.]+/);
    pkg_coverage = substr($0, RSTART, RLENGTH) "%";
    next;
}

# When a package finishes
/^(ok|FAIL)[[:space:]]+github.com/ {
    pkg_status = ($1 == "ok" ? "PASS" : "FAIL");
    pkg_color = ($1 == "ok" ? "" : red);
    pkg_full_name = $2;
    
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
    
    # Print package header with coverage
    cov_str = (pkg_coverage != "") ? " (Cov: " pkg_coverage ")" : "";
    printf "%s%s %-45s %12s [%s]%s\n", pkg_color, pkg_status, pkg_name, blue cov_str reset pkg_color, pkg_time, reset;
    
    for (i = 1; i <= test_count; i++) {
        print test_buffer[i];
    }
    
    test_count = 0;
    pkg_coverage = "";
    delete test_buffer;
    next;
}

END {
    printf "TOTAL_TIME:%.3fs|PASSED:%d|FAILED:%d\n", total_time, passed_count, failed_count;
}

{ next; }
' > .test_output

# Extract stats and clean output
STATS=$(grep "TOTAL_TIME:" .test_output)
TOTAL_TIME=$(echo $STATS | cut -d'|' -f1 | cut -d: -f2)
PASSED=$(echo $STATS | cut -d'|' -f2 | cut -d: -f2)
FAILED=$(echo $STATS | cut -d'|' -f3 | cut -d: -f2)
TOTAL=$((PASSED + FAILED))

sed -i '/TOTAL_TIME:/d' .test_output
cat .test_output
rm .test_output

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    printf "${GREEN}✓ All tests passed [${TOTAL_TIME}]${RESET}\n"
    printf "${BLUE}  Stats: ${GREEN}${PASSED} passed${BLUE}, ${TOTAL} total${RESET}\n"
else
    printf "${RED}✗ Some tests failed [${TOTAL_TIME}]${RESET}\n"
    printf "${BLUE}  Stats: ${GREEN}${PASSED} passed${BLUE}, ${RED}${FAILED} failed${BLUE}, ${TOTAL} total${RESET}\n"
fi

exit $EXIT_CODE
