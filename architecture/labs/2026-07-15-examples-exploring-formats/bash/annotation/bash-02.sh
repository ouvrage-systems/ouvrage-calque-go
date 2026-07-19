#!/bin/bash
# examples/bash/bash-02.sh (Annotation Syntax)
# This example demonstrates the AOT loop ('Loop') directive combined with block unwrapping.

# ==========================================
# PRODUCTION CONFIGURATION (LOOP COMPILATION)
# ==========================================
# @ocq:Loop(in=cluster.nodes, var="node")
# @ocq:Strip(lines=1)
_ouvrage_loop() {
    # The function wrapper keeps the loop body dormant and valid in local dev.
    # @ocq:Replace(with="    echo 'Registering node: ${node.name} at ${node.ip}'")
    echo "Registering node: mock-node at 127.0.0.1"
    # @ocq:Strip(lines=1)
}
# @ocq:EndLoop

# ==========================================
# LOCAL DEVELOPMENT FALLBACK (MOCK)
# ==========================================
# @ocq:Strip(lines=1)
echo "Registering mock node: local-dev at 127.0.0.1" # Local fallback

echo "Initialization complete."
