# ratoneando-go

<img alt="Ratoneando Logo" src="https://ratoneando.ar/assets/ratoneando.705c1d01.png" width="100" height="100" />

API de comparación de precios de supermercados utilizada en [https://ratoneando.ar](https://ratoneando.ar) 🐀

## Motivación

Este proyecto fué creado originalmente en Node.js y Fastify, pero decidí reescribirlo en Go para usarlo como una oportunidad de aprendizaje. Por este motivo, [la calidad del código puede no ser la mejor](https://i.kym-cdn.com/photos/images/newsfeed/001/330/845/f24.jpg).

A pesar de eso, el rendimiento de la API es significativamente mejor que la versión en Node.js.

![screenshot](https://utfs.io/f/1b808c20-f370-483f-880f-0459977352ca-1nq2cb.png)
🟡 Node
🔴 Go

## Pre-requisitos

- [Go](https://golang.org/dl/)
- [Redis](https://redis.io/download)
- [Air](https://github.com/air-verse/air) (Opcional)

## Instalación

```bash
git clone git@github.com:matiasbontempo/ratoneando-go.git
cd ratoneando-go
go mod download
```

Si querés usar Air para una mejor experiencia de desarrollo:

```bash
go install github.com/air-verse/air@latest
air init
```

## Configuración

```bash
cp .env.example .env
```

## Ejecución

```bash
go run main.go
```

Si estás usando Air:

```bash
air
```

## Estrucutra del proyecto

- **main.go**: Punto de entrada de la aplicación
- **/config**: Expone las variables de entorno
- **/controllers**: Controladores de la aplicación
- **/cores**: Núcleos de scraping reutilizables
- **/middlewares**: Middlewares para GIN
- **/products**: Modelos y utilidades de productos
- **/routes**: Rutas de la API
- **/scrapers**: Scrapers de los distintos supermercados
- **/units**: Utilidades para el procesamiento de unidades
- **/utils**: Utilidades generales

## Tests

```bash
go test ./...
```

## Scraping de tiendas VTEX

Carrefour, Dia, Disco, Farmacity, Jumbo, Mas Online y Vea son tiendas VTEX y se
consultan con la API pública del catálogo:

```
https://<tienda>/api/catalog_system/pub/products/search/?ft=<query>
```

Esta API no necesita hash, token ni sesión, así que no hay nada que actualizar
cuando la tienda actualiza sus apps. La versión anterior usaba la GraphQL de
`productSuggestions`, que requería un `sha256Hash` que expiraba cada algunos
meses y dejaba a todas las tiendas VTEX sin resultados.

## Contribuir

Si te interesa dar una mano, consultá la [Guía de Contribución](CONTRIBUTING.md) y el [Código de Conducta](CODE_OF_CONDUCT.md).

## Licencia

Este proyecto está licenciado bajo la Licencia MIT. Revisá la [licencia](LICENSE) para más información.
