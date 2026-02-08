# About

This application is a web app analytics platform akin to Google Analytics or Umami. It provides a web application with a database backend (SQLite or Postgres) to capture pageviews, browser stats, location stats.

# Architecture

This application is written using:

- Go 1.26
- Vanilla JavaScript and CSS 
- Uses github.com/adampresley/mux as a thin wrapper around The Go standard library mux and routing 
- Uses github.com/adampresley/rendering as a wrapper around Go's html/template standard library 
- Uses github.com/adampresley/httphelpers to provide generic methods to retrieve request data and respond
- Uses GORM (https://gorm.io) for database ORM to support SQLite and Postgres

# Rules

- After creating or modifying a Go file, always run `goimports -w <changed file or package>`, where **changed file or package** is the file or package for files where you made a change
- In Go functions, **always** declare variables at the top of the function in an `var ()` block. If the function needs an error variable, the very first variable decalred in the block must be `err error`. 
- For HTTP handlers, follow the pattern outlined below:

```go 
func (h ExampleHandler) ExamplePage(w http.ResponseWriter, r *http.Request) {
	pageName := "pages/example"

	viewData := viewdata.Example{
		BaseViewModel: rendering.BaseViewModel{
			IsHtmx: requests.IsHtmx(r),
		},
		ExampleData: "",
	}

	h.renderer.Render(pageName, viewData, w)
}
```

- Pages in this application have an `.html` extension, live in `app/pages`, and follow this format:

```html 
{{template "layouts/main-layout" .}}
{{define "title"}}Example{{end}}
{{define "content"}}

<h2>Example Page</h2>

<p>Hello world!</p>

{{end}}
```

- This application uses HTMX (https://htmx.org/). Here are the rules:
   - Links should not use HTMX. They should load a new page (route)
   - Interactions on a given page, like a form post, searching a table, etc... should use HTMX 
   - You should use `requests.IsHtmx(r)` to determine if a given request came from HTMX in a handler. Each view data will compose `rendering.BaseViewModel`, which has a field named `IsHtmx`, which can be used in HTML templates. This will allow you to determine how to render.
- Database models live in `internal/models`
- Services that perform logic, interact with a database or api, live in `internal/services`
- Handlers live in `internal/handlers`
- View data structures live in `internal/viewdata`
- All file names must be lower-hypen case.
- This application uses GORM (> 1.30) for ORM and database interactions. This version supports generics. You must use the generics interface. See https://gorm.io/docs/ for more information. 


