package serverdebug

import (
	"html/template"

	"github.com/labstack/echo/v4"

	"github.com/pttrulez/ninja-chat/internal/logger"
)

type page struct {
	Path        string
	Description string
}

type indexPage struct {
	pages []page
}

func newIndexPage() *indexPage {
	return &indexPage{}
}

func (i *indexPage) addPage(path string, description string) {
	i.pages = append(i.pages, page{Path: path, Description: description})
}

func (i indexPage) handler(eCtx echo.Context) error {
	return template.Must(template.New("index").Parse(`<html>
	<title>Chat Service Debug</title>
<body>
	<h2>Chat Service Debug</h2>
	<ul>
		{{ range .Pages }}
			<li>
				<a href={{ .Path }}>{{ .Path }}</a> <span>{{ .Description }}</span>
			</li>
		{{ end }}
	</ul>

	<h2>Log Level</h2>
	<form onSubmit="putLogLevel()">
		<select id="log-level-select" name="level">
			 <option value="debug" {{ if (eq .LogLevel "DEBUG") }} selected="selected" {{ end }} >DEBUG</option>
			 <option value="info" {{ if (eq .LogLevel "INFO") }} selected="selected" {{ end }} >INFO</option>
			 <option value="warn" {{ if (eq .LogLevel "WARN") }} selected="selected" {{ end }} >WARN</option>
			 <option value="error" {{ if (eq .LogLevel "ERROR") }} selected="selected" {{ end }}>ERROR</option>
		</select>
		<input type="submit" value="Change"></input>
	</form>
	
	
		<script>
    function putLogLevel() {
      const req = new XMLHttpRequest();
      req.open('PUT', '/log/level', false);
      req.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
      req.onload = function() { window.location.reload(); };
      req.send('level='+document.getElementById('log-level-select').value);
    };
  </script>
	
</body>
</html>
`)).Execute(eCtx.Response(), struct {
		Pages    []page
		LogLevel string
	}{
		Pages:    i.pages,
		LogLevel: logger.GetAtomicLevel().Level().CapitalString(),
	})
}
