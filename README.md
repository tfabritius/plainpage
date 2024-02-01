# ![Logo](frontend/public/favicon.svg) PlainPage

PlainPage, a self-hosted app that embodies simplicity, ease of use, and privacy. PlainPage offers a plain and simplistic user experience while providing essential wiki-like functionality. It adheres to the 😘 KISS principle, focusing on delivering a straightforward and efficient user experience without unnecessary complexities.

Powered by Go, Vue/Nuxt. Made with 💖

## Key Features

- **Markdown Syntax** ✨ Create and edit your content using Markdown syntax, a lightweight and intuitive markup language that makes writing and formatting a breeze.

- **Plain File Storage** 🤩 PlainPage utilizes a plain file storage system, your content is stored in a simple and accessible format.

- **Snappy User Interface** 🚀 PlainPage offers a snappy and responsive user interface, allowing you to focus on your content and enhancing your productivity.

- **Privacy-First** 🛑 PlainPage does not track your usage, collect statistics, or employ any analytics tools. Your data remains private and under your control.

- **Access Rights** 🔐 PlainPage provides robust access control with powerful yet user-friendly ACLs. You can easily manage and assign permissions to your content, ensuring that the right people have access to the right information.

### More Features

- Localization: English, German, Spanish
- Supports light/dark mode
- Embedded full text search

### Known Limitations

- Not optimized for access by machines, e.g., search engines
- Concurrent/conflicting edits are not prevented

## Installation and setup

### Get Started

To start using PlainPage, self-host the app in your preferred environment. The process is quick and hassle-free, allowing you to jump right into creating, organizing, and managing your content.

Trying it out is as simple as it can be:

1. Download the [latest release](https://github.com/tfabritius/plainpage/releases/latest),
2. Run the PlainPage executable, and
3. Browse to <http://localhost:8080/>.

Or just run the docker image 🔋:

```bash
docker run --rm -p 8080:8080 ghcr.io/tfabritius/plainpage
```

⚠️ Be aware, this setup is not production ready.

### Get ready for productive use

#### Persist data

Make sure to persist your data by mounting a volume or local folder to the container's `/data` directory:
```bash
docker run -p 8080:8080 -v /path/on/host:/data ghcr.io/tfabritius/plainpage
```

_As usual: Backup your data!_

#### Reverse proxy

When running in production you typically want to run a reverse proxy in front of PlainPage that provides encryption via TLS/SSL. Popular choices are: Caddy, Traeffic, Nginx, Apache, ...

### Configuration

PlainPage executable has very few configration settings with sensible default values that can be configured by environment variables (or a `.env` file):

```ini
# Directory to store data, defaults to `data` directory next to executable
DATA_DIR=./data

# Port to listen on, defaults to 8080
PORT=8080
```

All other settings can be done via the UI or by editing the `config.yml` file in the data directory:

- `appTitle` is the title of the app. Use it to customize your installation of PlainPage.

- `acl` contains the access rules that apply independent from access rules of individual pages and folders.

- `jwtSecret` is a the secret that is used to sign and verify JWT tokens. It is generated automatically. Keep it safe! For security reasons it's neither exposed nor can be changed via UI.

- `setupMode` enables anonymous user registration and user will be granted admin rights. Will be disabled after first registration.

## Pages and folders

PlainPage stores your information on pages. Pages are organized into folders, that can be nested.

Pages are written in Markdown syntax.

## Access rights

Access rights can be modified by administrators.

### Access rights for pages and folders

PlainPage's permission system is powerful and yet easy to use. If you don't adjust it, all registered users will be able to access and edit all content.

Permissions can be defined for both pages and folders. By default, permissions for pages and folders are inherited from parent folder.

- A page `/folder/page` will inherit its access rigts from the folder `/folder`, unless explicit permissions are defined for this page.
- The folder `/folder` will itself inherit its permissions from the parent folder `/`, unless permissions are defined for `/folder`.
- The home folder `/` cannot inherit permissions.

Permissions can be granted to individual users, but also to all registered users and users, that are not logged in ("anonymous" users).

Permissions are controlled finegrained for different operations:

- *Read* allows users to access pages and folder
- *Write* allows users to change pages and folders (e.g. create new pages or folders within a folder)
- *Delete* allows users to delete pages and folders completely (e.g. deleting a page instead of just editing it)

That allows you to restrict certain content of PlainPage to some users and/or expose certain content.

### Access rights beyond pages and folders

Besides pages and folders PlainPage allows you to grant additional permissions:

- The *Register* privilege allows to create new users in PlainPage. This way you can control whether new users (so far anonymous users without an account) can create their own account and whether existing users can create additional user accounts. By default, only administrator can register new users.

- The *Admin* privilege grants a user special rights, e.g. to change the permissions. Users with this privilege are automatically granted all other possible permissions on all content.

## Keyboard shortcuts

- `e`: Edit page
- `Ctrl+s`: Save page (when editing)
- `Esc`: Cancel edit (when editing), close full screen mode
- `Ctrl+Backsapce`: Delete current page/folder

## Contribute

🤩 You're welcome to contribute to PlainPage.

### Running for development

During development, there are two processes running:

1. Nuxt frontend

    ```bash
    cd frontend
    pnpm run dev
    ```

2. Go backend

    ```bash
    cd backend
    go run .
    ```

Browse to [http://localhost:3000](http://localhost:3000). Nuxt will proxy requests to the backend.

### Build executable

In production the static frontend files are served together with the backend:

1. Generate static frontend files

    ```bash
    cd frontend
    pnpm run generate
    ```

2. Serve frontend from Go backend process

    ```bash
    cd backend
    go generate ./...
    go build .
    ```

Browse to [http://localhost:8080](http://localhost:8080).
