#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// examples/c/annotation/c-01.c
//
// Ouvrage Calque - Advanced C Systems Configuration Template (Annotation Version)
//
// This file is a fully compilable and valid standard C program in local dev,
// which compiles AOT to produce a static production binary configuration.

// Mock definitions for local development compilation
#define INIT_HW_PORT(addr, baud) printf("Mock hardware setup at %s (Baud: %d)\n", addr, baud)
#define ENABLE_INTERRUPTS(addr) printf("Mock interrupts enabled for %s\n", addr)

char* DB_URL = "sqlite://dev.db";

// ==========================================
// 1. MACRO DEFINITION (DECLARATIVE WRAPPER)
// ==========================================
// @ocq:Macro(
//   name="hw_port_setup",
//   use="c_function",
//   args=["gateway_addr", "baud", "alias", "trailing"]
// )
void _define_hw_port_setup() {
    // @ocq:Replace(with='printf("Initializing corpA secure port for %s...\n", ${alias});')
    printf("Initializing corpA secure port for %s...\n", "mock_alias");
    
    // @ocq:Replace(with='INIT_HW_PORT(${gateway_addr}, ${baud});')
    INIT_HW_PORT("0x3F8", 115200);
    
    // @ocq:Replace(with='${trailing};')
    ENABLE_INTERRUPTS("0x3F8");
    
    DB_URL = "postgresql://corpA_admin:secret@127.0.0.1:5432/cRSP_audit";
}
// @ocq:EndMacro(name="hw_port_setup")

int main() {
    printf("--- Starting System Initialization ---\n");

    // ==========================================
    // 2. HARDWARE SETUP (ACTIVE SHADOW MOCK)
    // ==========================================
    // @ocq:Call(macro="hw_port_setup", gateway_addr="0x3F8", baud=115200, alias="corpA_COM1", trailing='ENABLE_INTERRUPTS("0x3F8")')
    printf("Local dev: fallback to simulated serial console.\n");

    printf("Active Database Endpoint: %s\n", DB_URL);

    // ==========================================
    // 3. REGISTERS INITIALIZATION (AOT LOOP GENERATION)
    // ==========================================
    // @ocq:Loop(in=env.REGISTERS, var="reg", if=(env.NAME == "production"), replace='printf("Registering hardware node: ${reg.name} (Addr: ${reg.addr})\\n");')
    printf("Local dev: mock register initialization (Index: virtual)\n");

    printf("--- Initialization Complete ---\n");
    return 0;
}
