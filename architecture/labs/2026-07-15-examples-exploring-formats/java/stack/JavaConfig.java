package examples.java;

// examples/java/JavaConfig.java
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
    // @ocalque:stdlib:macro:define<db_setup>
    // @ocalque:stdlib:macro:arg<db_setup> name="url" type="string" required="true"
    // @ocalque:stdlib:macro:arg<db_setup> name="user" type="string" required="true"
    // @ocalque:geometry:strip direction="next"
    private static void _define_db_setup() {
        // @ocalque:geometry:indent:pushd value="{{ macro.self.ref.indent - self.indent }}"
        
        // In local dev, we write valid Java. AOT replaces it with template variables.
        // @ocalque:replace_line --- DB_URL = {{ args.url }};
        DB_URL = "jdbc:postgresql://prod-db:5432/siemens";
        
        // @ocalque:replace_line --- DB_USER = {{ args.user }};
        DB_USER = "siemens_prod";
        
        System.out.println("Production DB active: " + DB_URL);
        // @ocalque:geometry:indent:popd
        // @ocalque:geometry:strip direction="next"
    }
    // @ocalque:stdlib:macro:end<db_setup>

    public static void main(String[] args) {
        System.out.println("--- Starting Java Configuration Setup ---");

        // ==========================================
        // 2. ACTIVE SHADOW MOCK FOR DATABASE SETUP
        // ==========================================
        // @ocalque:stdlib:if<prod_db> --- "{{ env.NAME }}" == "production"
            // @ocalque:stdlib:macro:call<db_setup> url="jdbc:postgresql://prod-db:5432/siemens" user="siemens_prod"
        // @ocalque:stdlib:else<prod_db>
            System.out.println("Local dev: active database mock mode.");
        // @ocalque:stdlib:fi<prod_db>

        System.out.println("Database URL: " + DB_URL + " (User: " + DB_USER + ")");

        // ==========================================
        // 3. HOSTS REGISTRATION (AOT LOOP GENERATION)
        // ==========================================
        // @ocalque:stdlib:for<host_loop> item="host" in="env.HOSTS"
        //   @ocalque:stdlib:if<prod_host> --- "{{ env.NAME }}" == "production"
        //     @ocalque:geometry:strip direction="next"
        if (false) {
            // @ocalque:replace_line --- registerHost({{ args.host.name }}, {{ args.host.ip }});
            registerHost("mock_host", "127.0.0.1");
            // @ocalque:geometry:strip direction="next"
        }
        //   @ocalque:stdlib:else<prod_host>
            System.out.println("Local dev: mock host registration (virtual dev network)");
        //   @ocalque:stdlib:fi<prod_host>
        // @ocalque:stdlib:end<host_loop>

        System.out.println("--- Java Setup Complete ---");
    }
}
