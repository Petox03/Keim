# --- Configuración de salida ---
Output demo.gif
Output demo.mp4

# --- Configuración visual y del shell ---
Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 850
Set Padding 20
Set Framerate 60
Set PlaybackSpeed 1.0
Set Theme "Dark+"

# --- Ocultar la preparación del entorno ---
Hide
Type "export PATH=$PATH:./bin"
Enter
Type "clear"
Enter
Show

# --- 1. Scaffolding del proyecto ---
Sleep 1s
Type "keim init --detect=manual test-app"
Sleep 500ms
Enter

# Esperar el Banner de Keim
Sleep 2s

# Responder al prompt interactivo de Devcontainer
Type "y"
Sleep 500ms
Enter

# Responder al prompt interactivo de la Versión de Go
Sleep 1.5s
Type "1.26.5"
Sleep 500ms
Enter

# Pausa para mostrar el reporte generado
Sleep 3.5s

# --- 2. Demostración funcional (Docker) ---
Type "cd test-app"
Enter
Sleep 500ms

# Levantar el entorno
Type "docker compose up -d"
Enter

# Dar tiempo a Docker Compose para levantar el servicio
Sleep 13s

# Ejecutar el 'Hello World' dentro del contenedor
Type "docker compose exec app go run ."
Enter

# Pausa final mostrando la salida de Go
Sleep 4s