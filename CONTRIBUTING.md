# Guía de Contribución para Dydi

¡Gracias por tu interés en Dydi! Este documento describe el flujo esperado para colaborar en el proyecto manteniendo la calidad técnica y consistencia del código.

## 1. Flujo de Ramas (Git Flow)

* Mantenemos una rama principal (`main`) que siempre debe ser desplegable.
* Todo trabajo nuevo debe realizarse en ramas derivadas de `main`.

### Nombrado de Ramas
Usa prefijos semánticos cortos:
* `feature/nombre-de-la-funcionalidad`
* `fix/bug-a-corregir`
* `chore/tarea-mantenimiento`
* `docs/cambio-en-documentacion`

```bash
git checkout main
git pull origin main
git checkout -b feature/nueva-ruleta
```

## 2. Convención de Commits

Sigue el formato de [Conventional Commits](https://www.conventionalcommits.org/). Esto facilita la lectura del historial y la automatización.

* `feat: agrega votación de propuestas`
* `fix: corrige error de WebSocket desconectado`
* `docs: actualiza diagrama de arquitectura`
* `refactor: extrae middleware de autenticación`

Mantenlos atómicos: un commit por cambio lógico coherente.

## 3. Uso de `verify.sh` (Obligatorio)

Todo el proceso de validación (Linter, Build, Tests y Race Detector) sucede localmente usando contenedores Docker. Así garantizamos que el código que compila en tu máquina compilará en CI.

Antes de hacer push de tus cambios o abrir un Pull Request, **debes correr el script de verificación y asegurar que todo esté en verde.**

```bash
# Validar TODO (Back, Front, Mobile)
./verify.sh

# Validar solo el código Go
./verify.sh go

# Validar solo el frontend
./verify.sh frontend
```
*Si estás en Windows, corre esto dentro de tu entorno WSL.*

## 4. Regla Estricta: Nunca subir `.env`

Las variables de entorno (`.env`, `.env.local`, etc.) contienen credenciales privadas de base de datos (`DATABASE_URL`, claves de JWT y Supabase). 
* **JAMÁS** deben ser añadidas al control de versiones.
* Asegúrate de que tu rama no incluya archivos de configuración privada.
* Si agregas una nueva variable, documenta su necesidad en `.env.example`.

## 5. Pull Requests

1. Sube tu rama: `git push origin tu-rama`.
2. Abre un PR contra `main`.
3. El GitHub Actions ejecutará `verify.sh`.
4. El PR solo podrá ser *merged* si el check **CI Success** se encuentra en verde y al menos otro mantenedor ha aprobado los cambios.

¡Disfruta contribuyendo a la cultura del accountability!
