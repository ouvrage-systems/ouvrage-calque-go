#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// examples/c/stack/c-01.c
//
// Ouvrage Calque - Advanced C Systems Configuration Template (Stack Version)
//
// This file is a fully compilable and valid standard C program in local dev,
// which compiles AOT to produce a static production binary configuration.

// Mock definitions for local development compilation
#define INIT_HW_PORT(addr, baud) printf("Mock hardware setup at %s (Baud: %d)\n", addr, baud)
#define ENABLE_INTERRUPTS(addr) printf("Mock interrupts enabled for %s\n", addr)

char* DB_URL = "sqlite://dev.db";

// ==========================================
// 1. MACRO DEFINITION (STDLIB LOGIC & GEOMETRY WRAPPING)
// ==========================================
// @ocalque:stdlib:macro:define<hw_port_setup>
// @ocalque:stdlib:macro:arg<hw_port_setup> name="gateway_addr" type="string" required="true"
// @ocalque:stdlib:macro:arg<hw_port_setup> name="baud" type="int" required="true"
// @ocalque:stdlib:macro:arg<hw_port_setup> name="alias" type="string" required="true"
// @ocalque:stdlib:macro:arg<hw_port_setup> name="trailing" type="string" required="true"
// @ocalque:geometry:strip direction="next"
void _define_hw_port_setup() {
    // @ocalque:geometry:indent:pushd value="{{ macro.self.ref.indent - self.indent }}"
    
    // In local dev, we write valid C code. AOT replaces it with template variables.
    // @ocalque:replace_line --- printf("Initializing corpA secure port for %s...\n", {{ args.alias }});
    printf("Initializing corpA secure port for %s...\n", "mock_alias");
    
    // @ocalque:replace_line --- INIT_HW_PORT({{ args.gateway_addr }}, {{ args.baud }});
    INIT_HW_PORT("0x3F8", 115200);
    
    // @ocalque:replace_line --- {{ args.trailing }};
    ENABLE_INTERRUPTS("0x3F8");
    
    DB_URL = "postgresql://corpA_admin:secret@127.0.0.1:5432/cRSP_audit";
    // @ocalque:geometry:indent:popd
    // @ocalque:geometry:strip direction="next"
}
// @ocalque:stdlib:macro:end<hw_port_setup>

int main() {
    printf("--- Starting System Initialization ---\n");

    // ==========================================
    // 2. HARDWARE SETUP (ACTIVE SHADOW MOCK)
    // ==========================================
    // @ocalque:stdlib:if<hw_setup> --- "{{ env.NAME }}" == "production"
        // @ocalque:stdlib:macro:call<hw_port_setup> gateway_addr="0x3F8" baud=115200 alias="corpA_COM1" --- ENABLE_INTERRUPTS("0x3F8")
    // @ocalque:stdlib:else<hw_setup>
        printf("Local dev: fallback to simulated serial console.\n");
    // @ocalque:stdlib:fi<hw_setup>

    printf("Active Database Endpoint: %s\n", DB_URL);

    // ==========================================
    // 3. REGISTERS INITIALIZATION (AOT LOOP GENERATION)
    // ==========================================
    // @ocalque:stdlib:for<reg_loop> item="reg" in="env.REGISTERS"
    //   @ocalque:stdlib:if<prod_reg> --- "{{ env.NAME }}" == "production"
    //     @ocalque:geometry:strip direction="next"
    if (0) {
        // @ocalque:replace_line --- printf("Registering hardware node: {{ args.reg.name }} (Addr: {{ args.reg.addr }})\n");
        printf("Registering hardware node: %s (Addr: %s)\n", "mock_name", "0x00");
        // @ocalque:geometry:strip direction="next"
    }
    //   @ocalque:stdlib:else<prod_reg>
        printf("Local dev: mock register initialization (Index: virtual)\n");
    //   @ocalque:stdlib:fi<prod_reg>
    // @ocalque:stdlib:end<reg_loop>

    printf("--- Initialization Complete ---\n");
    return 0;
}
