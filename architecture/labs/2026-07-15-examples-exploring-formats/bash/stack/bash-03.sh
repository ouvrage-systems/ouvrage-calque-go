#!/bin/bash
# examples/bash/bash-03.sh
#
# Ouvrage Calque - Advanced SRE Configuration Compilation Template
#
# This file is an active, executable Bash script in local dev that compiles
# AOT to a production-ready deployment script for corpA secure networks.

# ==========================================
# GLOBAL AST METADATA (For Calque Graph Extraction)
# ==========================================
# @ocalque:meta owner="Ouvrage-SRE-Core" security-level="restricted" compliance="corpA-cRSP-v4"

# @ocalque:geometry:strip lines=3
# ==========================================
# 1. SECURE GATEWAY MACRO DEFINITION
# ==========================================
# @ocalque:macro:define<secure_gateway_setup>
# @ocalque:macro:arg<secure_gateway_setup> name="gateway_host" type="string" required="true"
# @ocalque:macro:arg<secure_gateway_setup> name="port" type="int" required="true"
# @ocalque:macro:arg<secure_gateway_setup> name="alias" type="string" required="true"
# @ocalque:macro:arg<secure_gateway_setup> name="trailing" type="string" required="true" --- The raw SSH tunnel command
# @ocalque:strip_next
_define_secure_gateway_setup() {
    # @ocalque:indent:pushd value="{{ macro.self.ref.indent - self.indent }}"
    echo "Initializing secure network bridge for {{ args.alias }}..."
    
    # Establish tunnel in background (using implicit trailing parameter)
    {{ args.trailing }} &
    TUNNEL_PID=$!
    sleep 1 # Wait for tunnel to bind
    
    DB_URL="postgresql://{{ env.DB_USER }}:{{ env.DB_PASS }}@127.0.0.1:{{ args.port }}/{{ args.alias }}"
    # @ocalque:indent:popd
    # @ocalque:strip_next
}
# @ocalque:macro:end<secure_gateway_setup>

# ==========================================
# 2. APPLICATION CONFIGURATION (ACTIVE SHADOW MOCK PATTERN)
# ==========================================

# ocalque evaluates env.NAME and compiles the appropriate branch AOT.
# @ocalque:if<db_setup> --- "{{ env.NAME }}" == "production"

    # In production, call the secure gateway macro
    # @ocalque:macro:call<secure_gateway_setup> gateway_host="gw-crsp.corpA.net" port=5432 alias="cRSP_audit" --- ssh -N -L 5432:db-prod.corpA.internal:5432 gateway.corpA.com

# @ocalque:else<db_setup>

    # In local development, fall back to a harmless SQLite mock
    echo "Initializing local SQLite database for development..."
    touch dev_database.db
    DB_URL="sqlite://dev_database.db"

# @ocalque:fi<db_setup>

echo "Active Database Endpoint: $DB_URL"

# ==========================================
# 3. REPLICA MONITORING CONFIGURATION (AOT LOOP GENERATION)
# ==========================================

# Iterate AOT over DB replicas database configuration
# @ocalque:for<replica_loop> item="rep" in="env.DB_REPLICAS"
#   @ocalque:meta monitoring-type="ping" timeout-ms="500"
#   @ocalque:if<prod_monitoring> --- "{{ env.NAME }}" == "production"
#     @ocalque:strip_next
_prod_monitoring_wrapper() {
    echo "Spawning production ping daemon for: {{ args.rep.name }} ({{ args.rep.ip }})"
    # @ocalque:strip_next
}
#   @ocalque:else<prod_monitoring>
    echo "Local mock: skipping ping for dev replica (127.0.0.1)"
#   @ocalque:fi<prod_monitoring>
# @ocalque:end<replica_loop>

# ==========================================
# 4. PROGRESSIVE SYSTEM BACKUP (DYNAMIC GEOMETRY INDENTATION & VARIABLES)
# ==========================================

# Define a starting port variable for backups (explicit type 'int')
# @ocalque:stdlib:var:define name="backup_port" type="int" value="9000"

# Generates clean indented log hierarchy for backups
# @ocalque:stdlib:count<indent_loop> item="i" range=3
#   @ocalque:geometry:indent:pushd value="{{ count.indent_loop.i }}"
echo "Creating backup checkpoint level {{ count.indent_loop.i }} on port {{ var.backup_port }}..."
#   @ocalque:stdlib:var:set name="backup_port" value="{{ var.backup_port + 1 }}"
#   @ocalque:geometry:indent:popd
# @ocalque:stdlib:end<indent_loop>

# ==========================================
# 5. MODULAR TEMPLATE IMPORT (GEOMETRIC SLICING)
# ==========================================

# First, evaluate the external template to load its references into the 'lib' namespace
# @ocalque:import source="examples/bash/bash-01.sh" mode="eval" as="lib"

# Next, insert the exact slice between 'local_mock_start' and 'local_mock_end' from lib
# @ocalque:import source="examples/bash/bash-01.sh" mode="insert" from="{{ lib.ref.local_mock_start.line_nb }}" to="{{ lib.ref.local_mock_end.line_nb }}"

echo "Deployment script execution complete."
