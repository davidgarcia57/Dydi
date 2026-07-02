# Arquitectura de Dydi

Este documento detalla la estructura y el flujo de comunicación de los servicios en Dydi.

## Contexto del Sistema

Dydi es una plataforma de accountability social. Se basa en una arquitectura de microservicios escritos en Go, que exponen una API REST y un túnel WebSocket para clientes web (Vue 3) y móviles (React Native / Expo). El estado y la autenticación son gestionados en la nube a través de Supabase.

## Componentes y Diagrama

```mermaid
flowchart TD
    %% Clients
    Web[Frontend Web <br> Vue 3 + Vite]
    Mobile[App Móvil <br> Expo + React Native]
    
    %% API Gateway
    Gateway[API Gateway <br> Go 1.24 + chi]
    
    %% Microservices
    Groups[Groups Service <br> Grupos y Miembros]
    Habits[Habits Service <br> Hábitos y Check-ins]
    Realtime[Realtime Service <br> WebSockets Hub]
    
    %% Data / Auth
    Supabase[(Supabase <br> DB + Auth)]
    
    %% Connections
    Web -->|HTTPS| Gateway
    Web -->|WSS| Gateway
    Mobile -->|HTTPS| Gateway
    
    Gateway -->|JWT Valido + Internal Token| Groups
    Gateway -->|JWT Valido + Internal Token| Habits
    Gateway -->|Upgrade WS + Internal Token| Realtime
    
    Groups -->|Read/Write| Supabase
    Habits -->|Read/Write| Supabase
    Realtime -->|Eventos WS| Web
    
    %% Internal comms
    Groups -.->|Eventos (HTTP)| Realtime
    Habits -.->|Eventos (HTTP)| Realtime
```

## Responsabilidades de los Servicios

- **API Gateway**: Actúa como el único punto de entrada público. Valida los JWT provenientes de Supabase y enruta las solicitudes (REST o WebSocket) hacia los microservicios internos correspondientes. Inyecta un `INTERNAL_TOKEN` en cada petición proxieda para que los servicios confíen en el origen de la misma.
- **Groups Service**: Administra la creación de "squads", las invitaciones, y la membresía.
- **Habits Service**: Maneja la creación de propuestas de hábitos, los votos diarios, los check-ins y la ruleta semanal de penitencias.
- **Realtime Service**: Un hub de WebSocket para transmitir eventos en tiempo real a los clientes conectados de un mismo grupo. Los otros servicios le notifican vía HTTP interno y este hace *broadcast* a los websockets.

## Flujos Principales

### Flujo de Autenticación
1. El cliente (Web/Móvil) inicia sesión usando el SDK de Supabase Auth.
2. Supabase devuelve un token JWT.
3. El cliente adjunta este JWT en el header `Authorization` de todas sus llamadas HTTP hacia el API Gateway.
4. El API Gateway valida el JWT (usando la clave pública JWKS de Supabase).
5. Si es válido, el Gateway extrae el ID de usuario (`X-User-ID`) y agrega el secreto interno (`INTERNAL_TOKEN`), enviando ambos hacia los microservicios, los cuales ya no validan el JWT sino el token interno compartido.

### Flujo Representativo (Realizar un Check-in)
1. El usuario envía un `POST /habits/checkin` con su JWT.
2. API Gateway valida el JWT y enruta a `Habits Service`.
3. `Habits Service` verifica que el token interno coincida.
4. `Habits Service` registra el check-in en Supabase.
5. `Habits Service` hace una petición asíncrona interna a `Realtime Service`.
6. `Realtime Service` distribuye el evento "new_checkin" a todos los clientes (miembros del grupo) conectados a ese WebSocket.
7. Los UI de los demás miembros se actualizan al instante.

## Trade-off: ¿Por qué microservicios?

Dydi fue estructurado en microservicios como **ejercicio académico** para explorar y demostrar patrones distribuidos (API Gateways, comunicación sincrónica/asincrónica interna, validación descentralizada pero confiable de identidad). 
* **Ventaja:** Permite escalar componentes críticos (por ejemplo, el nodo de WebSocket) independientemente del backend de datos, y facilita el estudio de métricas separadas en Go.
* **Desventaja:** Añade complejidad de despliegue y orquestación para un volumen de datos que un monolito manejaría sin problemas, requiriendo validación cruzada y gestión del `INTERNAL_TOKEN`.
