package examples.java;

// examples/java/JavaConfig.java (Annotation Syntax)
//
// Ouvrage Calque - Advanced Java Systems Configuration Template
//
// This file is a fully compilable and valid Java class in local dev,
// which compiles AOT to produce a static production configuration module.

public class JavaConfig {

    private static String DB_URL = "jdbc:h2:mem:devdb";
    private static String DB_USER = "sa";

    // Mock register host function for type checking
    private static void registerHost(String name, String ip) {
        System.out.println("Mock register: " + name + " -> " + ip);
    }

    // ==========================================
    // 1. MACRO DEFINITION (STDLIB LOGIC & GEOMETRY WRAPPING)
    // ==========================================
    // @ocq:Macro(name="db_setup", args=["url", "user"])
    // @ocq:Strip(lines=1)
    private static void _define_db_setup() {
        // In local dev, we write valid Java. AOT replaces it with template variables.
        // @ocq:Replace(with="        DB_URL = ${url};")
        DB_URL = "jdbc:postgresql://prod-db:5432/siemens";
        
        // @ocq:Replace(with="        DB_USER = ${user};")
        DB_USER = "siemens_prod";
        
        System.out.println("Production DB active: " + DB_URL);
        // @ocq:Strip(lines=1)
    }
    // @ocq:EndMacro

    public static void main(String[] args) {
        System.out.println("--- Starting Java Configuration Setup ---");

        // ==========================================
        // 2. ACTIVE SHADOW MOCK FOR DATABASE SETUP
        // ==========================================
        // @ocq:If(cond=(env.NAME == "production"))
            // @ocq:Call(macro="db_setup", url="jdbc:postgresql://prod-db:5432/siemens", user="siemens_prod")
        // @ocq:Else
            System.out.println("Local dev: active database mock mode.");
        // @ocq:EndIf

        System.out.println("Database URL: " + DB_URL + " (User: " + DB_USER + ")");

        // ==========================================
        // 3. HOSTS REGISTRATION (AOT LOOP GENERATION)
        // ==========================================
        // @ocq:Loop(in=env.HOSTS, var="host")
        //   @ocq:If(cond=(env.NAME == "production"))
        //     @ocq:Strip(lines=1)
        if (false) {
            // @ocq:Replace(with="            registerHost(${host.name}, ${host.ip});")
            registerHost("mock_host", "127.0.0.1");
            // @ocq:Strip(lines=1)
        }
        //   @ocq:Else
            System.out.println("Local dev: mock host registration (virtual dev network)");
        //   @ocq:EndIf
        // @ocq:EndLoop

        System.out.println("--- Java Setup Complete ---");
    }
}
