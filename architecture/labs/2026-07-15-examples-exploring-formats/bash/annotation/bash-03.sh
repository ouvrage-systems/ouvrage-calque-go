#!/bin/bash
# examples/bash/bash-03.sh (Annotation Syntax)
#
# Ouvrage Calque - Advanced SRE Configuration Compilation Template
#
# This file is an active, executable Bash script in local dev that compiles
# AOT to a production-ready deployment script for corpA secure networks.

# ==========================================
# GLOBAL AST METADATA (For Calque Graph Extraction)
# ==========================================
# @ocq:Meta(owner="Ouvrage-SRE-Core", security_level="restricted", compliance="corpA-cRSP-v4")

# ==========================================
# 1. SECURE GATEWAY MACRO DEFINITION
# ==========================================
# @ocq:Macro(name="secure_gateway_setup", args=["gateway_host", "port", "alias", "trailing"])
# @ocq:Strip(lines=1)
_define_secure_gateway_setup() {
    # @ocq:Replace(with="    echo 'Initializing secure network bridge for ${alias}...'")
    echo "Initializing secure network bridge for mock_alias..."
    
    # Establish tunnel in background
    # @ocq:Replace(with="    ${trailing} &")
    ssh -N -L 5432:db-prod.corpA.internal:5432 gateway.corpA.com &
    TUNNEL_PID=$!
    sleep 1 # Wait for tunnel to bind
    
    # @ocq:Replace(with="    DB_URL='postgresql://${env.DB_USER}:${env.DB_PASS}@127.0.0.1:${port}/${alias}'")
    DB_URL="sqlite://dev_database.db"
    # @ocq:Strip(lines=1)
}
# @ocq:EndMacro

# ==========================================
# 2. APPLICATION CONFIGURATION (ACTIVE SHADOW MOCK PATTERN)
# ==========================================
# @ocq:If(cond=(env.NAME == "production"))
    # In production, call the secure gateway macro
    # @ocq:Call(macro="secure_gateway_setup", gateway_host="gw-crsp.corpA.net", port=5432, alias="cRSP_audit", trailing="ssh -N -L 5432:db-prod.corpA.internal:5432 gateway.corpA.com")
# @ocq:Else
    # In local development, fall back to a harmless SQLite mock
    echo "Initializing local SQLite database for development..."
    touch dev_database.db
    DB_URL="sqlite://dev_database.db"
# @ocq:EndIf

echo "Active Database Endpoint: $DB_URL"

# ==========================================
# 3. REPLICA MONITORING CONFIGURATION (AOT LOOP GENERATION)
# ==========================================
# @ocq:Loop(in=env.DB_REPLICAS, var="rep")
#   @ocq:If(cond=(env.NAME == "production"))
#     @ocq:Strip(lines=1)
_prod_monitoring_wrapper() {
    # @ocq:Replace(with="    echo 'Spawning production ping daemon for: ${rep.name} (${rep.ip})'")
    echo "Spawning production ping daemon for: mock_replica"
#     @ocq:Strip(lines=1)
}
#   @ocq:Else
    echo "Local mock: skipping ping for dev replica (127.0.0.1)"
#   @ocq:EndIf
# @ocq:EndLoop

# ==========================================
# 4. PROGRESSIVE SYSTEM BACKUP (DYNAMIC GEOMETRY INDENTATION & VARIABLES)
# ==========================================
# @ocq:Var(name="backup_port", type="int", value=9000)
# @ocq:Loop(in=[0, 1, 2], var="i")
#   @ocq:Indent(delta=i)
#   @ocq:Replace(with="echo 'Creating backup checkpoint level ${i} on port ${backup_port}...'")
echo "Creating backup checkpoint level 0 on port 9000..."
#   @ocq:Var(name="backup_port", value=(backup_port + 1))
# @ocq:EndLoop

# ==========================================
# 5. MODULAR TEMPLATE IMPORT (GEOMETRIC SLICING)
# ==========================================
# First, evaluate the external template to load its references into the 'lib' namespace
# @ocq:Import(source="examples/bash/bash-01.sh", mode="eval", as="lib")

# Next, insert the exact slice between 'local_mock_start' and 'local_mock_end' from lib
# @ocq:Import(source="examples/bash/bash-01.sh", mode="insert", from=lib.ref.local_mock_start.line_number, to=lib.ref.local_mock_end.line_number)

echo "Deployment script execution complete."
