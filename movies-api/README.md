# Movies API

A REST API for managing a movie database — genres, movies, and actors, with many-to-many relationships between them.

Built at Hive by Amankeldi Kurban (Movie) and Polina Filippova (Genre, Actor).

## Setup

```
go mod tidy
go run main.go
```

The server starts on `http://localhost:8081` and seeds the database with sample data on first run.

## All endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/movies` | Create a movie |
| `GET` | `/api/movies` | List all movies |
| `GET` | `/api/movies?genre={genreId}` | Movies filtered by genre |
| `GET` | `/api/movies?year={releaseYear}` | Movies filtered by release year |
| `GET` | `/api/movies?actor={actorId}` | Movies the specified actor has starred in |
| `GET` | `/api/movies?page={page}&size={size}` | Paginated list of movies |
| `GET` | `/api/movies/search?title={title}` | Search movies by title (case-insensitive, partial match) |
| `GET` | `/api/movies/{id}` | Get a movie by id |
| `GET` | `/api/movies/{id}/actors` | Get all actors starring in a movie |
| `PATCH` | `/api/movies/{id}` | Partially update a movie |
| `DELETE` | `/api/movies/{id}` | Delete a movie |
| `POST` | `/api/genres` | Create a genre |
| `GET` | `/api/genres` | List all genres |
| `GET` | `/api/genres?movies=true` | List genres including their movies |
| `GET` | `/api/genres/{id}` | Get a genre by id (includes its movies) |
| `PATCH` | `/api/genres/{id}` | Partially update a genre |
| `DELETE` | `/api/genres/{id}` | Delete a genre |
| `DELETE` | `/api/genres/deleteconnection/{id}` | Unlink specific movies from a genre |
| `POST` | `/api/actors` | Create an actor |
| `GET` | `/api/actors` | List all actors |
| `GET` | `/api/actors?movies=true` | List actors including their filmography |
| `GET` | `/api/actors?page={page}&size={size}` | Paginated list of actors |
| `GET` | `/api/actors?name={name}` | Search actors by name (case-insensitive, partial match) |
| `GET` | `/api/actors/{id}` | Get an actor by id (includes filmography) |
| `PATCH` | `/api/actors/{id}` | Partially update an actor |
| `DELETE` | `/api/actors/{id}` | Delete an actor |
| `DELETE` | `/api/actors/deleteconnection/{id}` | Unlink specific movies from an actor |

Movie query filters (`genre`, `year`, `actor`) and pagination (`page`/`size`) can't currently be combined in one request — use one at a time. When no filters are given, all movies are returned. Same for Actor: `movies`, `page`/`size`, and `name` can't be combined — `name` takes priority if present.

## Postman links:
- Movies: https://documenter.getpostman.com/view/56847520/2sBY4WnGRn#intro
- Actors and genres: https://documenter.getpostman.com/view/56933209/2sBY4WnGWB

## Genre endpoints

### `POST /api/genres` — create a genre

**Body**
- Required: `name` (string)

```json
{
  "name": "Neo-noir"
}
```

### `GET /api/genres` — list genres

**Query params** (optional)
- `movies=true` — include each genre's associated movies

### `GET /api/genres/{id}` — get one genre

Always includes associated movies.

### `PATCH /api/genres/{id}` — update a genre

**Body** (all optional — omit any field to leave it unchanged)
- `name`
- `movieIds` (`[]int`) — sets the genre's movies to exactly this list (adds missing ones, removes any not listed)

```json
{
  "name": "Neo-noir Thriller"
}
```

### `DELETE /api/genres/{id}` — delete a genre

**Query params**
- `force=true` (optional) — deletes the genre even if it's linked to movies (removes the links too). Without it, deleting a genre with movies fails with `400`

### `DELETE /api/genres/deleteconnection/{id}` — unlink specific movies

**Body**
- Required: `movieIds` (`[]int`) — only these links are removed; the genre and any other links stay intact

```json
{
  "movieIds": [1, 5]
}
```

## Actor endpoints

### `POST /api/actors` — create an actor

**Body**
- Required: `name` (string), `birthdate` (string, `YYYY-MM-DD`)
- Optional: `movieIds` (`[]int`) — link to existing movies right away

```json
{
  "name": "Tom Hardy",
  "birthdate": "1977-09-15",
  "movieIds": [1, 5]
}
```

Returns the created actor, including its `id` and `version`.

### `GET /api/actors` — list actors

**Query params** (all optional)
- `movies=true` — include each actor's filmography
- `page`, `size` — paginate; **must be provided together**. Response is `{"actors": [...], "page": ..., "size": ..., "total": ...}`
- `name={text}` — search by name instead of listing everyone (partial, case-insensitive). Always includes filmography

### `GET /api/actors/{id}` — get one actor

Always includes filmography.

### `PATCH /api/actors/{id}` — update an actor

**Body**
- **Required: `version` (int)** — the actor's current version (get it from a prior `GET`); the request fails without it
- Optional: `name`, `birthdate` — omit to leave unchanged
- Optional: `movieIds` (`[]int`) — sets the actor's movies to exactly this list (adds missing ones, removes any not listed)

```json
{
  "version": 1,
  "name": "Tom Hardy Jr."
}
```

If `version` doesn't match the actor's current version in the database (someone else updated it in the meantime), the request fails with `409 Conflict`. Re-fetch the actor and try again with the new version.

### `DELETE /api/actors/{id}` — delete an actor

**Query params**
- `force=true` (optional) — deletes the actor even if they're linked to movies (removes the links too). Without it, deleting an actor who has movies fails with `400`

### `DELETE /api/actors/deleteconnection/{id}` — unlink specific movies

**Body**
- Required: `movieIds` (`[]int`) — only these links are removed; the actor and any other links stay intact

```json
{
  "movieIds": [1, 5]
}
```

## What to expect

- `200 OK` — successful read
- `201 Created` — successful create
- `204 No Content` — successful delete
- `400 Bad Request` — invalid input (missing required field, bad id, bad date, etc.) — the response body explains what's wrong
- `404 Not Found` — the id doesn't exist
- `409 Conflict` — an actor was updated by someone else since you last fetched it (see `PATCH /api/actors/{id}` above)