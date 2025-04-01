package render

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

func RenderTemplate(c echo.Context, tmpl *template.Template, name string, data interface{}) error {
	return tmpl.ExecuteTemplate(c.Response().Writer, name, data)
}
