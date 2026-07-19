#!/bin/bash
# examples/bash/bash-01.sh (Annotation Syntax)
# This example demonstrates embedded macro definition, schema validation, direct calling,
# and compile-time block unwrapping using 'Strip' and automated indentation calibration.

# ==========================================
# 1. EMBEDDED MACRO DEFINITIONS
# ==========================================
# @ocq:Macro(
#   name="corpA_get",
#   args=["key", "trailing"]
# )
# @ocq:Strip(lines=1)
_ocalque_macro_corpA_get() {
    # The function wrapper keeps the definition dormant and valid in local dev.
    # @ocq:Replace(with="    echo '# FDGATE HTTP_GET ${key}' >&3")
    echo "# FDGATE HTTP_GET 42" >&3
    # @ocq:Replace(with="    read -r '${trailing}' <&4")
    read -r "MOCK_TRAILING" <&4
# @ocq:EndMacro

# @ocq:Macro(
#   name="file_copyright",
#   args=["filename"]
# )
# @ocq:Strip(lines=1)
_header_copyright() {
    # @ocq:Replace(with="    # Copyright ${runtime.date} - ${env.department}")
    # Copyright 2026 - SRE-Department
    # @ocq:Replace(with="    # For file: ${filename}")
    # For file: mock_filename
# @ocq:EndMacro

# ==========================================
# 2. DIRECT MACRO CALLS
# ==========================================
# @ocq:Call(macro="file_copyright", filename="bash-01.sh")
# @ocq:Call(macro="corpA_get", key=42, trailing="STATUS_ALL")

# ==========================================
# LOCAL DEVELOPMENT FALLBACK
# ==========================================
# @ocq:Ref(name="local_mock_start")
# @ocq:Strip(lines=2)
STATUS_ALL='{"status": "ACTIVE_MOCK"}' # Local dev mock response
echo "Copyright 2026 - local-dev" # Local dev mock response for copyright
# @ocq:Ref(name="local_mock_end")

# Processing results
STATUS=$(echo $STATUS_ALL | jq -r '.status')
echo "Target Endpoint Status: $STATUS"
