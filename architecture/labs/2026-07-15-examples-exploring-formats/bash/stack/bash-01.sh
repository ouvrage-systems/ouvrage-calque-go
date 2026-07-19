#!/bin/bash
# examples/bash/bash-01.sh
# This example demonstrates embedded macro definition, schema validation, direct calling,
# and compile-time block unwrapping using 'strip_next' and deferred 'strip_last' rules.

# ==========================================
# 1. EMBEDDED MACRO DEFINITION
# ==========================================
# @ocalque:macro:define<corpA_get>
# @ocalque:macro:arg<corpA_get> name="key" type="int" required="true" --- The key to retrieve from the corpA FDGate.
# @ocalque:macro:arg<corpA_get> name="trailing" type="string" required="true" --- The trailing command to execute.
# @ocalque:strip_next
_ocalque_macro_corpA_get() {
    # The function wrapper keeps the definition dormant and valid in local dev.
    echo "# FDGATE HTTP_GET {{ args.key }}" >&3
    read -r "{{ args.trailing }}" <&4
    # @ocalque:strip_next
}
# @ocalque:macro:end<corpA_get>

# @ocalque:macro:define<file_copyright>
# @ocalque:macro:arg name=filename type=string required=true --- The name of the file for the copyright header.
# @ocalque:indent:capture ref=_
# @ocalque:strip_next
_header_copyright() {
    # @ocalque:indent:apply relative_to=ref
    # Copyright {{ runtime.date }} - {{ env.department }}
    # For file: {{ args.filename }}
    # @ocalque:strip_next
}
# @ocalque:macro:end<file_copyright>

# ==========================================
# 2. DIRECT MACRO CALL
# ==========================================
# @ocalque:macro:call<file_copyright> filename=bash-01.sh
# @ocalque:macro:call<corpA_get> key=42 --- STATUS_ALL

# ==========================================
# LOCAL DEVELOPMENT FALLBACK
# ==========================================
# @ocalque:ref name="local_mock_start"
# @ocalque:strip_following lines=2
STATUS_ALL='{"status": "ACTIVE_MOCK"}' # Local dev mock response
echo "Copyright 2026 - local-dev" # Local dev mock response for copyright
# @ocalque:ref name="local_mock_end"

# Processing results
STATUS=$(echo $STATUS_ALL | jq -r '.status')
echo "Target Endpoint Status: $STATUS"
