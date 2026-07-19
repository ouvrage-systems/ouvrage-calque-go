; examples/asm/setup.asm
;
; Ouvrage Calque - x86_64 Assembly Template (NASM Syntax)
;
; This file is a fully valid and compilable Assembly script in local dev,
; which bypasses production instructions using a local GOTO/JMP instruction.

global _start

section .text
_start:
    ; ==========================================
    ; 1. SYSTEM INITIALIZATION
    ; ==========================================
    mov al, 0x01        ; Load control byte (e.g. initialization code)

    ; ==========================================
    ; 2. HARDWARE REGISTER SETUP (ACTIVE SHADOW MOCK)
    ; ==========================================
    ; @ocalque:stdlib:if<hw_setup> --- "{{ env.NAME }}" == "production"
        ; @ocalque:geometry:strip direction="next"
        jmp .local_dev_mock
        ; @ocalque:geometry:strip direction="next"
        
        ; In local dev, we write a valid register address. AOT replaces it with template variable.
        ; @ocalque:replace_line --- mov dx, {{ env.HW_ADDR }}
        mov dx, 0x3F8   ; Production device register port
        
        out dx, al      ; Write control byte to production port
    .local_dev_mock:
        ; @ocalque:stdlib:else<hw_setup>
        ; Local dev mock output: simulator write
        mov dx, 0x00    ; Simulated stdout / null register
        out dx, al
        ; @ocalque:stdlib:fi<hw_setup>

    ; ==========================================
    ; 3. EXIT SYSTEM CALL
    ; ==========================================
    mov rax, 60         ; sys_exit system call number
    mov rdi, 0          ; exit code 0
    syscall
