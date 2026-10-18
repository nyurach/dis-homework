# Open Classroom

A local education administration app built with Go, PostgreSQL, HTML, CSS, and JavaScript. Use the browser to view, add, edit, and delete students, teachers, courses, videos, and grades.

<video src="hw2.mov" width="320" height="240" controls></video>

## Requirements

- Go 1.25 or newer (see `go.mod`).
- PostgreSQL 18 installed and running separately from the app.

## Setup and run

Run commands from the project root, the folder containing `go.mod`.

Edit `internal/database/config.go`:

```go
const (
    host     = "127.0.0.1"
    port     = "5432"
    username = "your_postgres_username"
    password = "your_postgres_password"
)
```

Leave the password empty only if your local PostgreSQL authentication allows it.

Start the app:

```sh
go mod download
go run .
```

Open **http://127.0.0.1:8000** in your browser.

## Connect using pgAdmin 4

Right-click **Servers → Register → Server**. Give the connection a name, then enter:

| Connection setting | Value |
|---|---|
| Host | `127.0.0.1` |
| Port | `5432` (or your configured port) |
| Maintenance database | `postgres` |
| Username/password | The PostgreSQL login used in `config.go` |

After the app has initialized the databases, expand **Databases → students → Schemas → public → Tables**. Right-click a table and choose **View/Edit Data → All Rows**. Refresh the object tree if the databases are not visible yet.
