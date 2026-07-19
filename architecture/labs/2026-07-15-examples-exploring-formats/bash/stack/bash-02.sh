#!/bin/bash
# examples/bash/bash-02.sh
# This example demonstrates the AOT loop ('for') directive combined with block unwrapping.

# ==========================================
# PRODUCTION CONFIGURATION (LOOP COMPILATION)
# ==========================================
# @ocalque:for<node_loop> item="node" in="cluster.nodes" strip_start=1 strip_end=1
_ouvrage_loop() {
    # The function wrapper keeps the loop body dormant and valid in local dev.
    echo "Registering node: {{ args.node.name }} at {{ args.node.ip }}"
}
# @ocalque:block:end name="node_loop"

# ==========================================
# LOCAL DEVELOPMENT FALLBACK (MOCK)
# ==========================================
# @ocalque:strip_following lines=1
echo "Registering mock node: local-dev at 127.0.0.1" # Local fallback

echo "Initialization complete."
