; examples/asm/setup.asm (Annotation Syntax)
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
    ; @ocq:If(cond=(env.NAME == "production"))
        ; @ocq:Strip(lines=1)
        jmp .local_dev_mock
        
        ; In local dev, we write a valid register address. AOT replaces it with template variable.
        ; @ocq:Replace(with="        mov dx, ${env.HW_ADDR}")
        mov dx, 0x3F8   ; Production device register port
        
        out dx, al      ; Write control byte to production port
    .local_dev_mock:
    ; @ocq:Else
        ; Local dev mock output: simulator write
        mov dx, 0x00    ; Simulated stdout / null register
        out dx, al
    ; @ocq:EndIf

    ; ==========================================
    ; 3. EXIT SYSTEM CALL
    ; ==========================================
    mov rax, 60         ; sys_exit system call number
    mov rdi, 0          ; exit code 0
    syscall
