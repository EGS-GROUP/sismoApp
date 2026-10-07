# 🚀 Manual de Implementación SismoMonitor (Ubuntu Server)

Este documento detalla el procedimiento de despliegue ("paneo quirúrgico") para poner en producción **SismoMonitor** en un servidor Linux Ubuntu limpio, sin romper el entorno de desarrollo actual. 

La arquitectura en producción consiste en:
- **Backend (Go):** Compilado y ejecutado como proceso PM2 en el puerto `8082`.
- **Frontend (SvelteKit):** Construido y ejecutado como servicio systemd en el puerto `3001` (el puerto `3000` está reservado por Grafana en este servidor).
- **Nginx (Reverse Proxy):** Servirá como el punto de entrada (puerto `80`/`443`), redirigiendo `/api/` al backend y el resto al frontend.
- **Fuentes de datos:** USGS, EMSC, FUNVISIS, CSN, PRSN (Puerto Rico), UWI SRC (Caribe Oriental), IPGP (Antillas Francesas).

La arquitectura técnica completa, el flujo de datos y la metodología de desarrollo están documentados en [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

---

## 📦 1. Preparación del Servidor (Dependencias)

Conéctate a tu servidor Ubuntu por SSH y asegúrate de tener las herramientas necesarias:

```bash
# Actualizar repositorios
sudo apt update && sudo apt upgrade -y

# Instalar Nginx, Git, Node.js, NPM
sudo apt update && sudo apt install -y nginx git curl
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# PM2 se instala vía npx (no global) en este servidor
# Instalar Go (si no está instalado)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

*(Asegúrate de clonar o copiar tu repositorio `sismoApp` al servidor, por ejemplo en `/var/www/sismoApp`).*

---

## ⚙️ 2. Compilación del Backend (Go)

El backend debe compilarse en un ejecutable binario nativo de Linux y debe inicializar la base de datos de manera autónoma.

```bash
cd /var/www/sismoApp/backend

# Inicializar módulos de Go (si no existen)
go mod init sismoapp/backend
go mod tidy

# Compilar el binario optimizado para producción
go build -o sismoserver .
```

**Ejecutar el Backend con PM2:**
El backend en este servidor se gestiona con PM2 para aprovechar su reinicio automático y persistencia de procesos.

```bash
cd /var/www/sismoApp/backend

# Si PM2 no está en PATH, usar la ruta de npx
PM2=/root/.npm/_npx/5f7878ce38f1eb13/node_modules/pm2/bin/pm2
$PM2 start sismoserver --name sismoserver
$PM2 save
$PM2 startup
```

> **Nota:** El servicio systemd `sismobackend.service` existe pero está deshabilitado en este servidor para evitar conflicto con PM2.

---

## 🎨 3. Compilación del Frontend (SvelteKit)

El frontend está configurado con Vite y SvelteKit. Lo construiremos usando Node.

```bash
cd /var/www/sismoApp/frontend

# Instalar dependencias puras
npm install

# Construir para producción (genera la carpeta /build)
npm run build

# Lanzar como servicio systemd (no PM2)
# El puerto 3000 está ocupado por Grafana; el frontend usa 3001
sudo systemctl daemon-reload
sudo systemctl enable sismoapp
sudo systemctl start sismoapp
```

El servicio `sismoapp.service` ya está configurado con:

```ini
[Unit]
Description=SismoMonitor Frontend (SvelteKit)
After=network.target sismobackend.service

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/sismoApp/frontend
Environment=PORT=3001
Environment=HOST=127.0.0.1
Environment=ORIGIN=https://sismo.jagmedia.com.ve
ExecStart=/usr/bin/node build/index.js
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

---

## 🌐 4. Configuración Quirúrgica de Nginx

Nginx será el maestro de orquesta. Despachará la web por el puerto `80` (o `443` si usas SSL), servirá cualquier caché estático rápidamente y enrutará la API y los WebSockets hacia Go de manera transparente.

Crea el archivo de configuración:

```bash
sudo nano /etc/nginx/sites-available/sismomonitor
```

Pega el siguiente bloque de configuración:

```nginx
server {
    listen 80;
    server_name tu_dominio.com o_tu_ip;

    # 1. API y SSE (Backend Go en el puerto 8082)
    location /api/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";
        proxy_read_timeout 86400;
        proxy_buffering off;
        proxy_cache off;
    }

    # 2. WebSockets (legacy)
    location /socket.io/ {
        return 410;
    }

    # 3. Frontend de SvelteKit (puerto 3001)
    location / {
        proxy_pass http://127.0.0.1:3001;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
        proxy_buffering off;
    }
}
```

Activa el sitio, verifica la sintaxis y reinicia Nginx:

```bash
sudo ln -s /etc/nginx/sites-available/sismomonitor /etc/nginx/sites-enabled/
# Eliminar la página por defecto de nginx
sudo rm /etc/nginx/sites-enabled/default

# Verificar que no hay errores de sintaxis
sudo nginx -t

# Reiniciar Nginx para aplicar el ruteo
sudo systemctl restart nginx
```

## ✅ 5. Verificación de Salud

Todo debería estar operando en armonía. Para comprobar la salud de los servicios, ejecuta:

1. **Estado del Backend:** `pm2 status` (debe salir `sismoserver` en verde).
2. **Estado del Frontend:** `sudo systemctl status sismoapp` (Verificar que diga `active (running)`).
3. **Estado del Proxy:** Entra a `https://sismo.jagmedia.com.ve/` en tu navegador.
4. **Logs del Backend:** `pm2 logs sismoserver`.
5. **Logs del Frontend:** `sudo journalctl -u sismoapp -f`.

¡Listo! Producción blindada. Nginx se encargará de interceptar y balancear todo de forma segura.

---

## � 6. Estructura del Proyecto

```text
sismoApp/
├── AGENTS.md              # Contexto rápido para agentes / arquitectura actual
├── README.md              # Manual de implementación
├── FUTURE_IMPROVEMENTS.md # Mejoras futuras identificadas
├── RESUMEN_CAMBIOS.md     # Resumen de cambios recientes
├── docker-compose.yml     # Desarrollo local con Docker
├── nginx.conf             # Configuración de proxy (alternativa a sites-available)
├── index.html             # Página de mantenimiento
├── backend/
│   ├── main.go            # Servidor HTTP, SSE y endpoints /api/*
│   ├── api.go             # Handlers REST
│   ├── scraper.go         # Colectores de fuentes sísmicas
│   ├── db.go              # SQLite: esquema, queries, migración
│   ├── push.go            # Suscripciones Web Push VAPID
│   ├── middleware.go      # Rate limiting, CORS, logging
│   ├── *_test.go          # Tests unitarios
│   └── sismos.db          # Base de datos SQLite
└── frontend/
    ├── package.json       # Scripts y dependencias
    ├── svelte.config.js   # Configuración de SvelteKit
    ├── src/
    │   ├── app.css        # Variables CSS, temas dark/light
    │   ├── app.html       # Plantilla HTML
    │   ├── routes/+page.svelte       # Dashboard principal
    │   ├── lib/
    │   │   ├── sismoStore.svelte.ts  # Estado global reactivo ($state)
    │   │   ├── components/           # Componentes Svelte
    │   │   └── api.ts                # Cliente de API
    │   └── (generated)               # Rutas y assets generados
    ├── static/            # Assets estáticos (plates.json, hls.min.js, sw.js)
    └── build/             # Output de producción (generado)
```

## 🌍 7. Fuentes de Datos Sísmicos

| Fuente | Región | Formato | Notas |
|--------|--------|---------|-------|
| USGS | Global | GeoJSON | Feed continuo, magnitud ≥ 2.5 |
| EMSC | Euro-Mediterráneo | GeoJSON | Deduplicación con USGS/CSN |
| FUNVISIS | Venezuela | JSON propio | Mapeo de campos no convencional |
| CSN | Chile / Sudamérica | GeoJSON | |
| PRSN | Puerto Rico / Caribe NE | RSS/XML | Codificación ISO-8859-1 |
| UWI SRC | Caribe Oriental | JSON Leaflet | `data=EQ`, bbox del mapa |
| IPGP | Antillas Francesas | FDSNWS text | Últimos 7 días, M≥2 |

## ⚡ 8. Comandos Útiles

```bash
# Reconstruir y reiniciar backend
cd /var/www/sismoApp/backend
go build -o sismoserver .
PM2=/root/.npm/_npx/5f7878ce38f1eb13/node_modules/pm2/bin/pm2
$PM2 restart sismoserver

# Reconstruir y reiniciar frontend
cd /var/www/sismoApp/frontend
npm run build
sudo systemctl restart sismoapp

# Ver logs
$PM2 logs sismoserver
sudo journalctl -u sismoapp -f
sudo nginx -t && sudo systemctl restart nginx
```

## 🩺 9. Troubleshooting

| Síntoma | Causa probable | Solución |
|---------|---------------|----------|
| Backend no arranca | Puerto 8082 ocupado | `sudo pkill -9 sismoserver` o revisar PM2 |
| Frontend no arranca | Grafana usa el puerto 3000 | Asegurar `PORT=3001` en `sismoapp.service` |
| UWI devuelve 0 eventos | Endpoint incorrecto | Verificar `data=EQ` en `scraper.go` |
| PRSN con magnitud 0 | Eventos preliminares | Filtrar en frontend o ignorar |
| No se ven sismos caribeños | Ventana de 24 h | El feed normal muestra solo 24 h; la API sí guarda 7 días |

## �💾 10. Migración de Datos Históricos (SQLite)

Como SismoMonitor utiliza SQLite, toda la base de datos de producción es literalmente un solo archivo: `sismos.db`. Para migrar tu historial actual al nuevo servidor, el proceso es tan quirúrgico como copiar y pegar.

**Desde tu computadora local (Windows) o servidor antiguo:**
Transfiere el archivo `sismos.db` al nuevo servidor usando un cliente SFTP (como FileZilla, WinSCP) o mediante consola con SCP.

```bash
# Ejemplo de envío por consola (SCP)
scp G:\sismoApp\backend\sismos.db usuario@tu_ip_servidor:/var/www/sismoApp/backend/sismos.db
```

**En el nuevo servidor (Ubuntu):**
Una vez transferido, asegúrate de que el archivo tenga los permisos correctos para que el demonio del backend pueda leerlo y escribir nuevos sismos:

```bash
cd /var/www/sismoApp/backend

# Ajustar permisos
sudo chmod 664 sismos.db
sudo chown root:root sismos.db

# Reiniciar el backend para cargar el historial inyectado
PM2=/root/.npm/_npx/5f7878ce38f1eb13/node_modules/pm2/bin/pm2
$PM2 restart sismoserver
```

¡Eso es todo! Al recargar el sistema, todo tu registro histórico intacto estará vivo en producción.
