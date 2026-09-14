package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light dark">
  <title>RL Simulation Control API · Swagger</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.18.2/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *::before, *::after { box-sizing: inherit; }
    body { margin: 0; background: #f7f8fb; }
    .swagger-ui .topbar { background: #101828; }
    .swagger-ui .topbar-wrapper img { display: none; }
    .swagger-ui .topbar-wrapper::before {
      content: "P7 · RL Simulation Control API";
      color: #fff;
      font: 600 15px/1.4 system-ui, sans-serif;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.18.2/swagger-ui-bundle.js"></script>
  <script>
    window.addEventListener("load", function () {
      window.ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: "#swagger-ui",
        deepLinking: true,
        displayRequestDuration: true,
        persistAuthorization: false,
        supportedSubmitMethods: ["get", "post"],
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    });
  </script>
</body>
</html>`

func registerSwaggerUI(router *gin.Engine) {
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/docs/")
	})
	router.GET("/docs/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline' https://cdn.jsdelivr.net; script-src 'unsafe-inline' https://cdn.jsdelivr.net; img-src data: https:; connect-src 'self'; font-src https://cdn.jsdelivr.net")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})
}
